package infrastructure

import (
	"context"

	"skykin-platform/internal/events/application"
	"skykin-platform/internal/platform/messaging"
)

// BusEventPublisher publishes internal events through the in-process bus.
// Architecture is ready to swap for Kafka/NATS behind the same application port.
type BusEventPublisher struct {
	bus *messaging.Bus
}

func NewBusEventPublisher(bus *messaging.Bus) application.EventPublisher {
	return &BusEventPublisher{bus: bus}
}

func (p *BusEventPublisher) Publish(ctx context.Context, name string, payload any) {
	p.bus.Publish(messaging.Event{Name: name, Payload: payload, Ctx: ctx})
}
