package infrastructure

import (
	"context"

	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/billing/model"

	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

var _ billingdomain.ChannelRepository = (*ChannelRepository)(nil)

func (r *ChannelRepository) GetByID(ctx context.Context, id string) (*model.Channel, error) {
	var ch model.Channel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&ch).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChannelRepository) ListActive(ctx context.Context) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("is_premium ASC, name ASC").
		Find(&channels).Error
	return channels, err
}
