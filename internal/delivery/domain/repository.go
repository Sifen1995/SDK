package domain

import (
	"context"

	"github.com/google/uuid"
)

// DeliveryRepository tracks per-user campaign delivery jobs.
type DeliveryRepository interface {
	WasDelivered(ctx context.Context, userID, campaignID uuid.UUID) (bool, error)
	CountToday(ctx context.Context, userID, campaignID uuid.UUID) (int, error)
}
