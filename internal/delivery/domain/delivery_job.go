package domain

import "time"

// DeliveryJob tracks async campaign delivery work for a user.
type DeliveryJob struct {
	ID         string
	UserID     string
	CampaignID string
	CreatedAt  time.Time
}
