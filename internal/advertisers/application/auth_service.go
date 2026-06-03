package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"skykin-platform/configs"
	"skykin-platform/internal/advertisers/domain"
	"skykin-platform/internal/advertisers/infrastructure"
	"skykin-platform/internal/advertisers/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *infrastructure.Repository
	cfg  *configs.Config
}

func NewAuthService(repo *infrastructure.Repository, cfg *configs.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

type portalClaims struct {
	AdvertiserID string `json:"advertiser_id"`
	Role         string `json:"role"`
	Email        string `json:"email"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(ctx context.Context, contactName, email, password, company, role string) (*model.Advertiser, error) {
	if role == "" {
		role = domain.RoleAdvertiser
	}
	if role != domain.RoleAdvertiser && role != domain.RoleReadOnlyAnalyst {
		return nil, errors.New("registration only allowed for advertiser or read_only_analyst roles")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	apiKey, err := s.generateAPIKey(ctx)
	if err != nil {
		return nil, err
	}

	a := &model.Advertiser{
		CompanyName:  company,
		Email:        email,
		PasswordHash: string(hash),
		APIKey:       apiKey,
		Role:         role,
		ContactName:  contactName,
		IsActive:     true,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *model.Advertiser, error) {
	a, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	if !a.IsActive {
		return "", nil, errors.New("account is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	token, err := s.signToken(a)
	if err != nil {
		return "", nil, err
	}
	return token, a, nil
}

func (s *AuthService) signToken(a *model.Advertiser) (string, error) {
	claims := portalClaims{
		AdvertiserID: a.ID,
		Role:         a.Role,
		Email:        a.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.cfg.JwtSecret))
}

func ParsePortalToken(cfg *configs.Config, tokenStr string) (*portalClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &portalClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*portalClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

func (s *AuthService) EnsureOperatorAdmin(ctx context.Context, email, password, contactName, company string) error {
	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	apiKey, err := s.generateAPIKey(ctx)
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, &model.Advertiser{
		CompanyName:  company,
		Email:        email,
		PasswordHash: string(hash),
		APIKey:       apiKey,
		Role:         domain.RoleOperatorAdmin,
		ContactName:  contactName,
		IsActive:     true,
	})
}

func (s *AuthService) Me(ctx context.Context, userID string) (*model.Advertiser, error) {
	return s.repo.GetByID(ctx, userID)
}

func UserResponse(a *model.Advertiser) map[string]any {
	return map[string]any{
		"id":            a.ID,
		"company_name":  a.CompanyName,
		"email":         a.Email,
		"contact_name":  a.ContactName,
		"role":          a.Role,
		"api_key":       a.APIKey,
		"is_active":     a.IsActive,
	}
}

func (s *AuthService) CreateOperatorUser(ctx context.Context, contactName, email, password, role, company string) (*model.Advertiser, error) {
	if role != domain.RoleAdvertiser && role != domain.RoleReadOnlyAnalyst && role != domain.RoleOperatorAdmin {
		return nil, fmt.Errorf("invalid role")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	apiKey, err := s.generateAPIKey(ctx)
	if err != nil {
		return nil, err
	}
	a := &model.Advertiser{
		CompanyName:  company,
		Email:        email,
		PasswordHash: string(hash),
		APIKey:       apiKey,
		Role:         role,
		ContactName:  contactName,
		IsActive:     true,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AuthService) generateAPIKey(ctx context.Context) (string, error) {
	for i := 0; i < 5; i++ {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		key := hex.EncodeToString(b)
		exists, err := s.repo.APIKeyExists(ctx, key)
		if err != nil {
			return "", err
		}
		if !exists {
			return key, nil
		}
	}
	return "", errors.New("failed to generate unique api key")
}
