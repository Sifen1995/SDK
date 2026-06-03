package domain

import "context"

// EventRepository defines persistence contracts for the events domain.
type EventRepository interface {
	SaveBatch(ctx context.Context, events []Event) error
	FindByUser(ctx context.Context, userID string, limit int) ([]Event, error)
	FindByUserInternalID(ctx context.Context, internalUserID string, limit int) ([]Event, error)
	FindSessionEvents(ctx context.Context, sessionID string, limit int) ([]Event, error)
}
