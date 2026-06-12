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
