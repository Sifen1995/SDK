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

// SubscriptionRepository loads advertiser plans and manages subscriptions.
type SubscriptionRepository interface {
	GetActiveByAdvertiser(ctx context.Context, advertiserID string) (*model.AdvertiserSubscription, error)
	GetPlanByID(ctx context.Context, planID string) (*model.SubscriptionPlan, error)
	GetPlanByName(ctx context.Context, name string) (*model.SubscriptionPlan, error)
	ListActivePlans(ctx context.Context) ([]model.SubscriptionPlan, error)
	CreateSubscription(ctx context.Context, sub *model.AdvertiserSubscription) error
}

// ChannelRepository loads delivery channels for entitlement checks.
type ChannelRepository interface {
	GetByID(ctx context.Context, id string) (*model.Channel, error)
	ListActive(ctx context.Context) ([]model.Channel, error)
}

// CampaignQuotaReader counts campaigns for plan limit enforcement.
type CampaignQuotaReader interface {
	CountActiveByAdvertiser(ctx context.Context, advertiserID string) (int, error)
}

// SubscriptionPeriod defines the first billing cycle length for a new subscription.
func SubscriptionPeriod(now time.Time) (start, end time.Time) {
	start = now.UTC()
	end = start.AddDate(0, 1, 0)
	return start, end
}
