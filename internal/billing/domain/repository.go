package domain

import (
	"context"
	"time"
)

// SubscriptionContext is returned after a successful subscription gate check.
type SubscriptionContext struct {
	SubscriptionID string
	Plan           SubscriptionPlan
}

// SubscriptionRepository loads advertiser plans and manages subscriptions.
type SubscriptionRepository interface {
	GetActiveByAdvertiser(ctx context.Context, advertiserID string) (*AdvertiserSubscription, error)
	GetPlanByID(ctx context.Context, planID string) (*SubscriptionPlan, error)
	FindPlanByID(ctx context.Context, planID string) (*SubscriptionPlan, error)
	GetPlanByName(ctx context.Context, name string) (*SubscriptionPlan, error)
	FindPlanByName(ctx context.Context, name string) (*SubscriptionPlan, error)
	ListActivePlans(ctx context.Context) ([]SubscriptionPlan, error)
	ListAllPlans(ctx context.Context) ([]SubscriptionPlan, error)
	CreatePlan(ctx context.Context, plan *SubscriptionPlan) error
	UpdatePlan(ctx context.Context, plan *SubscriptionPlan) error
	CreateSubscription(ctx context.Context, sub *AdvertiserSubscription) error
}

// BillingRateRepository manages per-plan usage rates.
type BillingRateRepository interface {
	ListByPlanID(ctx context.Context, planID string) ([]BillingRate, error)
	GetByID(ctx context.Context, id string) (*BillingRate, error)
	UpdateRate(ctx context.Context, id string, rateETB float64, isActive bool) (*BillingRate, error)
	CreateBatch(ctx context.Context, rates []BillingRate) error
}

// BillingEventRepository persists calculated billing events.
type BillingEventRepository interface {
	CreateBatch(ctx context.Context, events []BillingEvent) error
}

// ChannelRepository loads delivery channels for entitlement checks.
type ChannelRepository interface {
	GetByID(ctx context.Context, id string) (*Channel, error)
	ListActive(ctx context.Context) ([]Channel, error)
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
