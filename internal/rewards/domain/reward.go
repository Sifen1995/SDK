package domain

import "time"

// Reward is a granted incentive tied to an intent prediction.
type Reward struct {
	ID             string
	PseudonymousID string
	IntentID       string
	RuleID         string
	RewardType     string
	Amount         float64
	Currency       string
	Status         string
	Message        string
	CreatedAt      time.Time
	SentAt         *time.Time
	ClaimedAt      *time.Time
}

// RewardRule maps an intent name to a reward template.
type RewardRule struct {
	ID         string
	IntentName string
	RewardType string
	Amount     float64
	Currency   string
	Message    string
	IsActive   bool
	CreatedAt  time.Time
}
