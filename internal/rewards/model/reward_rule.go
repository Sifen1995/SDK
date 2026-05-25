package model

import (
	"time"
)

type RewardRule struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	IntentName string    `gorm:"type:varchar(100);unique;not null" json:"intent_name"`
	RewardType string    `gorm:"type:varchar(50);not null" json:"reward_type"`
	Amount     float64   `gorm:"type:numeric(10,2);not null" json:"amount"`
	Currency   string    `gorm:"type:varchar(50);not null" json:"currency"`
	Message    string    `gorm:"type:text;not null" json:"message"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}
