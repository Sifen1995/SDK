package bootstrap

import (
	"context"
	"fmt"

	consentInfra "skykin-platform/internal/consent/infrastructure"
	"skykin-platform/internal/users/domain"
	usersInfra "skykin-platform/internal/users/infrastructure"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PseudonymousUserResolver resolves Flutter pseudonymous UUIDs to internal users
// via consent.pseudonymous_mappings. Composition-root adapter (BCs stay decoupled).
type PseudonymousUserResolver struct {
	users    domain.UserRepository
	mappings *consentInfra.PseudonymousMappingRepository
}

func NewPseudonymousUserResolver(db *gorm.DB) *PseudonymousUserResolver {
	return &PseudonymousUserResolver{
		users:    usersInfra.NewUserRepository(db),
		mappings: consentInfra.NewPseudonymousMappingRepository(db),
	}
}

// FindOrCreate looks up an existing consented user by pseudonymous_id.
// Users are only created through the consent registration saga — never here.
func (r *PseudonymousUserResolver) FindOrCreate(ctx context.Context, pseudoID string) (*domain.User, error) {
	id, err := uuid.Parse(pseudoID)
	if err != nil {
		return nil, fmt.Errorf("invalid pseudonymous_id: %w", err)
	}
	mapping, err := r.mappings.FindByPseudonymousID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("mapping lookup failed: %w", err)
	}
	return r.users.FindByID(ctx, mapping.UserID)
}
