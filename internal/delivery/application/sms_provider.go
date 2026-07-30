package application

import "context"

type SMSMessage struct {
	To                string
	Body              string
	StatusCallbackURL string
}

type SMSSendResult struct {
	ProviderMessageID string
}

type SMSProvider interface {
	ProviderName() string
	Send(ctx context.Context, msg SMSMessage) (*SMSSendResult, error)
}
