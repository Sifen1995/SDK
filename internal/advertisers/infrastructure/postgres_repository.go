package infrastructure

import (
	"context"

	"skykin-platform/internal/advertisers/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, a *model.Advertiser) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*model.Advertiser, error) {
	var a model.Advertiser
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.Advertiser, error) {
	var a model.Advertiser
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) APIKeyExists(ctx context.Context, key string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Advertiser{}).Where("api_key = ?", key).Count(&n).Error
	return n > 0, err
}
