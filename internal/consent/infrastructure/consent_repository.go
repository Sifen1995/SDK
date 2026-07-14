package infrastructure

import (
	"context"

	"skykin-platform/internal/consent/domain"
	"skykin-platform/internal/consent/infrastructure/persistence"

	"gorm.io/gorm"
)

type ConsentRepository struct {
	db *gorm.DB
}

func NewConsentRepository(db *gorm.DB) *ConsentRepository {
	return &ConsentRepository{db: db}
}

var _ domain.ConsentRepository = (*ConsentRepository)(nil)

func (r *ConsentRepository) Create(ctx context.Context, consent *domain.Consent) error {
	row := persistence.ConsentRow{}
	row.FromDomain(consent)
	if err := row.Validate(); err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	consent.ID = row.ID
	consent.CreatedAt = row.CreatedAt
	consent.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *ConsentRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Consent, error) {
	var row persistence.ConsentRow
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").First(&row).Error
	if err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *ConsentRepository) Update(ctx context.Context, consent *domain.Consent) error {
	row := persistence.ConsentRow{}
	row.FromDomain(consent)
	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return err
	}
	consent.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *ConsentRepository) ListActive(ctx context.Context) ([]domain.Consent, error) {
	var rows []persistence.ConsentRow
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Consent, len(rows))
	for i := range rows {
		out[i] = *rows[i].ToDomain()
	}
	return out, nil
}
