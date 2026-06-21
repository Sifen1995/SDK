package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"skykin-platform/configs"
	"skykin-platform/internal/ad_portal/domain"
	"skykin-platform/internal/ad_portal/infrastructure"
	"skykin-platform/internal/ad_portal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// PortalAuthService handles registration, login, and portal user management for all ad-portal roles.
type PortalAuthService struct {
	repo *infrastructure.Repository
	cfg  *configs.Config
}

func NewPortalAuthService(repo *infrastructure.Repository, cfg *configs.Config) *PortalAuthService {
	return &PortalAuthService{repo: repo, cfg: cfg}
}

type portalClaims struct {
	PortalUserID string `json:"portal_user_id"`
	AdvertiserID string `json:"advertiser_id"`
	Role         string `json:"role"`
	Email        string `json:"email"`
	jwt.RegisteredClaims
}

// Register creates an advertiser or read-only analyst portal user.
func (s *PortalAuthService) Register(ctx context.Context, name, email, password, company, roleSlug string) (*model.PortalUser, error) {
	if roleSlug == "" {
		roleSlug = domain.RoleAdvertiser
	}
	if roleSlug != domain.RoleAdvertiser && roleSlug != domain.RoleReadOnlyAnalyst {
		return nil, errors.New("registration only allowed for advertiser or read_only_analyst roles")
	}
	return s.createPortalUser(ctx, name, email, password, company, roleSlug)
}

// Login authenticates any portal role (operator_admin, advertiser, read_only_analyst).
func (s *PortalAuthService) Login(ctx context.Context, email, password string) (string, *model.PortalUser, error) {
	u, err := s.repo.GetPortalUserByEmail(ctx, email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	if !u.IsActive {
		return "", nil, errors.New("account is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	token, err := s.signToken(u)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (s *PortalAuthService) signToken(u *model.PortalUser) (string, error) {
	claims := portalClaims{
		PortalUserID: u.ID,
		AdvertiserID: u.AccountAdvertiserID(),
		Role:         u.RoleSlug(),
		Email:        u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.cfg.JwtSecret))
}

// ParsePortalToken validates a portal JWT and returns its claims.
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

// EnsureOperatorAdmin seeds the default operator admin if missing.
func (s *PortalAuthService) EnsureOperatorAdmin(ctx context.Context, email, password, name, company string) error {
	_, err := s.repo.GetPortalUserByEmail(ctx, email)
	if err == nil {
		return nil
	}
	role, err := s.repo.GetRoleBySlug(ctx, domain.RoleOperatorAdmin)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &model.PortalUser{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		RoleID:       role.ID,
		IsActive:     true,
	}
	return s.repo.CreatePortalUser(ctx, u)
}

// Me returns the current portal user profile.
func (s *PortalAuthService) Me(ctx context.Context, portalUserID string) (*model.PortalUser, error) {
	return s.repo.GetPortalUserByID(ctx, portalUserID)
}

// UserResponse maps a portal user to the API shape.
func UserResponse(u *model.PortalUser) map[string]any {
	resp := map[string]any{
		"id":            u.ID,
		"email":         u.Email,
		"name":          u.Name,
		"role":          u.RoleSlug(),
		"role_id":       u.RoleID,
		"advertiser_id": u.AccountAdvertiserID(),
		"is_active":     u.IsActive,
	}
	if u.Advertiser != nil {
		resp["company_name"] = u.Advertiser.CompanyName
	}
	return resp
}

// CreateOperatorUser is used by operator_admin to provision portal accounts.
func (s *PortalAuthService) CreateOperatorUser(ctx context.Context, name, email, password, roleSlug, company string) (*model.PortalUser, error) {
	if roleSlug != domain.RoleAdvertiser && roleSlug != domain.RoleReadOnlyAnalyst && roleSlug != domain.RoleOperatorAdmin {
		return nil, fmt.Errorf("invalid role")
	}
	if roleSlug == domain.RoleOperatorAdmin {
		role, err := s.repo.GetRoleBySlug(ctx, roleSlug)
		if err != nil {
			return nil, err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		u := &model.PortalUser{
			Email: email, PasswordHash: string(hash), Name: name, RoleID: role.ID, IsActive: true,
		}
		if err := s.repo.CreatePortalUser(ctx, u); err != nil {
			return nil, err
		}
		return s.repo.GetPortalUserByID(ctx, u.ID)
	}
	return s.createPortalUser(ctx, name, email, password, company, roleSlug)
}

func (s *PortalAuthService) createPortalUser(ctx context.Context, name, email, password, company, roleSlug string) (*model.PortalUser, error) {
	if company == "" {
		return nil, errors.New("company_name is required")
	}
	role, err := s.repo.GetRoleBySlug(ctx, roleSlug)
	if err != nil {
		return nil, fmt.Errorf("invalid role")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	adv := &model.Advertiser{CompanyName: company}
	u := &model.PortalUser{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		RoleID:       role.ID,
		IsActive:     true,
	}
	if err := s.repo.CreateAdvertiserAndPortalUser(ctx, adv, u); err != nil {
		return nil, err
	}
	return s.repo.GetPortalUserByID(ctx, u.ID)
}
