package messaging

import "sync"

type Handler func(Event)

// Bus is a lightweight async pub/sub for internal domain events.
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

func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.subscribers[e.Name]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		go h(e)
	}
}

