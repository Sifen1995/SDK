package application

import (
	"context"

	billingdomain "skykin-platform/internal/billing/domain"
)

// SeedDefaultPlanRates initializes default billing rates for a newly created plan.
func SeedDefaultPlanRates(ctx context.Context, rateRepo billingdomain.BillingRateRepository, planID string) error {
	defaults := []struct {
		eventType string
		model     string
		rate      float64
	}{
		{"impression", "CPM", 2.5},
		{"click", "CPC", 0.75},
		{"install", "CPI", 15.0},
		{"signup", "CPA", 25.0},
		{"purchase", "REV_SHARE", 5.0},
	}
	rates := make([]billingdomain.BillingRate, 0, len(defaults))
	for _, d := range defaults {
		rates = append(rates, billingdomain.BillingRate{
			PlanID:    planID,
			EventType: d.eventType,
			Model:     d.model,
			RateETB:   d.rate,
			IsActive:  true,
		})
	}
	return rateRepo.CreateBatch(ctx, rates)
}
