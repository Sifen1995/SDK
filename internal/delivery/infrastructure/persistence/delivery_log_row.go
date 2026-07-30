package persistence

import (
	"time"

	"skykin-platform/internal/delivery/domain"
)

// DeliveryLogRow maps to campaign_delivery_logs (shared table; written by delivery module).
type DeliveryLogRow struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CampaignID     string    `gorm:"type:uuid;not null;index"`
	PseudonymousID string    `gorm:"column:pseudonymous_id;type:uuid;not null"`
	SessionID      string    `gorm:"type:varchar(255);not null"`
	DeliveryStatus string    `gorm:"type:varchar(50);not null"`
	LoggedAt       time.Time `gorm:"not null;default:now()"`
}

func (DeliveryLogRow) TableName() string { return "campaign_delivery_logs" }

func DeliveryLogRowFromDomain(log *domain.DeliveryLog) *DeliveryLogRow {
	if log == nil {
		return nil
	}
	return &DeliveryLogRow{
		ID:             log.ID,
		CampaignID:     log.CampaignID,
		PseudonymousID: log.PseudonymousID,
		SessionID:      log.SessionID,
		DeliveryStatus: log.DeliveryStatus,
		LoggedAt:       log.LoggedAt,
	}
}
