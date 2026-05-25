package model

import "time"

type Reward struct {
	ID         string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string     `gorm:"type:uuid;not null" json:"user_id"`
	IntentID   string     `gorm:"type:uuid;not null" json:"intent_id"`
	RuleID     string     `gorm:"type:uuid;not null" json:"rule_id"`
	RewardType string     `gorm:"type:varchar(50);not null" json:"reward_type"`
	Amount     float64    `gorm:"type:numeric(10,2);not null" json:"amount"`
	Currency   string     `gorm:"type:varchar(50);not null" json:"currency"`
	Status     string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Message    string     `gorm:"type:text;not null" json:"message"`
	CreatedAt  time.Time  `json:"created_at"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
	ClaimedAt  *time.Time `json:"claimed_at,omitempty"`
}
