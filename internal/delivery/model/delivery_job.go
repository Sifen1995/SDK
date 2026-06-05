package model

import "time"

type DeliveryJob struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     string    `gorm:"type:uuid;not null;index"`
	CampaignID string    `gorm:"type:uuid;not null;index"`
	CreatedAt  time.Time `gorm:"not null;default:now()"`
}

func (DeliveryJob) TableName() string { return "delivery_jobs" }
