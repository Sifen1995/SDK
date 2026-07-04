package application

import (
	"context"
	"errors"
	"log/slog"

	"skykin-platform/internal/permissions/domain"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
)

type RevokePermissionUseCase struct {
	roleRepo interface {
		FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
		RevokePermission(ctx context.Context, roleID, permissionID uuid.UUID) error
	}
	permissionRepo interface {
		FindByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error)
	}
	cache interface {
		Invalidate(roleName string)
	}
	bus interface {
		Publish(event messaging.Event)
	}
	logger *slog.Logger
}

func NewRevokePermissionUseCase(
	roleRepo interface {
		FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
		RevokePermission(ctx context.Context, roleID, permissionID uuid.UUID) error
	},
	permissionRepo interface {
		FindByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error)
	},
	cache interface {
		Invalidate(roleName string)
	},
	bus interface {
		Publish(event messaging.Event)
	},
	logger *slog.Logger,
) *RevokePermissionUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &RevokePermissionUseCase{
		roleRepo: roleRepo, permissionRepo: permissionRepo,
		cache: cache, bus: bus, logger: logger,
	}
}

func (uc *RevokePermissionUseCase) Execute(
	ctx context.Context,
	roleID uuid.UUID,
	permissionID uuid.UUID,
) error {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return errors.New("role not found")
	}
	if role.Name == "operator_admin" {
		return errors.New("operator_admin permissions cannot be modified")
	}
	perm, err := uc.permissionRepo.FindByID(ctx, permissionID)
	if err != nil {
		return errors.New("permission not found")
	}
	if err := uc.roleRepo.RevokePermission(ctx, roleID, permissionID); err != nil {
		return err
	}
	uc.cache.Invalidate(role.Name)
	uc.bus.Publish(messaging.Event{
		Name: domain.EventPermissionRevoked,
		Payload: domain.PermissionRevokedPayload{
			RoleID: role.ID, RoleName: role.Name,
			PermissionID: perm.ID, Permission: perm.Name,
		},
		Ctx: ctx,
	})
	uc.logger.Info("permission revoked", "role", role.Name, "permission", perm.Name)
	return nil
}
