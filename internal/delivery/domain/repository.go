package domain

import (
	"context"

	"github.com/google/uuid"
)

// DeliveryRepository tracks per-user campaign delivery jobs.
type DeliveryRepository interface {
	WasDelivered(ctx context.Context, userID string, campaignID uuid.UUID) (bool, error)
	CountToday(ctx context.Context, userID string, campaignID uuid.UUID) (int, error)
	// RecordJob inserts a delivery_jobs row (no-op on duplicate user+campaign).
	RecordJob(ctx context.Context, userID, campaignID string) error
}
