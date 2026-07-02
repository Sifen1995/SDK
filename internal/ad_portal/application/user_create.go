package application

import (
	"context"
	"errors"
	"fmt"

	"skykin-platform/internal/ad_portal/domain"
)

func (s *AuthService) EnsureOperatorAdmin(ctx context.Context, email, password, name, company string) error {
	_, err := s.repo.GetPortalUserByEmail(ctx, email)
	if err == nil {
		return nil
	}
	role, err := s.repo.GetRoleBySlug(ctx, domain.RoleOperatorAdmin)
	if err != nil {
		return err
	}
	hash, err := bcryptHashPassword(password)
	if err != nil {
		return err
	}
	u := &domain.PortalUser{
		Email:        email,
		PasswordHash: hash,
		Name:         name,
		RoleID:       role.ID,
		IsActive:     true,
	}
	return s.repo.CreatePortalUser(ctx, u)
}

func (s *AuthService) createOperatorAdminUser(ctx context.Context, name, email, password, roleSlug string) (*domain.PortalUser, error) {
	role, err := s.repo.GetRoleBySlug(ctx, roleSlug)
	if err != nil {
		return nil, err
	}
	hash, err := bcryptHashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &domain.PortalUser{
		Email: email, PasswordHash: hash, Name: name, RoleID: role.ID, IsActive: true,
	}
	if err := s.repo.CreatePortalUser(ctx, u); err != nil {
		return nil, err
	}
	return s.repo.GetPortalUserByID(ctx, u.ID)
}

func (s *AuthService) createPortalUser(ctx context.Context, name, email, password, company, roleSlug string) (*domain.PortalUser, error) {
	if company == "" {
		return nil, errors.New("company_name is required")
	}
	role, err := s.repo.GetRoleBySlug(ctx, roleSlug)
	if err != nil {
		return nil, fmt.Errorf("invalid role")
	}
	hash, err := bcryptHashPassword(password)
	if err != nil {
		return nil, err
	}
	adv := &domain.Advertiser{CompanyName: company}
	u := &domain.PortalUser{
		Email:        email,
		PasswordHash: hash,
		Name:         name,
		RoleID:       role.ID,
		IsActive:     true,
	}
	if err := s.repo.CreateAdvertiserAndPortalUser(ctx, adv, u); err != nil {
		return nil, err
	}
	return s.repo.GetPortalUserByID(ctx, u.ID)
}
