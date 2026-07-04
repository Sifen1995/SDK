package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID
	Name        string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsSystem    bool
	Permissions []*Permission
	CreatedAt   time.Time
}

type PermissionRepository interface {
	FindAll(ctx context.Context) ([]*Permission, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Permission, error)
}

type RoleRepository interface {
	FindAll(ctx context.Context) ([]*Role, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)
	FindByName(ctx context.Context, name string) (*Role, error)
	Create(ctx context.Context, role *Role) error
	GetPermissionNames(ctx context.Context, roleName string) ([]string, error)
	AssignPermission(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID, grantedBy uuid.UUID) error
	RevokePermission(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error
}
