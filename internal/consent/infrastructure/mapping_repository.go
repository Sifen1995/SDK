package infrastructure

import (
	"context"
	"strings"

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

// FindOneDemoPseudonymousID picks a random active demo SMS recipient's mapping.
// Only users in demo_sms_recipients are returned so ingest-ad can mock-dispatch.
func (r *PseudonymousMappingRepository) FindOneDemoPseudonymousID(ctx context.Context) (string, error) {
	// Scan into string, not uuid.UUID. uuid.UUID is [16]byte, and GORM dispatches
	// on the destination's concrete type: an array kind is treated as a 16-element
	// uint8 result set rather than going through uuid's sql.Scanner, so the UUID
	// text is converted into a uint8 and every call fails.
	var raw string
	tx := r.db.WithContext(ctx).
		Table("demo_sms_recipients AS rec").
		Select("COALESCE(rec.pseudonymous_id, pm.pseudonymous_id) AS pseudonymous_id").
		Joins("INNER JOIN pseudonymous_mappings pm ON pm.user_id = rec.user_id").
		Where("rec.is_active = TRUE").
		Order("RANDOM()").
		Limit(1).
		Scan(&raw)
	if tx.Error != nil {
		return "", tx.Error
	}
	if tx.RowsAffected == 0 || strings.TrimSpace(raw) == "" {
		return "", gorm.ErrRecordNotFound
	}
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
