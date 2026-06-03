package infrastructure

import (
	"context"

	"skykin-platform/internal/users/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindOrCreate(ctx context.Context, externalUserID string) (*model.Users, error)
}

type postgresUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) FindOrCreate(ctx context.Context, externalUserID string) (*model.Users, error) {
	var user model.Users
	err := r.db.WithContext(ctx).
		Where("external_user_id = ?", externalUserID).
		FirstOrCreate(&user, model.Users{ExternalUserID: externalUserID}).
		Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
