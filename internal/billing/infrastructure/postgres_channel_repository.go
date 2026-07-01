package infrastructure

import (
	"context"

	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/billing/infrastructure/persistence"

	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

var _ billingdomain.ChannelRepository = (*ChannelRepository)(nil)

func (r *ChannelRepository) GetByID(ctx context.Context, id string) (*billingdomain.Channel, error) {
	var row persistence.ChannelRow
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *ChannelRepository) ListActive(ctx context.Context) ([]billingdomain.Channel, error) {
	var rows []persistence.ChannelRow
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("is_premium ASC, name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]billingdomain.Channel, len(rows))
	for i := range rows {
		out[i] = *rows[i].ToDomain()
	}
	return out, nil
}
