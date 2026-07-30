package bootstrap

import (
	"context"
	"fmt"

	consentInfra "skykin-platform/internal/consent/infrastructure"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PseudonymousConsentGate checks that a pseudonymous id was issued by the consent
// registration saga. Composition-root adapter so the events module never depends
// on consent or users directly.
type PseudonymousConsentGate struct {
	mappings *consentInfra.PseudonymousMappingRepository
}

func NewPseudonymousConsentGate(db *gorm.DB) *PseudonymousConsentGate {
	return &PseudonymousConsentGate{mappings: consentInfra.NewPseudonymousMappingRepository(db)}
}

// EnsureConsented rejects ids that are not valid UUIDs or have no consent mapping.
func (g *PseudonymousConsentGate) EnsureConsented(ctx context.Context, pseudonymousID string) error {
	id, err := uuid.Parse(pseudonymousID)
	if err != nil {
		return fmt.Errorf("invalid pseudonymous_id: %w", err)
	}
	if _, err := g.mappings.FindByPseudonymousID(ctx, id); err != nil {
		return fmt.Errorf("pseudonymous_id is not registered for consent: %w", err)
	}
	return nil
}
