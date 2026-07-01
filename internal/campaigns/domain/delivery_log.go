package domain

import "time"

// Delivery status values for campaign ad lifecycle events.
const (
	DeliveryDispatched = "DISPATCHED"
	DeliveryRendered   = "RENDERED"
	DeliveryClicked    = "CLICKED"
	DeliveryConverted  = "CONVERTED"
)

// DeliveryLog records campaign ad lifecycle events (table: campaign_delivery_logs).
type DeliveryLog struct {
	ID             string
	CampaignID     string
	UserID         string
	SessionID      string
	DeliveryStatus string
	LoggedAt       time.Time
}
