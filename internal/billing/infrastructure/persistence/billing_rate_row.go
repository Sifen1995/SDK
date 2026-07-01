package persistence

import (
	"time"

	"skykin-platform/internal/billing/domain"
)

// BillingRateRow is the GORM persistence model for billing_rates.
type BillingRateRow struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PlanID    string    `gorm:"type:uuid;not null;index"`
	EventType string    `gorm:"type:varchar(50);not null"`
	Model     string    `gorm:"type:varchar(10);not null"`
	RateETB   float64   `gorm:"type:numeric(10,4);not null"`
	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (BillingRateRow) TableName() string { return "billing_rates" }

func (row *BillingRateRow) ToDomain() *domain.BillingRate {
	if row == nil {
		return nil
	}
	return &domain.BillingRate{
		ID:        row.ID,
		PlanID:    row.PlanID,
		EventType: row.EventType,
		Model:     row.Model,
		RateETB:   row.RateETB,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
	}
}

func BillingRateRowFromDomain(r *domain.BillingRate) *BillingRateRow {
	if r == nil {
		return nil
	}
	return &BillingRateRow{
		ID:        r.ID,
		PlanID:    r.PlanID,
		EventType: r.EventType,
		Model:     r.Model,
		RateETB:   r.RateETB,
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt,
	}
}
