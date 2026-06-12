package model

import "time"

type BillingEvent struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID     string    `gorm:"type:uuid;not null;index"`
	CampaignID       string    `gorm:"type:uuid;not null;index"`
	SubscriptionID   string    `gorm:"type:uuid;not null"`
	EventType        string    `gorm:"type:varchar(50);not null"`
	BillingModel     string    `gorm:"type:varchar(20);not null"`
	RateApplied      float64   `gorm:"type:numeric(10,4);not null"`
	TransactionValue float64   `gorm:"type:numeric(12,2);default:0"`
	ChargeETB        float64   `gorm:"type:numeric(10,4);not null"`
	IsBilled         bool      `gorm:"not null;default:false;index"`
	OccurredAt       time.Time `gorm:"not null;index"`
	CreatedAt        time.Time `gorm:"not null;default:now()"`
}

func (BillingEvent) TableName() string { return "billing_events" }
