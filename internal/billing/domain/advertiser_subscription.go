package domain

import "time"

// AdvertiserSubscription links an advertiser to a subscription plan.
type AdvertiserSubscription struct {
	ID                 string
	AdvertiserID       string
	PlanID             string
	Plan               SubscriptionPlan
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	ImpressionsUsed    int
	CancelledAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
