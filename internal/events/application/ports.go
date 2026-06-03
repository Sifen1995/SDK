package application

import (
	"context"

	"skykin-platform/internal/users/model"
)

// DedupStore prevents duplicate event ingestion within a TTL window.
type DedupStore interface {
	TryAcquire(ctx context.Context, eventID string) (acquired bool, err error)
}

// EventPublisher publishes domain events to the messaging bus.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any)
}

// UserResolver resolves SDK external user IDs to internal UUIDs.
type UserResolver interface {
	FindOrCreate(ctx context.Context, externalUserID string) (*model.Users, error)
}
