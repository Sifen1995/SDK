package domain

const (
	// EventUserProvisionedForConsent is published after a users row is created
	// in response to ConsentRegistrationRequested.
	EventUserProvisionedForConsent = "UserProvisionedForConsent"
)

// UserProvisionedForConsentPayload continues the consent registration saga.
type UserProvisionedForConsentPayload struct {
	UserID         int64
	PseudonymousID string
	ConsentLevel   string
	SMSConsented   bool
	SDKVersion     string
}
