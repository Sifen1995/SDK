package infrastructure

import (
	"context"
	"strings"
	"time"

	deliverydomain "skykin-platform/internal/delivery/domain"
	"skykin-platform/internal/delivery/infrastructure/persistence"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) deliverydomain.DeliveryRepository {
	return &DeliveryRepository{db: db}
}

func (r *DeliveryRepository) WasDelivered(ctx context.Context, userID string, campaignID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&persistence.DeliveryJobRow{}).
		Where("user_id = ? AND campaign_id = ?", userID, campaignID.String()).
		Count(&n).Error
	return n > 0, err
}

func (r *DeliveryRepository) CountToday(ctx context.Context, userID string, campaignID uuid.UUID) (int, error) {
	start := time.Now().UTC().Truncate(24 * time.Hour)
	var n int64
	err := r.db.WithContext(ctx).Model(&persistence.DeliveryJobRow{}).
		Where("user_id = ? AND campaign_id = ? AND created_at >= ?", userID, campaignID.String(), start).
		Count(&n).Error
	return int(n), err
}

// RecordJob inserts a delivery_jobs row (ignores duplicate user+campaign pairs).
func (r *DeliveryRepository) RecordJob(ctx context.Context, userID, campaignID string) error {
	if r == nil || r.db == nil || userID == "" || campaignID == "" {
		return nil
	}
	row := &persistence.DeliveryJobRow{
		UserID:     userID,
		CampaignID: campaignID,
		CreatedAt:  time.Now().UTC(),
	}
	err := r.db.WithContext(ctx).Create(row).Error
	if err != nil && isDuplicateKey(err) {
		return nil
	}
	return err
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "23505")
}
