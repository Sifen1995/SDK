package domain

import (
	"context"
	"time"

	"skykin-platform/internal/billing/model"
)

// SubscriptionContext is returned after a successful subscription gate check.
type SubscriptionContext struct {
	SubscriptionID string
	Plan           model.SubscriptionPlan
}

// SubscriptionRepository loads advertiser plans and creates starter subscriptions.
type SubscriptionRepository interface {
	GetActiveByAdvertiser(ctx context.Context, advertiserID string) (*model.AdvertiserSubscription, error)
	GetPlanByName(ctx context.Context, name string) (*model.SubscriptionPlan, error)
	CreateSubscription(ctx context.Context, sub *model.AdvertiserSubscription) error
	// EnsureStarterForAdvertiser assigns Starter when the advertiser has no subscription yet.
	EnsureStarterForAdvertiser(ctx context.Context, advertiserID string) error
}

// ChannelRepository loads delivery channels for entitlement checks.
type ChannelRepository interface {
	GetByID(ctx context.Context, id string) (*model.Channel, error)
}

// CampaignQuotaReader counts campaigns for plan limit enforcement.
type CampaignQuotaReader interface {
	CountActiveByAdvertiser(ctx context.Context, advertiserID string) (int, error)
}

// StarterPeriod defines the first billing cycle length for new advertisers.
func StarterPeriod(now time.Time) (start, end time.Time) {
	start = now.UTC()
	end = start.AddDate(0, 1, 0)
	return start, end
}
