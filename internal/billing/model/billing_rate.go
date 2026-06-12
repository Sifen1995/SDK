package model

import "time"

type BillingRate struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PlanID    string    `gorm:"type:uuid;not null;index"`
	EventType string    `gorm:"type:varchar(50);not null"`
	Model     string    `gorm:"type:varchar(10);not null"`
	RateETB   float64   `gorm:"type:numeric(10,4);not null"`
	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (BillingRate) TableName() string { return "billing_rates" }
