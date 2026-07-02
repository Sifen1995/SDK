package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// IntentRepository persists and queries intent predictions.
type IntentRepository interface {
	Create(ctx context.Context, intent *Intent) (*Intent, error)
	FindUsersWithIntent(ctx context.Context, intentName string, minConfidence float64, since time.Time) ([]uuid.UUID, error)
	FindUsersWithAnyIntent(ctx context.Context, intentNames []string, minConfidence float64, since time.Time) ([]uuid.UUID, error)
	FindLatestByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*Intent, error)
}
