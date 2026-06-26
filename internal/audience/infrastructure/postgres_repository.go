package infrastructure

import (
	"context"
	"time"

	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/model"

	"gorm.io/gorm"
)

type SegmentRepository struct {
	db *gorm.DB
}

func NewSegmentRepository(db *gorm.DB) *SegmentRepository {
	return &SegmentRepository{db: db}
}

var _ audiencedomain.SegmentRepository = (*SegmentRepository)(nil)

func (r *SegmentRepository) GetByID(ctx context.Context, id string) (*model.AudienceSegment, error) {
	var seg model.AudienceSegment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&seg).Error; err != nil {
		return nil, err
	}
	return &seg, nil
}

func (r *SegmentRepository) GetByName(ctx context.Context, name string) (*model.AudienceSegment, error) {
	var seg model.AudienceSegment
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&seg).Error; err != nil {
		return nil, err
	}
	return &seg, nil
}

func (r *SegmentRepository) Create(ctx context.Context, seg *model.AudienceSegment) error {
	return r.db.WithContext(ctx).Create(seg).Error
}

// ListAvailableNow returns catalog segments that are active and within their availability window.
func (r *SegmentRepository) ListAvailableNow(ctx context.Context, now time.Time) ([]model.AudienceSegment, error) {
	var list []model.AudienceSegment
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND available_from <= ?", true, now).
		Where("available_until IS NULL OR available_until >= ?", now).
		Order("name asc").
		Find(&list).Error
	return list, err
}
