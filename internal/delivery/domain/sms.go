package domain

import (
	"context"
	"time"
)

const (
	SMSSendStatusQueued    = "queued"
	SMSSendStatusSent      = "sent"
	SMSSendStatusFailed    = "failed"
	SMSSendStatusDelivered = "delivered"
	SMSSendStatusClicked   = "clicked"
)

// DemoSMSRecipient stores demo-only phone routing for a seeded SDK user.
type DemoSMSRecipient struct {
	UserID             int64
	PhoneE164          string
	DisplayName        string
	IsActive           bool
	IsMock             bool
	ProviderExternalID string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// SMSSendAttempt tracks provider-native send state separate from billing_events.
type SMSSendAttempt struct {
	ID                string
	SendKey           string
	CampaignID        string
	PseudonymousID    string
	UserID            int64
	PhoneE164         string
	Provider          string
	ProviderMessageID string
	Status            string
	MessageBody       string
	TrackingToken     string
	DestinationURL    string
	ErrorMessage      string
	SentAt            *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time

	// Optional join fields for debug/list responses (not persisted on create).
	CampaignName string
	ImageURL     string
}

type DemoSMSRecipientRepository interface {
	FindActiveByPseudonymousID(ctx context.Context, pseudonymousID string) (*DemoSMSRecipient, error)
}

type SMSSendAttemptRepository interface {
	Create(ctx context.Context, attempt *SMSSendAttempt) error
	ExistsBySendKey(ctx context.Context, sendKey string) (bool, error)
	FindBySendKey(ctx context.Context, sendKey string) (*SMSSendAttempt, error)
	FindByTrackingToken(ctx context.Context, trackingToken string) (*SMSSendAttempt, error)
	UpdateStatus(ctx context.Context, attemptID, status, providerMessageID, errorMessage string) error
	ListRecent(ctx context.Context, limit int) ([]SMSSendAttempt, error)
}
