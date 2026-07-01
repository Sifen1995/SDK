package infrastructure

import (
	"context"

	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/billing/infrastructure/persistence"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

var _ billingdomain.SubscriptionRepository = (*SubscriptionRepository)(nil)

// GetActiveByAdvertiser returns the advertiser's subscription with plan preloaded.
func (r *SubscriptionRepository) GetActiveByAdvertiser(ctx context.Context, advertiserID string) (*billingdomain.AdvertiserSubscription, error) {
	var row persistence.AdvertiserSubscriptionRow
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("advertiser_id = ? AND status = ?", advertiserID, "active").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *SubscriptionRepository) GetPlanByName(ctx context.Context, name string) (*billingdomain.SubscriptionPlan, error) {
	var row persistence.SubscriptionPlanRow
	if err := r.db.WithContext(ctx).Where("name = ? AND is_active = ?", name, true).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *SubscriptionRepository) FindPlanByName(ctx context.Context, name string) (*billingdomain.SubscriptionPlan, error) {
	var row persistence.SubscriptionPlanRow
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *SubscriptionRepository) CreatePlan(ctx context.Context, plan *billingdomain.SubscriptionPlan) error {
	row := persistence.SubscriptionPlanRowFromDomain(plan)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	plan.ID = row.ID
	plan.CreatedAt = row.CreatedAt
	return nil
}

func (r *SubscriptionRepository) GetPlanByID(ctx context.Context, planID string) (*billingdomain.SubscriptionPlan, error) {
	var row persistence.SubscriptionPlanRow
	if err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", planID, true).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *SubscriptionRepository) FindPlanByID(ctx context.Context, planID string) (*billingdomain.SubscriptionPlan, error) {
	var row persistence.SubscriptionPlanRow
	if err := r.db.WithContext(ctx).Where("id = ?", planID).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *SubscriptionRepository) UpdatePlan(ctx context.Context, plan *billingdomain.SubscriptionPlan) error {
	row := persistence.SubscriptionPlanRowFromDomain(plan)
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *SubscriptionRepository) ListActivePlans(ctx context.Context) ([]billingdomain.SubscriptionPlan, error) {
	var rows []persistence.SubscriptionPlanRow
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("monthly_fee_etb ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDomainPlans(rows), nil
}

func (r *SubscriptionRepository) ListAllPlans(ctx context.Context) ([]billingdomain.SubscriptionPlan, error) {
	var rows []persistence.SubscriptionPlanRow
	err := r.db.WithContext(ctx).
		Order("is_active DESC, monthly_fee_etb ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDomainPlans(rows), nil
}

func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, sub *billingdomain.AdvertiserSubscription) error {
	row := persistence.AdvertiserSubscriptionRowFromDomain(sub)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	sub.ID = row.ID
	sub.CreatedAt = row.CreatedAt
	sub.UpdatedAt = row.UpdatedAt
	return nil
}

func toDomainPlans(rows []persistence.SubscriptionPlanRow) []billingdomain.SubscriptionPlan {
	out := make([]billingdomain.SubscriptionPlan, len(rows))
	for i := range rows {
		out[i] = *rows[i].ToDomain()
	}
	return out
}
