package infrastructure

import (
	"context"

	"skykin-platform/internal/consent/domain"
	"skykin-platform/internal/consent/infrastructure/persistence"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PseudonymousMappingRepository struct {
	db *gorm.DB
}

func NewPseudonymousMappingRepository(db *gorm.DB) *PseudonymousMappingRepository {
	return &PseudonymousMappingRepository{db: db}
}

var _ domain.PseudonymousMappingRepository = (*PseudonymousMappingRepository)(nil)

func (r *PseudonymousMappingRepository) Create(ctx context.Context, mapping *domain.PseudonymousMapping) error {
	row := persistence.MappingRowFromDomain(mapping)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	mapping.ID = row.ID
	mapping.CreatedAt = row.CreatedAt
	return nil
}

func (r *PseudonymousMappingRepository) FindByPseudonymousID(
	ctx context.Context,
	pseudonymousID uuid.UUID,
) (*domain.PseudonymousMapping, error) {
	var row persistence.PseudonymousMappingRow
	err := r.db.WithContext(ctx).Where("pseudonymous_id = ?", pseudonymousID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *PseudonymousMappingRepository) FindByUserID(
	ctx context.Context,
	userID int64,
) (*domain.PseudonymousMapping, error) {
	var row persistence.PseudonymousMappingRow
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *PseudonymousMappingRepository) FindPseudonymousIDsByUserIDs(
	ctx context.Context,
	userIDs []int64,
) (map[int64]string, error) {
	if len(userIDs) == 0 {
		return map[int64]string{}, nil
	}
	var rows []persistence.PseudonymousMappingRow
	if err := r.db.WithContext(ctx).
		Where("user_id IN ?", userIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[int64]string, len(rows))
	for i := range rows {
		out[rows[i].UserID] = rows[i].PseudonymousID.String()
	}
	return out, nil
}
