package domain

import "time"

// SegmentPurchase records an advertiser's paid entitlement to target a segment on a campaign.
type SegmentPurchase struct {
	ID           string
	AdvertiserID string
	SegmentID    string
	CampaignID   string
	AmountPaid   float64
	ValidFrom    time.Time
	ValidUntil   time.Time
	CreatedAt    time.Time
}
