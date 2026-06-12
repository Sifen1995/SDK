package model

import "time"

type Invoice struct {
	ID                 string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID       string     `gorm:"type:uuid;not null;index"`
	SubscriptionID     string     `gorm:"type:uuid;not null"`
	PeriodStart        time.Time  `gorm:"not null"`
	PeriodEnd          time.Time  `gorm:"not null"`
	SubscriptionFeeETB float64    `gorm:"type:numeric(12,2);not null"`
	UsageFeeETB        float64    `gorm:"type:numeric(12,2);not null"`
	TotalETB           float64    `gorm:"type:numeric(12,2);not null"`
	Status             string     `gorm:"type:varchar(20);not null;default:draft"`
	PaidAt             *time.Time
	PaymentRef         string     `gorm:"type:varchar(255)"`
	CreatedAt          time.Time  `gorm:"not null;default:now()"`
	UpdatedAt          time.Time  `gorm:"not null;default:now()"`
}

func (Invoice) TableName() string { return "invoices" }
