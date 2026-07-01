package domain

import "time"

// SubscriptionPlan defines advertiser plan limits and entitlements.
type SubscriptionPlan struct {
	ID                  string
	Name                string
	MonthlyFeeETB       float64
	MaxActiveCampaigns  int
	MaxDailyBudgetETB   float64
	IncludedImpressions int
	SMSPlusEnabled      bool
	AudiencemartEnabled bool
	CPCDiscountPct      float64
	IsActive            bool
	CreatedAt           time.Time
}
