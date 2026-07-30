package application

import (
	"context"
)

// DedupStore prevents duplicate event ingestion within a TTL window.
type DedupStore interface {
	TryAcquire(ctx context.Context, eventID string) (acquired bool, err error)
}

// EventPublisher publishes domain events to the messaging bus.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any)
}

// ConsentGate verifies a pseudonymous id belongs to a consented user. Events are
// stored against the pseudonymous id only; the internal user id never leaves consent.
type ConsentGate interface {
	EnsureConsented(ctx context.Context, pseudonymousID string) error
}
