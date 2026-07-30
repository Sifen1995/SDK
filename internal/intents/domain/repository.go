package domain

import (
	"context"
	"time"
)

// ConsistentSignal is a neutral projection of sustained intent activity.
// Analytics maps this into its own finding DTO at the composition root.
type ConsistentSignal struct {
	PseudonymousID string
	Confidence     float64
	DaysActive     int
	LastSeenAt     time.Time
}

// IntentRepository persists and queries intent predictions by pseudonymous id.
type IntentRepository interface {
	Create(ctx context.Context, intent *Intent) (*Intent, error)
	CreateBatch(ctx context.Context, intents []*Intent) error
	FindPseudonymousIDsWithIntent(ctx context.Context, intentName string, minConfidence float64, since time.Time) ([]string, error)
	FindPseudonymousIDsWithAnyIntent(ctx context.Context, intentNames []string, minConfidence float64, since time.Time) ([]string, error)
	FindLatestByPseudonymousIDs(ctx context.Context, pseudonymousIDs []string) (map[string]*Intent, error)
	FindConsistentSignals(ctx context.Context, intentName string, minConf float64, lookbackDays, minDays, maxAgeDays int) ([]*ConsistentSignal, error)
}
