package domain

import "time"

// BillingEvent records a billable campaign interaction for usage metering.
type BillingEvent struct {
	ID               string
	AdvertiserID     string
	CampaignID       string
	SubscriptionID   string
	EventType        string
	BillingModel     string
	RateApplied      float64
	TransactionValue float64
	ChargeETB        float64
	IsBilled         bool
	OccurredAt       time.Time
	CreatedAt        time.Time
}
