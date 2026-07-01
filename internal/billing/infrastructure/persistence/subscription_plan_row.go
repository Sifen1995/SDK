package persistence

import (
	"time"

	"skykin-platform/internal/billing/domain"
)

// SubscriptionPlanRow is the GORM persistence model for subscription_plans.
type SubscriptionPlanRow struct {
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

func (SubscriptionPlanRow) TableName() string { return "subscription_plans" }

func (row *SubscriptionPlanRow) ToDomain() *domain.SubscriptionPlan {
	if row == nil {
		return nil
	}
	return &domain.SubscriptionPlan{
		ID:                  row.ID,
		Name:                row.Name,
		MonthlyFeeETB:       row.MonthlyFeeETB,
		MaxActiveCampaigns:  row.MaxActiveCampaigns,
		MaxDailyBudgetETB:   row.MaxDailyBudgetETB,
		IncludedImpressions: row.IncludedImpressions,
		SMSPlusEnabled:      row.SMSPlusEnabled,
		AudiencemartEnabled: row.AudiencemartEnabled,
		CPCDiscountPct:      row.CPCDiscountPct,
		IsActive:            row.IsActive,
		CreatedAt:           row.CreatedAt,
	}
}

func SubscriptionPlanRowFromDomain(p *domain.SubscriptionPlan) *SubscriptionPlanRow {
	if p == nil {
		return nil
	}
	return &SubscriptionPlanRow{
		ID:                  p.ID,
		Name:                p.Name,
		MonthlyFeeETB:       p.MonthlyFeeETB,
		MaxActiveCampaigns:  p.MaxActiveCampaigns,
		MaxDailyBudgetETB:   p.MaxDailyBudgetETB,
		IncludedImpressions: p.IncludedImpressions,
		SMSPlusEnabled:      p.SMSPlusEnabled,
		AudiencemartEnabled: p.AudiencemartEnabled,
		CPCDiscountPct:      p.CPCDiscountPct,
		IsActive:            p.IsActive,
		CreatedAt:           p.CreatedAt,
	}
}
