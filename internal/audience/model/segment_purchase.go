package model

import "time"

// SegmentPurchase records an advertiser's paid entitlement to target a segment on a campaign.
type SegmentPurchase struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID string    `gorm:"type:uuid;not null;index"`
	SegmentID    string    `gorm:"type:uuid;not null;index"`
	CampaignID   string    `gorm:"type:uuid;not null;index"`
	AmountPaid   float64   `gorm:"type:numeric(12,2);not null"`
	ValidFrom    time.Time `gorm:"not null"`
	ValidUntil   time.Time `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

func (SegmentPurchase) TableName() string { return "segment_purchases" }
