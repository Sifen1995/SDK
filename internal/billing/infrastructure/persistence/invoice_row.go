package persistence

import (
	"time"

	"skykin-platform/internal/billing/domain"
)

// InvoiceRow is the GORM persistence model for invoices.
type InvoiceRow struct {
	ID                 string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID       string     `gorm:"type:uuid;not null;index"`
	SubscriptionID     string     `gorm:"type:uuid;not null"`
	PeriodStart        time.Time  `gorm:"not null"`
	PeriodEnd          time.Time  `gorm:"not null"`
	SubscriptionFeeETB float64    `gorm:"type:numeric(12,2);not null"`
	UsageFeeETB        float64    `gorm:"type:numeric(12,2);not null"`
	TotalETB           float64    `gorm:"type:numeric(12,2);not null"`
	Status             string     `gorm:"type:varchar(20);not null;default:draft"`
	PaidAt             *time.Time
	PaymentRef         string    `gorm:"type:varchar(255)"`
	CreatedAt          time.Time `gorm:"not null;default:now()"`
	UpdatedAt          time.Time `gorm:"not null;default:now()"`
}

func (InvoiceRow) TableName() string { return "invoices" }

func (row *InvoiceRow) ToDomain() *domain.Invoice {
	if row == nil {
		return nil
	}
	return &domain.Invoice{
		ID:                 row.ID,
		AdvertiserID:       row.AdvertiserID,
		SubscriptionID:     row.SubscriptionID,
		PeriodStart:        row.PeriodStart,
		PeriodEnd:          row.PeriodEnd,
		SubscriptionFeeETB: row.SubscriptionFeeETB,
		UsageFeeETB:        row.UsageFeeETB,
		TotalETB:           row.TotalETB,
		Status:             row.Status,
		PaidAt:             row.PaidAt,
		PaymentRef:         row.PaymentRef,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func InvoiceRowFromDomain(inv *domain.Invoice) *InvoiceRow {
	if inv == nil {
		return nil
	}
	return &InvoiceRow{
		ID:                 inv.ID,
		AdvertiserID:       inv.AdvertiserID,
		SubscriptionID:     inv.SubscriptionID,
		PeriodStart:        inv.PeriodStart,
		PeriodEnd:          inv.PeriodEnd,
		SubscriptionFeeETB: inv.SubscriptionFeeETB,
		UsageFeeETB:        inv.UsageFeeETB,
		TotalETB:           inv.TotalETB,
		Status:             inv.Status,
		PaidAt:             inv.PaidAt,
		PaymentRef:         inv.PaymentRef,
		CreatedAt:          inv.CreatedAt,
		UpdatedAt:          inv.UpdatedAt,
	}
}
