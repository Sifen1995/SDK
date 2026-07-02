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

func (r *postgresUserRepository) FindAll(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.User, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&persistence.UserRow{}).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []persistence.UserRow
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(rows))
	for i := range rows {
		users[i] = rows[i].ToDomain()
	}
	return users, total, nil
}
