package infrastructure

import (
	"context"

	"skykin-platform/configs"
	"skykin-platform/internal/intents/model"

	"gorm.io/gorm"
)

type IntentRepository interface {
	Create(ctx context.Context, intent *model.Intent) (*model.Intent, error)
}

type postgresIntentRepository struct {
	db     *gorm.DB
	config *configs.Config
}

func NewIntentRepository(db *gorm.DB, cfg *configs.Config) IntentRepository {
	return &postgresIntentRepository{db: db, config: cfg}
}

func (r *postgresIntentRepository) Create(ctx context.Context, intent *model.Intent) (*model.Intent, error) {
	if err := r.db.WithContext(ctx).Create(intent).Error; err != nil {
		return nil, err
	}
	return intent, nil
}
