package infrastructure

import (
	"context"
	"time"

	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/infrastructure/persistence"

	"gorm.io/gorm"
)

type PurchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

var _ audiencedomain.PurchaseRepository = (*PurchaseRepository)(nil)

// GetValidForCampaign returns an active segment purchase for the campaign, if any.
func (r *PurchaseRepository) GetValidForCampaign(ctx context.Context, campaignID string, now time.Time) (*audiencedomain.SegmentPurchase, error) {
	var row persistence.SegmentPurchaseRow
	err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND valid_from <= ? AND valid_until >= ?", campaignID, now, now).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *PurchaseRepository) CreatePurchase(ctx context.Context, purchase *audiencedomain.SegmentPurchase) error {
	return r.CreatePurchaseTx(ctx, r.db, purchase)
}

func (r *PurchaseRepository) CreatePurchaseTx(
	ctx context.Context,
	tx any,
	purchase *audiencedomain.SegmentPurchase,
) error {
	db, ok := tx.(*gorm.DB)
	if !ok || db == nil {
		return gorm.ErrInvalidTransaction
	}
	row := persistence.SegmentPurchaseRowFromDomain(purchase)
	if err := db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	purchase.ID = row.ID
	purchase.CreatedAt = row.CreatedAt
	return nil
}
