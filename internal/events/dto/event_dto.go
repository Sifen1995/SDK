package dto

// EventResponseDTO is what the mobile app/SDK receives
type EventResponseDTO struct {
	EventID           string             `json:"event_id"`
	UserID            string             `json:"user_id"`
	Intent            string             `json:"intent,omitempty"`
	Confidence        float64            `json:"confidence,omitempty"`
	RewardTriggered   bool               `json:"reward_triggered"`
	Reward            *RewardResponseDTO `json:"reward,omitempty"`
	QueuedForAnalysis bool               `json:"queued_for_analysis"`
}

// RewardResponseDTO simplifies the reward data for the frontend
type RewardResponseDTO struct {
	RewardID   string  `json:"reward_id"`
	RewardType string  `json:"reward_type"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Message    string  `json:"message"`
}
