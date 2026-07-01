package domain

import "time"

// Channel is a delivery channel (IN_APP_BANNER, PUSH, SMS_PLUS, …).
type Channel struct {
	ID          string
	Code        string
	Name        string
	Description string
	IsPremium   bool
	IsActive    bool
	CreatedAt   time.Time
}
