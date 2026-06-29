package validation

import (
	"errors"
	"strings"
)

// CreatePlanInput is the validated shape for a new subscription plan.
type CreatePlanInput struct {
	Name                string
	MonthlyFeeETB       float64
	MaxActiveCampaigns  int
	MaxDailyBudgetETB   float64
	IncludedImpressions int
	CPCDiscountPct      float64
}

// ValidateCreatePlan checks plan fields before persistence.
func ValidateCreatePlan(in CreatePlanInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("name is required")
	}
	if in.MonthlyFeeETB < 0 {
		return errors.New("monthly_fee_etb must be >= 0")
	}
	if in.MaxActiveCampaigns < 1 {
		return errors.New("max_active_campaigns must be at least 1")
	}
	if in.MaxDailyBudgetETB <= 0 {
		return errors.New("max_daily_budget_etb must be > 0")
	}
	if in.IncludedImpressions < 0 {
		return errors.New("included_impressions must be >= 0")
	}
	if in.CPCDiscountPct < 0 || in.CPCDiscountPct > 100 {
		return errors.New("cpc_discount_pct must be between 0 and 100")
	}
	return nil
}

// CreateSegmentInput is the validated shape for a new audience segment.
type CreateSegmentInput struct {
	Name             string
	TopIntentSignals []string
	ApproximateSize  int
	EstimatedCPM     float64
}

// ValidateCreateSegment checks segment catalog fields before persistence.
func ValidateCreateSegment(in CreateSegmentInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("name is required")
	}
	if len(in.TopIntentSignals) == 0 {
		return errors.New("top_intent_signals must include at least one intent")
	}
	for _, sig := range in.TopIntentSignals {
		if strings.TrimSpace(sig) == "" {
			return errors.New("top_intent_signals cannot contain empty values")
		}
	}
	if in.ApproximateSize < 0 {
		return errors.New("approximate_size must be >= 0")
	}
	if in.EstimatedCPM <= 0 {
		return errors.New("estimated_cpm must be > 0")
	}
	return nil
}
