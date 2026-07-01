package persistence

import (
	"time"

	"skykin-platform/internal/campaigns/domain"
)

// DeliveryLogRow is the GORM persistence model for campaign_delivery_logs.
type DeliveryLogRow struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CampaignID     string    `gorm:"type:uuid;not null;index"`
	UserID         string    `gorm:"type:uuid;not null"`
	SessionID      string    `gorm:"type:varchar(255);not null"`
	DeliveryStatus string    `gorm:"type:varchar(50);not null"`
	LoggedAt       time.Time `gorm:"not null;default:now()"`
}

func (DeliveryLogRow) TableName() string { return "campaign_delivery_logs" }

func (row *DeliveryLogRow) ToDomain() *domain.DeliveryLog {
	if row == nil {
		return nil
	}
	return &domain.DeliveryLog{
		ID:             row.ID,
		CampaignID:     row.CampaignID,
		UserID:         row.UserID,
		SessionID:      row.SessionID,
		DeliveryStatus: row.DeliveryStatus,
		LoggedAt:       row.LoggedAt,
	}
}

func DeliveryLogRowFromDomain(log *domain.DeliveryLog) *DeliveryLogRow {
	if log == nil {
		return nil
	}
	return &DeliveryLogRow{
		ID:             log.ID,
		CampaignID:     log.CampaignID,
		UserID:         log.UserID,
		SessionID:      log.SessionID,
		DeliveryStatus: log.DeliveryStatus,
		LoggedAt:       log.LoggedAt,
	}
}
