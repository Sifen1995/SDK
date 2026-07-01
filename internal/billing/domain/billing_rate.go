package domain

import "time"

// BillingRate is a per-plan usage rate for a billing event type and model.
type BillingRate struct {
	ID        string
	PlanID    string
	EventType string
	Model     string
	RateETB   float64
	IsActive  bool
	CreatedAt time.Time
}
