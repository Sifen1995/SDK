package infrastructure

import (
	"context"
	"time"

	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/model"

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
func (r *PurchaseRepository) GetValidForCampaign(ctx context.Context, campaignID string, now time.Time) (*model.SegmentPurchase, error) {
	var purchase model.SegmentPurchase
	err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND valid_from <= ? AND valid_until >= ?", campaignID, now, now).
		First(&purchase).Error
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}
