package persistence

import (
	"time"

	"skykin-platform/internal/billing/domain"
)

// BillingEventRow is the GORM persistence model for billing_events.
type BillingEventRow struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID     string    `gorm:"type:uuid;not null;index"`
	CampaignID       string    `gorm:"type:uuid;not null;index"`
	SubscriptionID   string    `gorm:"type:uuid;not null"`
	EventType        string    `gorm:"type:varchar(50);not null"`
	BillingModel     string    `gorm:"type:varchar(20);not null"`
	RateApplied      float64   `gorm:"type:numeric(10,4);not null"`
	TransactionValue float64   `gorm:"type:numeric(12,2);default:0"`
	ChargeETB        float64   `gorm:"type:numeric(10,4);not null"`
	IsBilled         bool      `gorm:"not null;default:false;index"`
	OccurredAt       time.Time `gorm:"not null;index"`
	CreatedAt        time.Time `gorm:"not null;default:now()"`
}

func (BillingEventRow) TableName() string { return "billing_events" }

func (row *BillingEventRow) ToDomain() *domain.BillingEvent {
	if row == nil {
		return nil
	}
	return &domain.BillingEvent{
		ID:               row.ID,
		AdvertiserID:     row.AdvertiserID,
		CampaignID:       row.CampaignID,
		SubscriptionID:   row.SubscriptionID,
		EventType:        row.EventType,
		BillingModel:     row.BillingModel,
		RateApplied:      row.RateApplied,
		TransactionValue: row.TransactionValue,
		ChargeETB:        row.ChargeETB,
		IsBilled:         row.IsBilled,
		OccurredAt:       row.OccurredAt,
		CreatedAt:        row.CreatedAt,
	}
}

func BillingEventRowFromDomain(e *domain.BillingEvent) *BillingEventRow {
	if e == nil {
		return nil
	}
	return &BillingEventRow{
		ID:               e.ID,
		AdvertiserID:     e.AdvertiserID,
		CampaignID:       e.CampaignID,
		SubscriptionID:   e.SubscriptionID,
		EventType:        e.EventType,
		BillingModel:     e.BillingModel,
		RateApplied:      e.RateApplied,
		TransactionValue: e.TransactionValue,
		ChargeETB:        e.ChargeETB,
		IsBilled:         e.IsBilled,
		OccurredAt:       e.OccurredAt,
		CreatedAt:        e.CreatedAt,
	}
}
