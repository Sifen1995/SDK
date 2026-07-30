package events

// Segment candidate review is synchronous and transactional; it does not emit events.
const (
	TopicSubscriptionPlanCreated    = "admin.subscription_plan_created"
	TopicCampaignModerationPassed   = "admin.campaign_moderation_passed"
	TopicCampaignModerationRejected = "admin.campaign_moderation_rejected"
	TopicCampaignActivated          = "admin.campaign_activated"
)

// SubscriptionPlanCreatedEvent is published after a base subscription plan row is created.
type SubscriptionPlanCreatedEvent struct {
	PlanID string
}

// CampaignModerationPassedEvent is published when operator approves a pending campaign.
type CampaignModerationPassedEvent struct {
	CampaignID     string
	OperatorUserID string
	ChannelCode    string
}

// CampaignModerationRejectedEvent is published when operator rejects a pending campaign.
type CampaignModerationRejectedEvent struct {
	CampaignID     string
	OperatorUserID string
	Notes          string
}

// CampaignActivatedEvent is published when an approved campaign goes live.
type CampaignActivatedEvent struct {
	CampaignID     string
	OperatorUserID string
}
