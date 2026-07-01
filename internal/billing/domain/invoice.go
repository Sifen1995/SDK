package domain

import "time"

// Invoice records a billing period charge for an advertiser subscription.
type Invoice struct {
	ID                 string
	AdvertiserID       string
	SubscriptionID     string
	PeriodStart        time.Time
	PeriodEnd          time.Time
	SubscriptionFeeETB float64
	UsageFeeETB        float64
	TotalETB           float64
	Status             string
	PaidAt             *time.Time
	PaymentRef         string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
