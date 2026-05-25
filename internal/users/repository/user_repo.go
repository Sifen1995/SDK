package repository

import (
	"context"
	"skykin-platform/configs"
	"skykin-platform/internal/users/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindOrCreate(ctx context.Context, externalUserID string) (*model.Users, error)
}

type userRepo struct {
	db     *gorm.DB
	config *configs.Config
}

func NewUserRepository(db *gorm.DB, cfg *configs.Config) UserRepository {
	return &userRepo{db: db, config: cfg}
}

func (r *userRepo) FindOrCreate(ctx context.Context, externalUserID string) (*model.Users, error) {
	var user model.Users
	err := r.db.WithContext(ctx).Where("external_user_id = ?", externalUserID).First(&user).Error
	if err == nil {
		return &user, nil // User found
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err // Some other error occurred
	}

	// User not found, create a new one
	user = model.Users{ExternalUserID: externalUserID}
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err // Error creating user
	}
	return &user, nil // New user created
}
