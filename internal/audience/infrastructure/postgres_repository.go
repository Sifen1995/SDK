package infrastructure

import (
	"context"
	"time"

	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/infrastructure/persistence"

	"gorm.io/gorm"
)

type SegmentRepository struct {
	db *gorm.DB
}

func NewSegmentRepository(db *gorm.DB) *SegmentRepository {
	return &SegmentRepository{db: db}
}

var _ audiencedomain.SegmentRepository = (*SegmentRepository)(nil)

func (r *SegmentRepository) GetByID(ctx context.Context, id string) (*audiencedomain.AudienceSegment, error) {
	var row persistence.AudienceSegmentRow
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain()
}

func (r *SegmentRepository) GetByName(ctx context.Context, name string) (*audiencedomain.AudienceSegment, error) {
	var row persistence.AudienceSegmentRow
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain()
}

func (r *SegmentRepository) Create(ctx context.Context, seg *audiencedomain.AudienceSegment) error {
	row, err := persistence.AudienceSegmentRowFromDomain(seg)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	seg.ID = row.ID
	seg.CreatedAt = row.CreatedAt
	seg.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *SegmentRepository) Update(ctx context.Context, seg *audiencedomain.AudienceSegment) error {
	row, err := persistence.AudienceSegmentRowFromDomain(seg)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(row).Error
}

// ListAvailableNow returns catalog segments that are active and within their availability window.
func (r *SegmentRepository) ListAvailableNow(ctx context.Context, now time.Time) ([]audiencedomain.AudienceSegment, error) {
	var rows []persistence.AudienceSegmentRow
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND available_from <= ?", true, now).
		Where("available_until IS NULL OR available_until >= ?", now).
		Order("name asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDomainSegments(rows)
}

func (r *SegmentRepository) ListAll(ctx context.Context) ([]audiencedomain.AudienceSegment, error) {
	var rows []persistence.AudienceSegmentRow
	err := r.db.WithContext(ctx).
		Order("name asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDomainSegments(rows)
}

func (r *SegmentRepository) SuspendSegment(ctx context.Context, id string) (*audiencedomain.AudienceSegment, error) {
	var row persistence.AudienceSegmentRow
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	row.IsActive = false
	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain()
}

func toDomainSegments(rows []persistence.AudienceSegmentRow) ([]audiencedomain.AudienceSegment, error) {
	out := make([]audiencedomain.AudienceSegment, len(rows))
	for i := range rows {
		seg, err := rows[i].ToDomain()
		if err != nil {
			return nil, err
		}
		out[i] = *seg
	}
	return out, nil
}
