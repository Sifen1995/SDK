package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	billingdomain "skykin-platform/internal/billing/domain"
	billingvalidation "skykin-platform/internal/billing/validation"
	"skykin-platform/internal/billing/model"

	"gorm.io/gorm"
)

// UpdatePlanCmd carries operator updates to a subscription plan.
type UpdatePlanCmd struct {
	PlanID              string
	Name                string
	MonthlyFeeETB       float64
	MaxActiveCampaigns  int
	MaxDailyBudgetETB   float64
	IncludedImpressions int
	SMSPlusEnabled      bool
	AudiencemartEnabled bool
	CPCDiscountPct      float64
	IsActive            bool
}

// PlanService handles subscription plan reads and updates.
type PlanService struct {
	plans billingdomain.SubscriptionRepository
}

func NewPlanService(plans billingdomain.SubscriptionRepository) *PlanService {
	return &PlanService{plans: plans}
}

// GetPlanByID returns any plan by id (including inactive) for operator admin.
func (s *PlanService) GetPlanByID(ctx context.Context, planID string) (*model.SubscriptionPlan, error) {
	if err := billingvalidation.ValidatePlanID(planID); err != nil {
		return nil, err
	}
	plan, err := s.plans.FindPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan not found")
		}
		return nil, err
	}
	return plan, nil
}

// ListAllPlans returns every subscription plan for operator admin (active and inactive).
func (s *PlanService) ListAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error) {
	return s.plans.ListAllPlans(ctx)
}

// SuspendPlan deactivates an active subscription plan so it is hidden from advertisers.
func (s *PlanService) SuspendPlan(ctx context.Context, planID string) (*model.SubscriptionPlan, error) {
	if err := billingvalidation.ValidatePlanID(planID); err != nil {
		return nil, err
	}
	plan, err := s.plans.FindPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan not found")
		}
		return nil, err
	}
	if !plan.IsActive {
		return nil, errors.New("plan is already suspended")
	}
	plan.IsActive = false
	if err := s.plans.UpdatePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("suspend plan: %w", err)
	}
	return plan, nil
}

// UpdatePlan updates mutable plan fields for operator admin.
func (s *PlanService) UpdatePlan(ctx context.Context, cmd UpdatePlanCmd) (*model.SubscriptionPlan, error) {
	if err := billingvalidation.ValidatePlanID(cmd.PlanID); err != nil {
		return nil, err
	}
	if err := billingvalidation.ValidatePlanFields(billingvalidation.PlanFieldsInput{
		Name:                cmd.Name,
		MonthlyFeeETB:       cmd.MonthlyFeeETB,
		MaxActiveCampaigns:  cmd.MaxActiveCampaigns,
		MaxDailyBudgetETB:   cmd.MaxDailyBudgetETB,
		IncludedImpressions: cmd.IncludedImpressions,
		CPCDiscountPct:      cmd.CPCDiscountPct,
	}); err != nil {
		return nil, err
	}

	plan, err := s.plans.FindPlanByID(ctx, cmd.PlanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan not found")
		}
		return nil, err
	}

	trimmedName := strings.TrimSpace(cmd.Name)
	if !strings.EqualFold(plan.Name, trimmedName) {
		if existing, err := s.plans.FindPlanByName(ctx, trimmedName); err == nil && existing.ID != plan.ID {
			return nil, errors.New("plan with this name already exists")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	plan.Name = trimmedName
	plan.MonthlyFeeETB = cmd.MonthlyFeeETB
	plan.MaxActiveCampaigns = cmd.MaxActiveCampaigns
	plan.MaxDailyBudgetETB = cmd.MaxDailyBudgetETB
	plan.IncludedImpressions = cmd.IncludedImpressions
	plan.SMSPlusEnabled = cmd.SMSPlusEnabled
	plan.AudiencemartEnabled = cmd.AudiencemartEnabled
	plan.CPCDiscountPct = cmd.CPCDiscountPct
	plan.IsActive = cmd.IsActive

	if err := s.plans.UpdatePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("update plan: %w", err)
	}
	return plan, nil
}
