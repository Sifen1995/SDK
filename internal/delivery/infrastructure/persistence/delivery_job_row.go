package persistence

import (
	"time"

	"skykin-platform/internal/delivery/domain"
)

// DeliveryJobRow is the GORM persistence model for delivery_jobs.
type DeliveryJobRow struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PseudonymousID string    `gorm:"column:pseudonymous_id;type:uuid;not null;index"`
	CampaignID     string    `gorm:"type:uuid;not null;index"`
	CreatedAt      time.Time `gorm:"not null;default:now()"`
}

func (DeliveryJobRow) TableName() string { return "delivery_jobs" }

func (row *DeliveryJobRow) ToDomain() *domain.DeliveryJob {
	if row == nil {
		return nil
	}
	return &domain.DeliveryJob{
		ID:             row.ID,
		PseudonymousID: row.PseudonymousID,
		CampaignID:     row.CampaignID,
		CreatedAt:      row.CreatedAt,
	}
}

func DeliveryJobRowFromDomain(job *domain.DeliveryJob) *DeliveryJobRow {
	if job == nil {
		return nil
	}
	return &DeliveryJobRow{
		ID:             job.ID,
		PseudonymousID: job.PseudonymousID,
		CampaignID:     job.CampaignID,
		CreatedAt:      job.CreatedAt,
	}
}
