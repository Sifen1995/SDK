package service_test

import (
	"context"
	"testing"

	"skykin-platform/configs"
	"skykin-platform/internal/auth/domain"
	dto "skykin-platform/internal/auth/dto"
	authService "skykin-platform/internal/auth/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type MockAuthRepository struct {
	GetDeveloperByEmailFunc          func(ctx context.Context, email string) (*domain.Developer, error)
	CreateDeveloperFunc              func(ctx context.Context, dev *domain.Developer) error
	CreateApplicationFunc            func(ctx context.Context, app *domain.Application) error
	CreateAPIKeyFunc                 func(ctx context.Context, key *domain.APIKey) error
	VerifyAPIKeyFunc                 func(ctx context.Context, token string) (*domain.APIKey, *domain.Application, error)
	GetApplicationsByDeveloperFunc   func(ctx context.Context, devID string) ([]domain.Application, error)
}

func (m *MockAuthRepository) GetDeveloperByEmail(ctx context.Context, email string) (*domain.Developer, error) {
	return m.GetDeveloperByEmailFunc(ctx, email)
}
func (m *MockAuthRepository) CreateDeveloper(ctx context.Context, dev *domain.Developer) error {
	return m.CreateDeveloperFunc(ctx, dev)
}
func (m *MockAuthRepository) CreateApplication(ctx context.Context, app *domain.Application) error {
	return m.CreateApplicationFunc(ctx, app)
}
func (m *MockAuthRepository) CreateAPIKey(ctx context.Context, key *domain.APIKey) error {
	return m.CreateAPIKeyFunc(ctx, key)
}
func (m *MockAuthRepository) VerifyAPIKey(ctx context.Context, token string) (*domain.APIKey, *domain.Application, error) {
	return m.VerifyAPIKeyFunc(ctx, token)
}
func (m *MockAuthRepository) GetApplicationsByDeveloper(ctx context.Context, devID string) ([]domain.Application, error) {
	return m.GetApplicationsByDeveloperFunc(ctx, devID)
}

func TestLoginDeveloper_Success(t *testing.T) {
	mockRepo := &MockAuthRepository{}
	cfg := &configs.Config{JwtSecret: "test_secret_key"}
	service := authService.NewAuthService(mockRepo, cfg)

	plainPassword := "securePassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	mockRepo.GetDeveloperByEmailFunc = func(ctx context.Context, email string) (*domain.Developer, error) {
		return &domain.Developer{
			ID:           uuid.New(),
			Name:         "Sifen Getachew",
			Email:        "sifen@skykin.com",
			PasswordHash: string(hashedPassword),
		}, nil
	}

	req := dto.DeveloperLoginRequest{
		Email:    "sifen@skykin.com",
		Password: plainPassword,
	}

	response, err := service.LoginDeveloper(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, "Sifen Getachew", response.Developer["name"])
}

func TestLoginDeveloper_InvalidPassword(t *testing.T) {
	mockRepo := &MockAuthRepository{}
	cfg := &configs.Config{JwtSecret: "test_secret_key"}
	service := authService.NewAuthService(mockRepo, cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctPassword"), bcrypt.DefaultCost)

	mockRepo.GetDeveloperByEmailFunc = func(ctx context.Context, email string) (*domain.Developer, error) {
		return &domain.Developer{
			ID:           uuid.New(),
			Email:        "sifen@skykin.com",
			PasswordHash: string(hashedPassword),
		}, nil
	}

	req := dto.DeveloperLoginRequest{
		Email:    "sifen@skykin.com",
		Password: "wrongPassword!!",
	}

	response, err := service.LoginDeveloper(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "invalid email address or account password credentials", err.Error())
}
