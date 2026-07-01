package infrastructure

import (
	"context"

	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/billing/infrastructure/persistence"

	"gorm.io/gorm"
)

type BillingRateRepository struct {
	db *gorm.DB
}

func NewBillingRateRepository(db *gorm.DB) *BillingRateRepository {
	return &BillingRateRepository{db: db}
}

var _ billingdomain.BillingRateRepository = (*BillingRateRepository)(nil)

func (r *BillingRateRepository) ListByPlanID(ctx context.Context, planID string) ([]billingdomain.BillingRate, error) {
	var rows []persistence.BillingRateRow
	err := r.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Order("event_type asc, model asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDomainBillingRates(rows), nil
}

func (r *BillingRateRepository) GetByID(ctx context.Context, id string) (*billingdomain.BillingRate, error) {
	var row persistence.BillingRateRow
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *BillingRateRepository) UpdateRate(ctx context.Context, id string, rateETB float64, isActive bool) (*billingdomain.BillingRate, error) {
	var row persistence.BillingRateRow
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	row.RateETB = rateETB
	row.IsActive = isActive
	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *BillingRateRepository) CreateBatch(ctx context.Context, rates []billingdomain.BillingRate) error {
	if len(rates) == 0 {
		return nil
	}
	rows := make([]persistence.BillingRateRow, len(rates))
	for i := range rates {
		rows[i] = *persistence.BillingRateRowFromDomain(&rates[i])
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func toDomainBillingRates(rows []persistence.BillingRateRow) []billingdomain.BillingRate {
	out := make([]billingdomain.BillingRate, len(rows))
	for i := range rows {
		out[i] = *rows[i].ToDomain()
	}
	return out
}
