package infrastructure

import (
	"context"

	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/billing/model"

	"gorm.io/gorm"
)

type BillingRateRepository struct {
	db *gorm.DB
}

func NewBillingRateRepository(db *gorm.DB) *BillingRateRepository {
	return &BillingRateRepository{db: db}
}

var _ billingdomain.BillingRateRepository = (*BillingRateRepository)(nil)

func (r *BillingRateRepository) ListByPlanID(ctx context.Context, planID string) ([]model.BillingRate, error) {
	var rates []model.BillingRate
	err := r.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Order("event_type asc, model asc").
		Find(&rates).Error
	return rates, err
}

func (r *BillingRateRepository) GetByID(ctx context.Context, id string) (*model.BillingRate, error) {
	var rate model.BillingRate
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rate).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *BillingRateRepository) UpdateRate(ctx context.Context, id string, rateETB float64, isActive bool) (*model.BillingRate, error) {
	var rate model.BillingRate
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rate).Error; err != nil {
		return nil, err
	}
	rate.RateETB = rateETB
	rate.IsActive = isActive
	if err := r.db.WithContext(ctx).Save(&rate).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *BillingRateRepository) CreateBatch(ctx context.Context, rates []model.BillingRate) error {
	if len(rates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&rates).Error
}
