package domain

import (
	"fmt"
	"time"

	"skykin-platform/internal/consent/validation"
)

// Consent records whether an SDK user allows linking predicted intent to a
// reversible identity (via pseudonymous_mappings → users).
type Consent struct {
	ID           string
	UserID       int64
	ConsentLevel string
	IsActive     bool
	GrantedAt    *time.Time
	RevokedAt    *time.Time
	SDKVersion   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ValidateConsentLevel returns an error when ConsentLevel is not allowed.
func (c *Consent) ValidateConsentLevel() error {
	if validation.ValidateConsentLevel(c.ConsentLevel) {
		return nil
	}
	return fmt.Errorf("invalid consent level: %s", c.ConsentLevel)
}
