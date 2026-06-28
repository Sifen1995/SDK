package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	segmentdomain "skykin-platform/internal/audience/domain"
	segmentmodel "skykin-platform/internal/audience/model"
	billingdomain "skykin-platform/internal/billing/domain"
	billingmodel "skykin-platform/internal/billing/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CreatePlanCmd is the application-level command for creating subscription plans.
type CreatePlanCmd struct {
	Name                string
	MonthlyFeeETB       float64
	MaxActiveCampaigns  int
	MaxDailyBudgetETB   float64
	IncludedImpressions int
	SMSPlusEnabled      bool
	AudiencemartEnabled bool
	CPCDiscountPct      float64
}

// CreateSegmentCmd is the application-level command for creating audience segments.
type CreateSegmentCmd struct {
	Name             string
	Description      string
	TopIntentSignals []string
	ApproximateSize  int
	EstimatedCPM     float64
	AvailableFrom    *time.Time
	AvailableUntil   *time.Time
	IsActive         bool
}

// PlanAndSegmentService manages admin operations for plans and audience segments.
type PlanAndSegmentService struct {
	planRepo    billingdomain.SubscriptionRepository
	rateRepo    billingdomain.BillingRateRepository
	segmentRepo segmentdomain.SegmentRepository
}

func NewPlanAndSegmentService(
	planRepo billingdomain.SubscriptionRepository,
	rateRepo billingdomain.BillingRateRepository,
	segmentRepo segmentdomain.SegmentRepository,
) *PlanAndSegmentService {
	return &PlanAndSegmentService{planRepo: planRepo, rateRepo: rateRepo, segmentRepo: segmentRepo}
}

// CreatePlan creates a subscription plan and seeds default billing rates.
func (s *PlanAndSegmentService) CreatePlan(ctx context.Context, cmd CreatePlanCmd) (*billingmodel.SubscriptionPlan, error) {
	if err := validateCreatePlan(cmd); err != nil {
		return nil, err
	}
	if _, err := s.planRepo.FindPlanByName(ctx, cmd.Name); err == nil {
		return nil, errors.New("plan with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	plan := &billingmodel.SubscriptionPlan{
		Name:                strings.TrimSpace(cmd.Name),
		MonthlyFeeETB:       cmd.MonthlyFeeETB,
		MaxActiveCampaigns:  cmd.MaxActiveCampaigns,
		MaxDailyBudgetETB:   cmd.MaxDailyBudgetETB,
		IncludedImpressions: cmd.IncludedImpressions,
		SMSPlusEnabled:      cmd.SMSPlusEnabled,
		AudiencemartEnabled: cmd.AudiencemartEnabled,
		CPCDiscountPct:      cmd.CPCDiscountPct,
		IsActive:            true,
	}
	if err := s.planRepo.CreatePlan(ctx, plan); err != nil {
		return nil, err
	}
	if err := s.seedDefaultRates(ctx, plan.ID); err != nil {
		return nil, fmt.Errorf("plan created but billing rates failed: %w", err)
	}
	return plan, nil
}

// CreateSegment creates an audience segment in the catalog.
func (s *PlanAndSegmentService) CreateSegment(ctx context.Context, cmd CreateSegmentCmd) (*segmentmodel.AudienceSegment, error) {
	if err := validateCreateSegment(cmd); err != nil {
		return nil, err
	}
	if _, err := s.segmentRepo.GetByName(ctx, strings.TrimSpace(cmd.Name)); err == nil {
		return nil, errors.New("segment with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	signalsJSON, err := json.Marshal(cmd.TopIntentSignals)
	if err != nil {
		return nil, fmt.Errorf("invalid intent signals: %w", err)
	}

	availableFrom := time.Now().UTC()
	if cmd.AvailableFrom != nil {
		availableFrom = cmd.AvailableFrom.UTC()
	}

	seg := &segmentmodel.AudienceSegment{
		Name:             strings.TrimSpace(cmd.Name),
		Description:      strings.TrimSpace(cmd.Description),
		TopIntentSignals: datatypes.JSON(signalsJSON),
		ApproximateSize:  cmd.ApproximateSize,
		EstimatedCPM:     cmd.EstimatedCPM,
		AvailableFrom:    availableFrom,
		AvailableUntil:   cmd.AvailableUntil,
		IsActive:         cmd.IsActive,
	}
	if err := s.segmentRepo.Create(ctx, seg); err != nil {
		return nil, err
	}
	return seg, nil
}

// ListSegments returns all active catalog audience segments for operator admin.
func (s *PlanAndSegmentService) ListSegments(ctx context.Context) ([]segmentmodel.AudienceSegment, error) {
	return s.segmentRepo.ListAvailableNow(ctx, time.Now().UTC())
}

func validateCreatePlan(cmd CreatePlanCmd) error {
	if strings.TrimSpace(cmd.Name) == "" {
		return errors.New("name is required")
	}
	if cmd.MonthlyFeeETB < 0 {
		return errors.New("monthly_fee_etb must be >= 0")
	}
	if cmd.MaxActiveCampaigns < 1 {
		return errors.New("max_active_campaigns must be at least 1")
	}
	if cmd.MaxDailyBudgetETB <= 0 {
		return errors.New("max_daily_budget_etb must be > 0")
	}
	if cmd.IncludedImpressions < 0 {
		return errors.New("included_impressions must be >= 0")
	}
	if cmd.CPCDiscountPct < 0 || cmd.CPCDiscountPct > 100 {
		return errors.New("cpc_discount_pct must be between 0 and 100")
	}
	return nil
}

func validateCreateSegment(cmd CreateSegmentCmd) error {
	if strings.TrimSpace(cmd.Name) == "" {
		return errors.New("name is required")
	}
	if len(cmd.TopIntentSignals) == 0 {
		return errors.New("top_intent_signals must include at least one intent")
	}
	for _, sig := range cmd.TopIntentSignals {
		if strings.TrimSpace(sig) == "" {
			return errors.New("top_intent_signals cannot contain empty values")
		}
	}
	if cmd.ApproximateSize < 0 {
		return errors.New("approximate_size must be >= 0")
	}
	if cmd.EstimatedCPM <= 0 {
		return errors.New("estimated_cpm must be > 0")
	}
	return nil
}

func (s *PlanAndSegmentService) seedDefaultRates(ctx context.Context, planID string) error {
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
	rates := make([]billingmodel.BillingRate, 0, len(defaults))
	for _, d := range defaults {
		rates = append(rates, billingmodel.BillingRate{
			PlanID:    planID,
			EventType: d.eventType,
			Model:     d.model,
			RateETB:   d.rate,
			IsActive:  true,
		})
	}
	return s.rateRepo.CreateBatch(ctx, rates)
}
