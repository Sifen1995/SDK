package repository

import (
	"context"
	"skykin-platform/configs"
	"skykin-platform/internal/auth/domain"
	"skykin-platform/internal/auth/infrastructure/persistence"

	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateDeveloper(ctx context.Context, dev *domain.Developer) error
	GetDeveloperByEmail(ctx context.Context, email string) (*domain.Developer, error)
	CreateApplication(ctx context.Context, app *domain.Application) error
	GetApplicationsByDeveloper(ctx context.Context, devID string) ([]domain.Application, error)
	CreateAPIKey(ctx context.Context, key *domain.APIKey) error
	VerifyAPIKey(ctx context.Context, token string) (*domain.APIKey, *domain.Application, error)
}

type authRepository struct {
	db  *gorm.DB
	cfg *configs.Config
}

func NewAuthRepository(db *gorm.DB, cfg *configs.Config) AuthRepository {
	return &authRepository{db: db, cfg: cfg}
}

func (r *authRepository) CreateDeveloper(ctx context.Context, dev *domain.Developer) error {
	row := persistence.DeveloperRowFromDomain(dev)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	*dev = *row.ToDomain()
	return nil
}

func (r *authRepository) GetDeveloperByEmail(ctx context.Context, email string) (*domain.Developer, error) {
	var row persistence.DeveloperRow
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&row).Error
	if err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
}

func (r *authRepository) CreateApplication(ctx context.Context, app *domain.Application) error {
	row := persistence.ApplicationRowFromDomain(app)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	*app = *row.ToDomain()
	return nil
}

func (r *authRepository) GetApplicationsByDeveloper(ctx context.Context, devID string) ([]domain.Application, error) {
	var rows []persistence.ApplicationRow
	err := r.db.WithContext(ctx).Where("developer_id = ?", devID).Preload("APIKeys").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	apps := make([]domain.Application, len(rows))
	for i := range rows {
		apps[i] = *rows[i].ToDomain()
	}
	return apps, nil
}

func (r *authRepository) CreateAPIKey(ctx context.Context, key *domain.APIKey) error {
	row := persistence.APIKeyRowFromDomain(key)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	*key = *row.ToDomain()
	return nil
}

func (r *authRepository) VerifyAPIKey(ctx context.Context, token string) (*domain.APIKey, *domain.Application, error) {
	var keyRow persistence.APIKeyRow
	var appRow persistence.ApplicationRow

	err := r.db.WithContext(ctx).Where("key_value = ? AND is_active = ?", token, true).First(&keyRow).Error
	if err != nil {
		return nil, nil, err
	}

	err = r.db.WithContext(ctx).Where("id = ? AND status = ?", keyRow.ApplicationID, "active").First(&appRow).Error
	if err != nil {
		return nil, nil, err
	}

	return keyRow.ToDomain(), appRow.ToDomain(), nil
}
