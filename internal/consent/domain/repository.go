package domain

import (
	"context"

	"github.com/google/uuid"
)

// ConsentRepository persists consent grant/revoke records.
type ConsentRepository interface {
	Create(ctx context.Context, consent *Consent) error
	GetByUserID(ctx context.Context, userID int64) (*Consent, error)
	Update(ctx context.Context, consent *Consent) error
	ListActive(ctx context.Context) ([]Consent, error)
}

// PseudonymousMappingRepository persists user ↔ pseudonymous_id links.
type PseudonymousMappingRepository interface {
	Create(ctx context.Context, mapping *PseudonymousMapping) error
	FindByPseudonymousID(ctx context.Context, pseudonymousID uuid.UUID) (*PseudonymousMapping, error)
	FindByUserID(ctx context.Context, userID int64) (*PseudonymousMapping, error)
	// FindPseudonymousIDsByUserIDs resolves a batch of internal user ids for operator reads.
	FindPseudonymousIDsByUserIDs(ctx context.Context, userIDs []int64) (map[int64]string, error)
}
