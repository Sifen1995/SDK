package model

import "time"

type Intent struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"type:uuid;not null" json:"user_id"`
	IntentName string    `gorm:"type:varchar(100);not null" json:"intent_name"`
	Confidence float64   `gorm:"type:numeric(4,3);not null" json:"confidence"`
	CreatedAt  time.Time `json:"created_at"`
}
