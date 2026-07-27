package domain

import (
	"context"
	"time"
)

// AnonymousUserID is a fixed sentinel UUID for non-consented delivery log rows
// (campaign_delivery_logs.user_id is NOT NULL).
const AnonymousUserID = "00000000-0000-0000-0000-000000000000"

// Delivery lifecycle statuses written to campaign_delivery_logs.
const (
	StatusDispatched = "DISPATCHED"
	StatusRendered   = "RENDERED"
	StatusClicked    = "CLICKED"
	StatusConverted  = "CONVERTED"
)

// DeliveryLog is a campaign ad lifecycle event owned by the delivery module.
type DeliveryLog struct {
	ID             string
	CampaignID     string
	UserID         string
	SessionID      string
	DeliveryStatus string
	LoggedAt       time.Time
}

// DeliveryLogRepository persists campaign_delivery_logs rows.
type DeliveryLogRepository interface {
	CreateBatch(ctx context.Context, logs []DeliveryLog) error
}
