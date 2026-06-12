package model

import "time"

type AdvertiserSubscription struct {
	ID                 string           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID       string           `gorm:"type:uuid;not null;unique"`
	PlanID             string           `gorm:"type:uuid;not null"`
	Plan               SubscriptionPlan `gorm:"foreignKey:PlanID"`
	Status             string           `gorm:"type:varchar(20);not null;default:active"`
	CurrentPeriodStart time.Time        `gorm:"not null"`
	CurrentPeriodEnd   time.Time        `gorm:"not null"`
	ImpressionsUsed    int              `gorm:"not null;default:0"`
	CancelledAt        *time.Time
	CreatedAt          time.Time `gorm:"not null;default:now()"`
	UpdatedAt          time.Time `gorm:"not null;default:now()"`
}

func (AdvertiserSubscription) TableName() string { return "advertiser_subscriptions" }
