package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	billingdomain "skykin-platform/internal/billing/domain"
)

// SubscriptionEnforcer validates that an advertiser's plan allows campaign creation.
type SubscriptionEnforcer struct {
	subs     billingdomain.SubscriptionRepository
	channels billingdomain.ChannelRepository
	campaign billingdomain.CampaignQuotaReader
}

func NewSubscriptionEnforcer(
	subs billingdomain.SubscriptionRepository,
	channels billingdomain.ChannelRepository,
	campaign billingdomain.CampaignQuotaReader,
) *SubscriptionEnforcer {
	return &SubscriptionEnforcer{subs: subs, channels: channels, campaign: campaign}
}

// AssertCanCreateCampaign checks subscription status, plan limits, channel tier, and Audiencemart access.
func (e *SubscriptionEnforcer) AssertCanCreateCampaign(
	ctx context.Context,
	advertiserID, channelID string,
	dailyBudgetETB float64,
	wantsSegment bool,
) (*billingdomain.SubscriptionContext, error) {
	sub, err := e.subs.GetActiveByAdvertiser(ctx, advertiserID)
	if err != nil {
		return nil, errors.New("no active subscription; contact support or re-register")
	}
	now := time.Now().UTC()
	if sub.Status != "active" {
		return nil, fmt.Errorf("subscription is %s; cannot create campaigns", sub.Status)
	}
	if now.Before(sub.CurrentPeriodStart) || now.After(sub.CurrentPeriodEnd) {
		return nil, errors.New("subscription billing period has expired")
	}

	plan := sub.Plan
	activeCount, err := e.campaign.CountActiveByAdvertiser(ctx, advertiserID)
	if err != nil {
		return nil, err
	}
	if activeCount >= plan.MaxActiveCampaigns {
		return nil, fmt.Errorf("plan %q allows at most %d active campaigns", plan.Name, plan.MaxActiveCampaigns)
	}
	if dailyBudgetETB > plan.MaxDailyBudgetETB {
		return nil, fmt.Errorf("daily_budget_cap exceeds plan limit of %.2f ETB", plan.MaxDailyBudgetETB)
	}

	ch, err := e.channels.GetByID(ctx, channelID)
	if err != nil {
		return nil, errors.New("channel not found")
	}
	if !ch.IsActive {
		return nil, errors.New("channel is not available")
	}
	if ch.IsPremium && !plan.SMSPlusEnabled {
		return nil, fmt.Errorf("plan %q does not include premium channel %q", plan.Name, ch.Code)
	}
	if wantsSegment && !plan.AudiencemartEnabled {
		return nil, fmt.Errorf("plan %q does not include Audiencemart segments", plan.Name)
	}

	return &billingdomain.SubscriptionContext{
		SubscriptionID: sub.ID,
		Plan:           plan,
	}, nil
}

// StarterSubscriptionService assigns the default Starter plan to new advertisers.
type StarterSubscriptionService struct {
	subs billingdomain.SubscriptionRepository
}

func NewStarterSubscriptionService(subs billingdomain.SubscriptionRepository) *StarterSubscriptionService {
	return &StarterSubscriptionService{subs: subs}
}

// AssignStarter creates a Starter subscription if the advertiser has none yet.
func (s *StarterSubscriptionService) AssignStarter(ctx context.Context, advertiserID string) error {
	if advertiserID == "" {
		return nil
	}
	return s.subs.EnsureStarterForAdvertiser(ctx, advertiserID)
}
