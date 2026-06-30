package events

import "github.com/google/uuid"

const (
	TopicCandidateApproved          = "admin.candidate_approved"
	TopicCandidateRejected          = "admin.candidate_rejected"
	TopicSubscriptionPlanCreated    = "admin.subscription_plan_created"
	TopicCampaignModerationPassed   = "admin.campaign_moderation_passed"
	TopicCampaignModerationRejected = "admin.campaign_moderation_rejected"
	TopicCampaignActivated          = "admin.campaign_activated"
)

// CandidateApprovedEvent is published when an operator approves a segment candidate.
type CandidateApprovedEvent struct {
	CandidateID  uuid.UUID
	AdminID      uuid.UUID
	Name         string
	Description  string
	IntentName   string
	UserCount    int
	EstimatedCPM float64
}

// CandidateRejectedEvent is published when an operator rejects a segment candidate.
type CandidateRejectedEvent struct {
	CandidateID uuid.UUID
	AdminID     uuid.UUID
	Notes       string
}

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
