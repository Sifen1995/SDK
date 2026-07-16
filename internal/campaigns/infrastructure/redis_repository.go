package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	"skykin-platform/internal/platform/redis"
)

const eligibleCampaignsTTL = 5 * time.Minute

// EligibleCampaignReader is the Postgres (or other) source behind the Redis cache.
type EligibleCampaignReader interface {
	ListActiveByIntent(ctx context.Context, intentName string) ([]campaigndomain.Campaign, error)
}

// RedisCampaignRepository handles Redis-backed campaign delivery helpers
// (frequency capping + eligible-campaign cache) via the platform Redis wrapper.
type RedisCampaignRepository struct {
	rdb *redis.RedisClient
}

func NewRedisCampaignRepository(rdb *redis.RedisClient) *RedisCampaignRepository {
	return &RedisCampaignRepository{rdb: rdb}
}

// ============================================================================
// JOB 2 — Frequency capping
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
// JOB 1 — Eligible campaigns cache (Redis + Postgres fallback)
// ============================================================================

// CachedCampaignRepository wraps Postgres campaign reads with a Redis cache.
type CachedCampaignRepository struct {
	postgres EligibleCampaignReader
	rdb      *redis.RedisClient
}

func NewCachedCampaignRepository(postgres EligibleCampaignReader, rdb *redis.RedisClient) *CachedCampaignRepository {
	return &CachedCampaignRepository{
		postgres: postgres,
		rdb:      rdb,
	}
}

// GetEligibleCampaigns returns active campaigns for an intent (cache-aside).
func (r *CachedCampaignRepository) GetEligibleCampaigns(
	ctx context.Context,
	intentName string,
) ([]campaigndomain.Campaign, error) {
	if r == nil || r.postgres == nil {
		return nil, fmt.Errorf("cached campaign repository is not configured")
	}

	key := "eligible_campaigns:" + intentName

	if r.rdb != nil {
		raw, err := r.rdb.Get(ctx, key)
		if err == nil {
			var campaigns []campaigndomain.Campaign
			if err := json.Unmarshal([]byte(raw), &campaigns); err == nil {
				return campaigns, nil
			}
		} else if err != redis.ErrNil {
			// Unexpected Redis error — fall through to Postgres.
		}
	}

	campaigns, err := r.postgres.ListActiveByIntent(ctx, intentName)
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
