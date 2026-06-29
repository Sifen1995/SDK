package validation

import (
	"errors"
	"strings"
)

// PlanFieldsInput is the shared validated shape for plan create/update.
type PlanFieldsInput struct {
	Name                string
	MonthlyFeeETB       float64
	MaxActiveCampaigns  int
	MaxDailyBudgetETB   float64
	IncludedImpressions int
	CPCDiscountPct      float64
}

// ValidatePlanFields checks mutable subscription plan fields.
func ValidatePlanFields(in PlanFieldsInput) error {
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

// ValidatePlanID ensures a plan id is present for lookups and updates.
func ValidatePlanID(planID string) error {
	if strings.TrimSpace(planID) == "" {
		return errors.New("plan id is required")
	}
	return nil
}
