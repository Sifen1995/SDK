package messaging

import (
	"context"
	"sync"
)

type Handler func(Event)

// Bus is a lightweight pub/sub for internal domain events.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string][]Handler),
	}
}

func (b *Bus) Subscribe(eventName string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventName] = append(b.subscribers[eventName], h)
}

// Publish delivers the event to subscribers asynchronously.
func (b *Bus) Publish(e Event) {
	if e.Ctx != nil {
		e.Ctx = context.WithoutCancel(e.Ctx)
	}

	b.mu.RLock()
	handlers := append([]Handler(nil), b.subscribers[e.Name]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		go h(e)
	}
}

// PublishSync delivers the event to subscribers on the calling goroutine
// in subscription order. Used for request-path sagas that must finish
// before an HTTP response is returned.
func (b *Bus) PublishSync(e Event) {
	if e.Ctx != nil {
		e.Ctx = context.WithoutCancel(e.Ctx)
	}

	b.mu.RLock()
	handlers := append([]Handler(nil), b.subscribers[e.Name]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		h(e)
	}
}
