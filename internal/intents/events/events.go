package events

const TopicIntentPredicted = "intent.predicted"

// IntentPredicted is published after a successful ML intent prediction.
type IntentPredicted struct {
	ExternalUserID  string
	Intent          string
	Confidence      float64
	TopSignals      []string
	RewardTriggered bool
}
