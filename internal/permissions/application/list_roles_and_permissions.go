package application

import (
	"context"

	"skykin-platform/internal/permissions/domain"
)

type RoleLister interface {
	FindAll(ctx context.Context) ([]*domain.Role, error)
}

type PermissionLister interface {
	FindAll(ctx context.Context) ([]*domain.Permission, error)
}

type ListRolesUseCase struct {
	roleRepo RoleLister
}

func NewListRolesUseCase(roleRepo RoleLister) *ListRolesUseCase {
	return &ListRolesUseCase{roleRepo: roleRepo}
}

func (uc *ListRolesUseCase) List(ctx context.Context) ([]*domain.Role, error) {
	return uc.roleRepo.FindAll(ctx)
}

type ListPermissionsUseCase struct {
	permRepo PermissionLister
}

func NewListPermissionsUseCase(permRepo PermissionLister) *ListPermissionsUseCase {
	return &ListPermissionsUseCase{permRepo: permRepo}
}

func (uc *ListPermissionsUseCase) List(ctx context.Context) ([]*domain.Permission, error) {
	return uc.permRepo.FindAll(ctx)
}
