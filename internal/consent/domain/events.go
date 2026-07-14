package domain

const (
	// EventConsentRegistrationRequested starts the cross-module consent saga.
	EventConsentRegistrationRequested = "ConsentRegistrationRequested"

	// EventConsentCreated is published after consent + mapping are persisted.
	EventConsentCreated = "ConsentCreated"
)

// ConsentRegistrationRequestedPayload asks the users module to provision a user.
type ConsentRegistrationRequestedPayload struct {
	PseudonymousID string
	ConsentLevel   string
	SDKVersion     string
}

// ConsentCreatedPayload is published after a successful consent registration.
type ConsentCreatedPayload struct {
	ConsentID      string
	UserID         int64
	PseudonymousID string
	ConsentLevel   string
	SDKVersion     string
}
