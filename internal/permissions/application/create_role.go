package application

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"skykin-platform/internal/permissions/domain"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateRoleUseCase struct {
	roleRepo interface {
		FindByName(ctx context.Context, name string) (*domain.Role, error)
		Create(ctx context.Context, role *domain.Role) error
	}
	bus interface {
		Publish(event messaging.Event)
	}
	logger *slog.Logger
}

func NewCreateRoleUseCase(
	roleRepo interface {
		FindByName(ctx context.Context, name string) (*domain.Role, error)
		Create(ctx context.Context, role *domain.Role) error
	},
	bus interface {
		Publish(event messaging.Event)
	},
	logger *slog.Logger,
) *CreateRoleUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &CreateRoleUseCase{roleRepo: roleRepo, bus: bus, logger: logger}
}

func (uc *CreateRoleUseCase) Execute(
	ctx context.Context,
	name string,
	description string,
	createdBy uuid.UUID,
) (*domain.Role, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return nil, errors.New("role name is required")
	}
	for _, reserved := range []string{"operator_admin", "advertiser", "analyst"} {
		if name == reserved {
			return nil, errors.New("reserved role name")
		}
	}
	existing, err := uc.roleRepo.FindByName(ctx, name)
	if err == nil && existing != nil {
		return nil, errors.New("role already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	role := &domain.Role{
		Name: name, Description: description, IsSystem: false, CreatedAt: time.Now().UTC(),
	}
	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	uc.bus.Publish(messaging.Event{
		Name: domain.EventRoleCreated,
		Payload: domain.RoleCreatedPayload{
			RoleID: role.ID, RoleName: role.Name, CreatedBy: createdBy,
		},
		Ctx: ctx,
	})
	uc.logger.Info("role created", "role", role.Name)
	return role, nil
}
