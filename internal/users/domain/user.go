package domain

import "time"

// User is an SDK end-user resolved from external identifiers.
type User struct {
	ID             string
	ExternalUserID string
	PhoneNumber    string
	CreatedAt      time.Time
}
