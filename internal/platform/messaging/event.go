package messaging

import "context"

// Event carries a named payload through the in-process bus.
type Event struct {
	Name    string
	Payload any
	Ctx     context.Context
}

