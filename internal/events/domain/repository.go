package domain

import "context"

// EventRepository defines persistence contracts for the events domain.
type EventRepository interface {
	SaveBatch(ctx context.Context, events []Event) error
	FindByPseudonymousID(ctx context.Context, pseudonymousID string, limit int) ([]Event, error)
	FindSessionEvents(ctx context.Context, sessionID string, limit int) ([]Event, error)
}
