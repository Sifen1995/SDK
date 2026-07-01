package persistence

import (
	"time"

	"skykin-platform/internal/audience/domain"
)

// SegmentPurchaseRow is the GORM persistence model for segment_purchases.
type SegmentPurchaseRow struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID string    `gorm:"type:uuid;not null;index"`
	SegmentID    string    `gorm:"type:uuid;not null;index"`
	CampaignID   string    `gorm:"type:uuid;not null;index"`
	AmountPaid   float64   `gorm:"type:numeric(12,2);not null"`
	ValidFrom    time.Time `gorm:"not null"`
	ValidUntil   time.Time `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

func (SegmentPurchaseRow) TableName() string { return "segment_purchases" }

func (row *SegmentPurchaseRow) ToDomain() *domain.SegmentPurchase {
	if row == nil {
		return nil
	}
	return &domain.SegmentPurchase{
		ID:           row.ID,
		AdvertiserID: row.AdvertiserID,
		SegmentID:    row.SegmentID,
		CampaignID:   row.CampaignID,
		AmountPaid:   row.AmountPaid,
		ValidFrom:    row.ValidFrom,
		ValidUntil:   row.ValidUntil,
		CreatedAt:    row.CreatedAt,
	}
}

func SegmentPurchaseRowFromDomain(p *domain.SegmentPurchase) *SegmentPurchaseRow {
	if p == nil {
		return nil
	}
	return &SegmentPurchaseRow{
		ID:           p.ID,
		AdvertiserID: p.AdvertiserID,
		SegmentID:    p.SegmentID,
		CampaignID:   p.CampaignID,
		AmountPaid:   p.AmountPaid,
		ValidFrom:    p.ValidFrom,
		ValidUntil:   p.ValidUntil,
		CreatedAt:    p.CreatedAt,
	}
}
