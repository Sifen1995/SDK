package application

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"skykin-platform/internal/permissions/domain"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
)

type AssignPermissionUseCase struct {
	roleRepo interface {
		FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
		AssignPermission(ctx context.Context, roleID, permissionID, grantedBy uuid.UUID) error
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

func NewAssignPermissionUseCase(
	roleRepo interface {
		FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
		AssignPermission(ctx context.Context, roleID, permissionID, grantedBy uuid.UUID) error
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
) *AssignPermissionUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &AssignPermissionUseCase{
		roleRepo: roleRepo, permissionRepo: permissionRepo,
		cache: cache, bus: bus, logger: logger,
	}
}

func (uc *AssignPermissionUseCase) Execute(
	ctx context.Context,
	roleID uuid.UUID,
	permissionID uuid.UUID,
	grantedBy uuid.UUID,
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
	if err := uc.roleRepo.AssignPermission(ctx, roleID, permissionID, grantedBy); err != nil {
		if isDuplicateErr(err) {
			return nil
		}
		return err
	}
	uc.cache.Invalidate(role.Name)
	uc.bus.Publish(messaging.Event{
		Name: domain.EventPermissionAssigned,
		Payload: domain.PermissionAssignedPayload{
			RoleID: role.ID, RoleName: role.Name,
			PermissionID: perm.ID, Permission: perm.Name, GrantedBy: grantedBy,
		},
		Ctx: ctx,
	})
	uc.logger.Info("permission assigned", "role", role.Name, "permission", perm.Name)
	return nil
}

func isDuplicateErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
