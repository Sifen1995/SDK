package infrastructure

import (
	"context"

	"skykin-platform/internal/users/domain"
	"skykin-platform/internal/users/infrastructure/persistence"

	"gorm.io/gorm"
)

type postgresUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) FindOrCreate(ctx context.Context, externalUserID string) (*domain.User, error) {
	var row persistence.UserRow
	err := r.db.WithContext(ctx).
		Where("external_user_id = ?", externalUserID).
		FirstOrCreate(&row, persistence.UserRow{ExternalUserID: externalUserID}).
		Error
	if err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}
