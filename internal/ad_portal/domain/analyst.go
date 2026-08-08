package domain

import "time"

// Analyst is the profile row for read_only_analyst portal users.
type Analyst struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
