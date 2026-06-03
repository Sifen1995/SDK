package events

const TopicCampaignAdDelivered = "campaign.ad.delivered"

type CampaignAdDelivered struct {
	ExternalUserID string
	InternalUserID string
	SessionID      string
	Intent         string
	Ad             map[string]any
}
