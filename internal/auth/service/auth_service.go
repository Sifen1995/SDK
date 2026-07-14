package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"skykin-platform/internal/auth/domain"
	"skykin-platform/internal/auth/dto"
	"skykin-platform/internal/auth/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"skykin-platform/configs"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterDeveloper(ctx context.Context, req dto.DeveloperRegisterRequest) (*domain.Developer, error)
	RegisterApplication(ctx context.Context, devID string, req dto.ApplicationCreateRequest) (*dto.ApplicationResponse, *dto.APIKeyCredentialResponse, error)
	GetApplications(ctx context.Context, devID string) ([]dto.ApplicationResponse, error)
	AuthenticateSDKKey(ctx context.Context, token string) (*domain.Application, error)
	LoginDeveloper(ctx context.Context, req dto.DeveloperLoginRequest) (*dto.LoginResponse, error)
}

type authService struct {
	repo repository.AuthRepository
	cfg  *configs.Config
}

func NewAuthService(repo repository.AuthRepository, cfg *configs.Config) AuthService {
	return &authService{repo: repo, cfg: cfg}
}

func hashKey(key string) string {
	h := sha256.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

type mountaineerClaims struct {
	DeveloperID string `json:"developer_id"`
	jwt.RegisteredClaims
}

func (s *authService) RegisterDeveloper(ctx context.Context, req dto.DeveloperRegisterRequest) (*domain.Developer, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to process security credentials: %w", err)
	}

	dev := &domain.Developer{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.CreateDeveloper(ctx, dev); err != nil {
		return nil, err
	}
	return dev, nil
}

func (s *authService) RegisterApplication(ctx context.Context, devID string, req dto.ApplicationCreateRequest) (*dto.ApplicationResponse, *dto.APIKeyCredentialResponse, error) {
	devUUID, err := uuid.Parse(devID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid developer id format")
	}

	app := &domain.Application{
		DeveloperID: devUUID,
		AppName:     req.AppName,
		Platform:    req.Platform,
		BundleID:    req.BundleID,
		Status:      "active",
	}

	if err := s.repo.CreateApplication(ctx, app); err != nil {
		return nil, nil, err
	}

	pBytes := make([]byte, 20)
	rand.Read(pBytes)
	pubKey := fmt.Sprintf("pk_live_%s", hex.EncodeToString(pBytes))

	sBytes := make([]byte, 32)
	rand.Read(sBytes)
	secretKey := fmt.Sprintf("sk_secret_%s", hex.EncodeToString(sBytes))

	hashedPubKey := hashKey(pubKey)

	// Secret key is stored in recoverable form for HMAC request verification.
	// Publishable key is hashed so DB lookup never stores the raw pk_live_ value.
	apiKeyRecord := &domain.APIKey{
		ApplicationID:  app.ID,
		KeyValue:       hashedPubKey,
		SecretKeyValue: secretKey,
		IsActive:       true,
		RateLimit:      120,
	}

	if err := s.repo.CreateAPIKey(ctx, apiKeyRecord); err != nil {
		return nil, nil, err
	}

	return &dto.ApplicationResponse{
			ID:        app.ID.String(),
			AppName:   app.AppName,
			Platform:  app.Platform,
			BundleID:  app.BundleID,
			Status:    app.Status,
			CreatedAt: app.CreatedAt,
		}, &dto.APIKeyCredentialResponse{
			ApplicationID:  app.ID.String(),
			PublishableKey: pubKey,
			RawSecretKey:   secretKey,
			RateLimit:      apiKeyRecord.RateLimit,
		}, nil
}
func (s *authService) AuthenticateSDKKey(ctx context.Context, token string) (*domain.Application, error) {
	key, app, err := s.repo.VerifyAPIKey(ctx, token)
	if err != nil {
		return nil, errors.New("invalid or revoked SDK credentials")
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("provided access key token has expired")
	}

	return app, nil
}
func (s *authService) LoginDeveloper(ctx context.Context, req dto.DeveloperLoginRequest) (*dto.LoginResponse, error) {
	dev, err := s.repo.GetDeveloperByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email address or account password credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(dev.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email address or account password credentials")
	}

	claims := mountaineerClaims{
		DeveloperID: dev.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	tokenPayload := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := tokenPayload.SignedString([]byte(s.cfg.JwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure access token session: %w", err)
	}

	return &dto.LoginResponse{
		Token: signedToken,
		Developer: map[string]interface{}{
			"id":    dev.ID.String(),
			"name":  dev.Name,
			"email": dev.Email,
		},
	}, nil
}

func (s *authService) GetApplications(ctx context.Context, devID string) ([]dto.ApplicationResponse, error) {
	apps, err := s.repo.GetApplicationsByDeveloper(ctx, devID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ApplicationResponse, len(apps))
	for i, app := range apps {
		result[i] = dto.ApplicationResponse{
			ID:        app.ID.String(),
			AppName:   app.AppName,
			Platform:  app.Platform,
			BundleID:  app.BundleID,
			Status:    app.Status,
			CreatedAt: app.CreatedAt,
		}
	}
	return result, nil
}
