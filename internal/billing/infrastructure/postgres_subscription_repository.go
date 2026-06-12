package infrastructure

import (
	"context"
	"errors"
	"time"

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

func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, sub *model.AdvertiserSubscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *SubscriptionRepository) CountAdvertisersWithoutSubscription(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM advertisers a
		WHERE NOT EXISTS (
			SELECT 1 FROM advertiser_subscriptions s WHERE s.advertiser_id = a.id
		)
	`).Scan(&n).Error
	return n, err
}

// EnsureStarterForAdvertiser assigns Starter when missing (idempotent).
func (r *SubscriptionRepository) EnsureStarterForAdvertiser(ctx context.Context, advertiserID string) error {
	if _, err := r.GetActiveByAdvertiser(ctx, advertiserID); err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	plan, err := r.GetPlanByName(ctx, "Starter")
	if err != nil {
		return err
	}
	start, end := billingdomain.StarterPeriod(time.Now())
	return r.CreateSubscription(ctx, &model.AdvertiserSubscription{
		AdvertiserID:       advertiserID,
		PlanID:             plan.ID,
		Status:             "active",
		CurrentPeriodStart: start,
		CurrentPeriodEnd:   end,
	})
}
