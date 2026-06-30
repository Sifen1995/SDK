package events

import "time"

const TopicCampaignCreated = "campaign.created"

// CampaignCreatedEvent is published when a new campaign row is persisted.
type CampaignCreatedEvent struct {
	CampaignID   string
	AdvertiserID string
	SegmentID    string
	AmountPaid   float64
	ValidFrom    time.Time
	ValidUntil   time.Time
	HasPurchase  bool
}


const TopicCampaignAdDelivered = "campaign.ad.delivered"

type CampaignAdDelivered struct {
	ExternalUserID string
	InternalUserID string
	SessionID      string
	Intent         string
	Ad             map[string]any
}
