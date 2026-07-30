package domain

import "time"

// Intent is a persisted ML prediction row, keyed by the pseudonymous UUID the SDK sends.
type Intent struct {
	ID             string
	PseudonymousID string
	IntentName     string
	Confidence     float64
	CreatedAt      time.Time
}

// IntentProfile is the SDK payload after on-device ML (accessibility + app usage).
type IntentProfile struct {
	PseudonymousID string
	IntentName     string
	Confidence     float64
	ModelVersion   string
	RecordedAt     time.Time
	ExpiresAt      time.Time
}
