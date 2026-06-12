package domain

import (
	"context"
	"time"

	"skykin-platform/internal/intents/model"

	"github.com/google/uuid"
)

// IntentRepository persists and queries intent predictions.
type IntentRepository interface {
	Create(ctx context.Context, intent *model.Intent) (*model.Intent, error)
	FindUsersWithIntent(ctx context.Context, intentName string, minConfidence float64, since time.Time) ([]uuid.UUID, error)
	// FindUsersWithAnyIntent returns users whose latest prediction matches any intent in the list.
	FindUsersWithAnyIntent(ctx context.Context, intentNames []string, minConfidence float64, since time.Time) ([]uuid.UUID, error)
}
