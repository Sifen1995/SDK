package application

import (
	"context"

	"skykin-platform/internal/users/domain"
)

// DedupStore prevents duplicate event ingestion within a TTL window.
type DedupStore interface {
	TryAcquire(ctx context.Context, eventID string) (acquired bool, err error)
}

// EventPublisher publishes domain events to the messaging bus.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any)
}

// UserResolver resolves Flutter pseudonymous IDs to internal users (via consent mapping).
type UserResolver interface {
	FindOrCreate(ctx context.Context, pseudonymousID string) (*domain.User, error)
}
