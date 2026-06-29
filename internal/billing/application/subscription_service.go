package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/billing/model"

	"gorm.io/gorm"
)

// PlanDTO is the portal-facing view of a subscription plan.
type PlanDTO struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	MonthlyFeeETB       float64 `json:"monthly_fee_etb"`
	MaxActiveCampaigns  int     `json:"max_active_campaigns"`
	MaxDailyBudgetETB   float64 `json:"max_daily_budget_etb"`
	IncludedImpressions int     `json:"included_impressions"`
	SMSPlusEnabled      bool    `json:"sms_plus_enabled"`
	AudiencemartEnabled bool    `json:"audiencemart_enabled"`
	CPCDiscountPct      float64 `json:"cpc_discount_pct"`
}

// SubscriptionDTO is the portal-facing view of an advertiser subscription.
type SubscriptionDTO struct {
	ID                 string    `json:"id"`
	Plan               PlanDTO   `json:"plan"`
	Status             string    `json:"status"`
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	ImpressionsUsed    int       `json:"impressions_used"`
}

// ChannelDTO is the portal-facing view of a delivery channel.
type ChannelDTO struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPremium   bool   `json:"is_premium"`
}

// SubscriptionService handles plan browsing and advertiser subscriptions.
type SubscriptionService struct {
	subs     billingdomain.SubscriptionRepository
	channels billingdomain.ChannelRepository
}

func NewSubscriptionService(subs billingdomain.SubscriptionRepository, channels billingdomain.ChannelRepository) *SubscriptionService {
	return &SubscriptionService{subs: subs, channels: channels}
}

func (s *SubscriptionService) ListPlans(ctx context.Context) ([]PlanDTO, error) {
	plans, err := s.subs.ListActivePlans(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PlanDTO, 0, len(plans))
	for i := range plans {
		out = append(out, toPlanDTO(&plans[i]))
	}
	return out, nil
}

func (s *SubscriptionService) ListChannels(ctx context.Context) ([]ChannelDTO, error) {
	channels, err := s.channels.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ChannelDTO, 0, len(channels))
	for i := range channels {
		out = append(out, ChannelDTO{
			ID:          channels[i].ID,
			Code:        channels[i].Code,
			Name:        channels[i].Name,
			Description: channels[i].Description,
			IsPremium:   channels[i].IsPremium,
		})
	}
	return out, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, advertiserID string) (*SubscriptionDTO, error) {
	if advertiserID == "" {
		return nil, errors.New("no advertiser account linked to this user")
	}
	sub, err := s.subs.GetActiveByAdvertiser(ctx, advertiserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	dto := toSubscriptionDTO(sub)
	return &dto, nil
}

func (s *SubscriptionService) Subscribe(ctx context.Context, advertiserID, planID string) (*SubscriptionDTO, error) {
	if advertiserID == "" {
		return nil, errors.New("no advertiser account linked to this user")
	}
	if planID == "" {
		return nil, errors.New("plan_id is required")
	}

	if existing, err := s.subs.GetActiveByAdvertiser(ctx, advertiserID); err == nil && existing != nil {
		return nil, fmt.Errorf("already subscribed to plan %q; contact support to change plans", existing.Plan.Name)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	plan, err := s.subs.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan not found")
		}
		return nil, err
	}

	start, end := billingdomain.SubscriptionPeriod(time.Now())
	sub := &model.AdvertiserSubscription{
		AdvertiserID:       advertiserID,
		PlanID:             plan.ID,
		Status:             "active",
		CurrentPeriodStart: start,
		CurrentPeriodEnd:   end,
	}
	if err := s.subs.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	sub.Plan = *plan
	dto := toSubscriptionDTO(sub)
	return &dto, nil
}

func (s *SubscriptionService) GetPlanByID(ctx context.Context, planID string) (*PlanDTO, error) {
	plan, err := s.subs.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	dto := toPlanDTO(plan)
	return &dto, nil

}
func (s *SubscriptionService) UpdatePlan(ctx context.Context, palnID string) (*PlanDTO, error) {
	plan, err := s.subs.UpdatePlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	dto := toPlanDTO(plan)
	return &dto, nil
}

func toPlanDTO(p *model.SubscriptionPlan) PlanDTO {
	return PlanDTO{
		ID:                  p.ID,
		Name:                p.Name,
		MonthlyFeeETB:       p.MonthlyFeeETB,
		MaxActiveCampaigns:  p.MaxActiveCampaigns,
		MaxDailyBudgetETB:   p.MaxDailyBudgetETB,
		IncludedImpressions: p.IncludedImpressions,
		SMSPlusEnabled:      p.SMSPlusEnabled,
		AudiencemartEnabled: p.AudiencemartEnabled,
		CPCDiscountPct:      p.CPCDiscountPct,
	}
}

func toSubscriptionDTO(sub *model.AdvertiserSubscription) SubscriptionDTO {
	return SubscriptionDTO{
		ID:                 sub.ID,
		Plan:               toPlanDTO(&sub.Plan),
		Status:             sub.Status,
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		ImpressionsUsed:    sub.ImpressionsUsed,
	}
}
