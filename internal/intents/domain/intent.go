package domain

import "time"

// Intent is a persisted ML prediction for one user.
type Intent struct {
	ID         string
	UserID     string
	IntentName string
	Confidence float64
	CreatedAt  time.Time
}
