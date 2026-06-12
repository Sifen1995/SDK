package model

import "time"

type SubscriptionPlan struct {
	ID                  string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name                string    `gorm:"type:varchar(100);not null;unique"`
	MonthlyFeeETB       float64   `gorm:"type:numeric(12,2);not null"`
	MaxActiveCampaigns  int       `gorm:"not null;default:3"`
	MaxDailyBudgetETB   float64   `gorm:"type:numeric(12,2);not null"`
	IncludedImpressions int       `gorm:"not null;default:0"`
	SMSPlusEnabled      bool      `gorm:"not null;default:false"`
	AudiencemartEnabled bool      `gorm:"not null;default:false"`
	CPCDiscountPct      float64   `gorm:"type:numeric(5,2);default:0"`
	IsActive            bool      `gorm:"not null;default:true"`
	CreatedAt           time.Time `gorm:"not null;default:now()"`
}

func (SubscriptionPlan) TableName() string { return "subscription_plans" }
