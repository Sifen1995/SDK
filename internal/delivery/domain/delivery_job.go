package domain

import "time"

// DeliveryJob tracks async campaign delivery work for one pseudonymous identity.
type DeliveryJob struct {
	ID             string
	PseudonymousID string
	CampaignID     string
	CreatedAt      time.Time
}
