package model

import (
	"time"
)

type Event struct {
	ID         string                 `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EventID    string                 `gorm:"type:varchar(255);not null;uniqueIndex" json:"event_id"`
	UserID     string                 `gorm:"type:uuid;not null" json:"user_id"`
	EventType  string                 `gorm:"type:varchar(50);not null" json:"event_type" binding:"required,oneof=search product_view category_view add_to_cart remove_from_cart signup_complete checkout_started checkout_complete store_visit ad_impression ad_click app_open"`
	Metadata   map[string]interface{} `gorm:"type:jsonb;serializer:json;default:'{}'" json:"metadata"`
	IsFlagged  bool                   `gorm:"type:boolean;not null;default:false" json:"is_flagged"`
	FlagReason *string                `gorm:"type:varchar(255)" json:"flag_reason,omitempty"`
	Timestamp  time.Time              `gorm:"not null" json:"timestamp" binding:"required"`
	ReceivedAt time.Time              `gorm:"not null;default:now()" json:"received_at"`
}

func (Event) TableName() string {
	return "events"
}
