package infrastructure

import (
	"context"

	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/billing/model"

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
func (r *SubscriptionRepository) GetActiveByAdvertiser(ctx context.Context, advertiserID string) (*model.AdvertiserSubscription, error) {
	var sub model.AdvertiserSubscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("advertiser_id = ? AND status = ?", advertiserID, "active").
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepository) GetPlanByName(ctx context.Context, name string) (*model.SubscriptionPlan, error) {
	var plan model.SubscriptionPlan
	if err := r.db.WithContext(ctx).Where("name = ? AND is_active = ?", name, true).First(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *SubscriptionRepository) FindPlanByName(ctx context.Context, name string) (*model.SubscriptionPlan, error) {
	var plan model.SubscriptionPlan
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *SubscriptionRepository) CreatePlan(ctx context.Context, plan *model.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *SubscriptionRepository) GetPlanByID(ctx context.Context, planID string) (*model.SubscriptionPlan, error) {
	var plan model.SubscriptionPlan
	if err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", planID, true).First(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *SubscriptionRepository) ListActivePlans(ctx context.Context) ([]model.SubscriptionPlan, error) {
	var plans []model.SubscriptionPlan
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("monthly_fee_etb ASC").
		Find(&plans).Error
	return plans, err
}

func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, sub *model.AdvertiserSubscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}
