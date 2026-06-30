package events

const TopicIntentRewardEligible = "intent.reward_eligible"

// IntentRewardEligible is published when ML signals a reward should be created.
type IntentRewardEligible struct {
	ExternalUserID string
	InternalUserID string
	IntentID       string
	IntentName     string
}
