package domain

import (
	"time"

	"github.com/google/uuid"
)

// PseudonymousMapping links a reversible user row to the Flutter-side
// pseudonymous UUID used on predicted intents — without storing it on users.
type PseudonymousMapping struct {
	ID             string
	UserID         int64
	PseudonymousID uuid.UUID
	CreatedAt      time.Time
}
