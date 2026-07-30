package domain

import (
	"context"

	"github.com/google/uuid"
)

// DeliveryRepository tracks campaign delivery jobs per pseudonymous identity.
type DeliveryRepository interface {
	WasDelivered(ctx context.Context, pseudonymousID string, campaignID uuid.UUID) (bool, error)
	CountToday(ctx context.Context, pseudonymousID string, campaignID uuid.UUID) (int, error)
	// RecordJob inserts a delivery_jobs row (no-op on duplicate pseudonymous id + campaign).
	RecordJob(ctx context.Context, pseudonymousID, campaignID string) error
}
