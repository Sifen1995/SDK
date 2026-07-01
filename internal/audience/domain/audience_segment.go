package domain

import "time"

// AudienceSegment is a purchasable Audiencemart cohort definition (rules, not user rows).
type AudienceSegment struct {
	ID               string
	Name             string
	Description      string
	TopIntentSignals []string
	ApproximateSize  int
	EstimatedCPM     float64
	AvailableFrom    time.Time
	AvailableUntil   *time.Time
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
