package model

import "time"

// DeliveryLog records campaign ad lifecycle events (table: campaign_delivery_logs).
type DeliveryLog struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CampaignID     string    `gorm:"type:uuid;not null;index"`
	UserID         string    `gorm:"type:uuid;not null"`
	SessionID      string    `gorm:"type:varchar(255);not null"`
	DeliveryStatus string    `gorm:"type:varchar(50);not null"`
	LoggedAt       time.Time `gorm:"not null;default:now()"`
}

func (DeliveryLog) TableName() string { return "campaign_delivery_logs" }

const (
	DeliveryDispatched = "DISPATCHED"
	DeliveryRendered   = "RENDERED"
	DeliveryClicked    = "CLICKED"
	DeliveryConverted  = "CONVERTED"
)
