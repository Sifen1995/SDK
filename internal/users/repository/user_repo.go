package repository

import (
	"context"
	"skykin-platform/internal/users/model" // Ensure this path points to your Users model

	"gorm.io/gorm"
)

type UserRepository interface {
	FindOrCreate(ctx context.Context, externalUserID string) (*model.Users, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// FindOrCreate automatically fetches an existing user or creates a new one cleanly
func (r *userRepository) FindOrCreate(ctx context.Context, externalUserID string) (*model.Users, error) {
	var user model.Users

	// GORM FirstOrCreate method natively checks for existence, handles thread safety,
	// and drops an entry if missing using default values (like gen_random_uuid)
	err := r.db.WithContext(ctx).
		Where("external_user_id = ?", externalUserID).
		FirstOrCreate(&user, model.Users{ExternalUserID: externalUserID}).
		Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
