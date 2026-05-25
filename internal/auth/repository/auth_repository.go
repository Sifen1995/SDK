package repository

import (
	"context"
	"skykin-platform/configs"
	"skykin-platform/internal/auth/model"

	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateDeveloper(ctx context.Context, dev *model.Developer) error
	GetDeveloperByEmail(ctx context.Context, email string) (*model.Developer, error)
	CreateApplication(ctx context.Context, app *model.Application) error
	GetApplicationsByDeveloper(ctx context.Context, devID string) ([]model.Application, error)
	CreateAPIKey(ctx context.Context, key *model.APIKey) error
	VerifyAPIKey(ctx context.Context, token string) (*model.APIKey, *model.Application, error)
}

type authRepository struct {
	db  *gorm.DB
	cfg *configs.Config
}

func NewAuthRepository(db *gorm.DB, cfg *configs.Config) AuthRepository {
	return &authRepository{db: db, cfg: cfg}
}

func (r *authRepository) CreateDeveloper(ctx context.Context, dev *model.Developer) error {
	return r.db.WithContext(ctx).Create(dev).Error
}

func (r *authRepository) GetDeveloperByEmail(ctx context.Context, email string) (*model.Developer, error) {
	var dev model.Developer
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&dev).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

func (r *authRepository) CreateApplication(ctx context.Context, app *model.Application) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *authRepository) GetApplicationsByDeveloper(ctx context.Context, devID string) ([]model.Application, error) {
	var apps []model.Application
	err := r.db.WithContext(ctx).Where("developer_id = ?", devID).Preload("APIKeys").Find(&apps).Error
	return apps, err
}

func (r *authRepository) CreateAPIKey(ctx context.Context, key *model.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

// VerifyAPIKey pulls both the key metadata and joins the application status context in one clean optimized trip
func (r *authRepository) VerifyAPIKey(ctx context.Context, token string) (*model.APIKey, *model.Application, error) {
	var key model.APIKey
	var app model.Application

	err := r.db.WithContext(ctx).Where("key_value = ? AND is_active = ?", token, true).First(&key).Error
	if err != nil {
		return nil, nil, err
	}

	err = r.db.WithContext(ctx).Where("id = ? AND status = ?", key.ApplicationID, "active").First(&app).Error
	if err != nil {
		return nil, nil, err
	}

	return &key, &app, nil
}
