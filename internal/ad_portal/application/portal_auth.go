package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"skykin-platform/configs"
	"skykin-platform/internal/ad_portal/domain"
	"skykin-platform/internal/ad_portal/infrastructure"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenTTL = 24 * time.Hour

type AuthService struct {
	repo *infrastructure.Repository
	cfg  *configs.Config
}

func NewAuthService(repo *infrastructure.Repository, cfg *configs.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

type RegisterResult struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type UserInfo struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type LoginResult struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresAt    string   `json:"expires_at"`
	User         UserInfo `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*RegisterResult, error) {
	u, err := s.createPortalUser(ctx, name, email, password, name, domain.RoleAdvertiser)
	if err != nil {
		return nil, err
	}
	return registerResultFromUser(u), nil
}

func registerResultFromUser(u *domain.PortalUser) *RegisterResult {
	return &RegisterResult{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.RoleSlug(),
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func ParsePortalToken(cfg *configs.Config, tokenStr string) (*PortalClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &PortalClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*PortalClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

func (s *AuthService) Me(ctx context.Context, portalUserID string) (*domain.PortalUser, error) {
	return s.repo.GetPortalUserByID(ctx, portalUserID)
}

func UserResponse(u *domain.PortalUser) map[string]any {
	resp := map[string]any{
		"id":        u.ID,
		"email":     u.Email,
		"name":      u.Name,
		"role":      u.RoleSlug(),
		"role_id":   u.RoleID,
		"is_active": u.IsActive,
	}
	switch u.RoleSlug() {
	case domain.RoleAdvertiser:
		if id := u.AccountAdvertiserID(); id != "" {
			resp["advertiser_id"] = id
		}
		if u.Advertiser != nil {
			resp["company_name"] = u.Advertiser.CompanyName
		}
	case domain.RoleReadOnlyAnalyst:
		if id := u.AccountAnalystID(); id != "" {
			resp["analyst_id"] = id
		}
	}
	return resp
}

func (s *AuthService) CreateOperatorUser(ctx context.Context, name, email, password, roleSlug, company string) (*domain.PortalUser, error) {
	switch roleSlug {
	case domain.RoleOperatorAdmin:
		return s.createOperatorAdminUser(ctx, name, email, password, roleSlug)
	case domain.RoleAdvertiser:
		return s.createAdvertiserUser(ctx, name, email, password, company, roleSlug)
	case domain.RoleReadOnlyAnalyst:
		return s.createAnalystUser(ctx, name, email, password, roleSlug)
	default:
		return nil, fmt.Errorf("invalid role")
	}
}
