package repository

import (
	"context"
	"skykin-platform/configs"
	"skykin-platform/internal/intents/model"

	"gorm.io/gorm"
)

type IntentRepository interface {
	Create(ctx context.Context, intent *model.Intent) (*model.Intent, error)
}

type intentRepo struct {
	db     *gorm.DB
	config *configs.Config
}

func NewIntentRepository(db *gorm.DB, cfg *configs.Config) IntentRepository {
	return &intentRepo{db: db, config: cfg}
}

func (r *intentRepo) Create(ctx context.Context, intent *model.Intent) (*model.Intent, error) {
	err := r.db.WithContext(ctx).Create(intent).Error
	if err != nil {
		return nil, err
	}
	return intent, nil
}
