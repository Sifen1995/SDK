package domain

import (
	"context"
	"time"
)

// IntentRepository persists and queries intent predictions.
type IntentRepository interface {
	Create(ctx context.Context, intent *Intent) (*Intent, error)
	FindUsersWithIntent(ctx context.Context, intentName string, minConfidence float64, since time.Time) ([]string, error)
	FindUsersWithAnyIntent(ctx context.Context, intentNames []string, minConfidence float64, since time.Time) ([]string, error)
	FindLatestByUserIDs(ctx context.Context, userIDs []string) (map[string]*Intent, error)
}
