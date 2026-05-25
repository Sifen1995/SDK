package model

import (
	"time"
)

type Event struct {
	ID        string                 `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string                 `gorm:"type:uuid;not null" json:"user_id"`
	EventType string                 `gorm:"type:varchar(50);not null" json:"event_type" binding:"required,oneof=search product_view category_view add_to_cart remove_from_cart signup_complete checkout_started"`
	Metadata  map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"metadata"`
	Timestamp time.Time              `gorm:"not null" json:"timestamp" binding:"required"`
	CreatedAt time.Time              `gorm:"default:now()" json:"created_at"`
}

func (Event) TableName() string {
	return "events"
}
