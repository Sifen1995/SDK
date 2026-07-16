package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	"skykin-platform/internal/platform/redis"
)

const eligibleCampaignsTTL = 5 * time.Minute

// EligibleCampaignReader is the Postgres source behind the Redis eligible-campaign cache.
type EligibleCampaignReader interface {
	ListEligibleForDelivery(ctx context.Context, intentName, channelCode string) ([]campaigndomain.Campaign, error)
}

// RedisCampaignRepository handles Redis-backed campaign delivery helpers via the platform wrapper.
type RedisCampaignRepository struct {
	rdb *redis.RedisClient
}

func NewRedisCampaignRepository(rdb *redis.RedisClient) *RedisCampaignRepository {
	return &RedisCampaignRepository{rdb: rdb}
}

// ============================================================================
// Budget exhaustion flags (set by billing background jobs)
// ============================================================================

func budgetExhaustedKey(campaignID string) string {
	return "budget_exhausted:" + campaignID
}

// IsBudgetExhausted returns true when billing flagged the campaign as out of budget.
func (r *RedisCampaignRepository) IsBudgetExhausted(ctx context.Context, campaignID string) (bool, error) {
	if r == nil || r.rdb == nil {
		return false, nil
	}
	return r.rdb.Exists(ctx, budgetExhaustedKey(campaignID))
}

// SetBudgetExhausted marks a campaign as budget-exhausted until TTL expires.
func (r *RedisCampaignRepository) SetBudgetExhausted(ctx context.Context, campaignID string, ttl time.Duration) error {
	if r == nil || r.rdb == nil {
		return nil
	}
	return r.rdb.Set(ctx, budgetExhaustedKey(campaignID), "1", ttl)
}

// ============================================================================
// Frequency capping
// ============================================================================

func (r *RedisCampaignRepository) IncrementDeliveryCount(
	ctx context.Context,
	pseudonymousID string,
	campaignID string,
	ttl time.Duration,
) (int64, error) {
	if r == nil || r.rdb == nil {
		return 0, nil
	}
	key := fmt.Sprintf("freq:%s:%s", pseudonymousID, campaignID)
	count, err := r.rdb.Incr(ctx, key)
	if err != nil {
		return 0, err
	}
	if count == 1 && ttl > 0 {
		_ = r.rdb.Expire(ctx, key, ttl)
	}
	return count, nil
}

func (r *RedisCampaignRepository) IsFrequencyCapped(
	ctx context.Context,
	pseudonymousID string,
	campaignID string,
	limit int,
) (bool, error) {
	if r == nil || r.rdb == nil {
		return false, nil
	}
	key := fmt.Sprintf("freq:%s:%s", pseudonymousID, campaignID)
	val, err := r.rdb.Get(ctx, key)
	if err == redis.ErrNil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return false, fmt.Errorf("parse frequency count for %s: %w", key, err)
	}
	return n >= limit, nil
}

// ============================================================================
// Eligible campaigns cache + in-memory plan-tier ranker
// ============================================================================

// CachedCampaignRepository wraps Postgres reads with Redis cache-aside and delivery filters.
type CachedCampaignRepository struct {
	postgres EligibleCampaignReader
	redis    *RedisCampaignRepository
	rdb      *redis.RedisClient
}

func NewCachedCampaignRepository(
	postgres EligibleCampaignReader,
	redisRepo *RedisCampaignRepository,
	rdb *redis.RedisClient,
) *CachedCampaignRepository {
	return &CachedCampaignRepository{
		postgres: postgres,
		redis:    redisRepo,
		rdb:      rdb,
	}
}

func eligibleCampaignsCacheKey(intentName, channelCode string) string {
	if channelCode == "" {
		return "eligible_campaigns:" + intentName + ":all"
	}
	return "eligible_campaigns:" + intentName + ":" + channelCode
}

// GetEligibleCampaigns returns cached active campaigns for an intent/channel (no delivery filters).
func (r *CachedCampaignRepository) GetEligibleCampaigns(
	ctx context.Context,
	intentName, channelCode string,
) ([]campaigndomain.Campaign, error) {
	return r.loadEligibleCampaigns(ctx, intentName, channelCode)
}

// SelectBestCampaign applies budget + frequency filters and ranks by subscription plan tier in Go memory.
func (r *CachedCampaignRepository) SelectBestCampaign(
	ctx context.Context,
	intentName, channelCode, pseudonymousID string,
) (*campaigndomain.Campaign, error) {
	if r == nil || r.postgres == nil {
		return nil, fmt.Errorf("cached campaign repository is not configured")
	}

	campaigns, err := r.loadEligibleCampaigns(ctx, intentName, channelCode)
	if err != nil {
		return nil, err
	}

	filtered := make([]campaigndomain.Campaign, 0, len(campaigns))
	for _, campaign := range campaigns {
		if r.redis != nil {
			exhausted, err := r.redis.IsBudgetExhausted(ctx, campaign.ID)
			if err != nil {
				return nil, err
			}
			if exhausted {
				continue
			}

			capLimit := campaign.FrequencyCapPerDay
			if capLimit <= 0 {
				capLimit = 3
			}
			capped, err := r.redis.IsFrequencyCapped(ctx, pseudonymousID, campaign.ID, capLimit)
			if err != nil {
				return nil, err
			}
			if capped {
				continue
			}
		}
		filtered = append(filtered, campaign)
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no eligible campaign for intent %s", intentName)
	}

	rankCampaignsByPlanSubscription(filtered)

	selected := filtered[0]
	if r.redis != nil && pseudonymousID != "" {
		_, _ = r.redis.IncrementDeliveryCount(ctx, pseudonymousID, selected.ID, 24*time.Hour)
	}

	return &selected, nil
}

func (r *CachedCampaignRepository) loadEligibleCampaigns(
	ctx context.Context,
	intentName, channelCode string,
) ([]campaigndomain.Campaign, error) {
	key := eligibleCampaignsCacheKey(intentName, channelCode)

	if r.rdb != nil {
		raw, err := r.rdb.Get(ctx, key)
		if err == nil {
			var campaigns []campaigndomain.Campaign
			if err := json.Unmarshal([]byte(raw), &campaigns); err == nil {
				return campaigns, nil
			}
		} else if err != redis.ErrNil {
			// fall through to Postgres on unexpected Redis errors
		}
	}

	campaigns, err := r.postgres.ListEligibleForDelivery(ctx, intentName, channelCode)
	if err != nil {
		return nil, err
	}

	if r.rdb != nil {
		if payload, err := json.Marshal(campaigns); err == nil {
			_ = r.rdb.Set(ctx, key, string(payload), eligibleCampaignsTTL)
		}
	}

	return campaigns, nil
}

// rankCampaignsByPlanSubscription sorts campaigns by active plan monthly fee (tier) descending.
func rankCampaignsByPlanSubscription(campaigns []campaigndomain.Campaign) {
	slices.SortFunc(campaigns, func(a, b campaigndomain.Campaign) int {
		if a.PlanMonthlyFeeETB > b.PlanMonthlyFeeETB {
			return -1
		}
		if a.PlanMonthlyFeeETB < b.PlanMonthlyFeeETB {
			return 1
		}
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		}
		if a.CreatedAt.Before(b.CreatedAt) {
			return 1
		}
		return 0
	})
}
