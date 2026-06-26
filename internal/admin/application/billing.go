package application

import (
	"context"
	"errors"

	billingdomain "skykin-platform/internal/billing/domain"
	billingmodel "skykin-platform/internal/billing/model"

	"gorm.io/gorm"
)

// UpdateBillingRateCmd updates a single billing rate row.
type UpdateBillingRateCmd struct {
	RateID   string
	RateETB  float64
	IsActive bool
}

// BillingAdminService handles operator billing rate management.
type BillingAdminService struct {
	planRepo billingdomain.SubscriptionRepository
	rateRepo billingdomain.BillingRateRepository
}

func NewBillingAdminService(
	planRepo billingdomain.SubscriptionRepository,
	rateRepo billingdomain.BillingRateRepository,
) *BillingAdminService {
	return &BillingAdminService{planRepo: planRepo, rateRepo: rateRepo}
}

// ListBillingRates returns all rates configured for a plan.
func (s *BillingAdminService) ListBillingRates(ctx context.Context, planID string) ([]billingmodel.BillingRate, error) {
	if _, err := s.planRepo.GetPlanByID(ctx, planID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan not found")
		}
		return nil, err
	}
	return s.rateRepo.ListByPlanID(ctx, planID)
}

// UpdateBillingRate updates rate_etb and is_active for a billing rate row.
func (s *BillingAdminService) UpdateBillingRate(ctx context.Context, cmd UpdateBillingRateCmd) (*billingmodel.BillingRate, error) {
	if cmd.RateID == "" {
		return nil, errors.New("rate id is required")
	}
	if cmd.RateETB < 0 {
		return nil, errors.New("rate_etb must be >= 0")
	}
	rate, err := s.rateRepo.GetByID(ctx, cmd.RateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("billing rate not found")
		}
		return nil, err
	}
	return s.rateRepo.UpdateRate(ctx, rate.ID, cmd.RateETB, cmd.IsActive)
}
