package persistence

import (
	"time"

	"skykin-platform/internal/billing/domain"
)

// AdvertiserSubscriptionRow is the GORM persistence model for advertiser_subscriptions.
type AdvertiserSubscriptionRow struct {
	ID                 string                `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID       string                `gorm:"type:uuid;not null;unique"`
	PlanID             string                `gorm:"type:uuid;not null"`
	Plan               SubscriptionPlanRow   `gorm:"foreignKey:PlanID"`
	Status             string                `gorm:"type:varchar(20);not null;default:active"`
	CurrentPeriodStart time.Time             `gorm:"not null"`
	CurrentPeriodEnd   time.Time             `gorm:"not null"`
	ImpressionsUsed    int                   `gorm:"not null;default:0"`
	CancelledAt        *time.Time
	CreatedAt          time.Time `gorm:"not null;default:now()"`
	UpdatedAt          time.Time `gorm:"not null;default:now()"`
}

func (AdvertiserSubscriptionRow) TableName() string { return "advertiser_subscriptions" }

func (row *AdvertiserSubscriptionRow) ToDomain() *domain.AdvertiserSubscription {
	if row == nil {
		return nil
	}
	sub := &domain.AdvertiserSubscription{
		ID:                 row.ID,
		AdvertiserID:       row.AdvertiserID,
		PlanID:             row.PlanID,
		Status:             row.Status,
		CurrentPeriodStart: row.CurrentPeriodStart,
		CurrentPeriodEnd:   row.CurrentPeriodEnd,
		ImpressionsUsed:    row.ImpressionsUsed,
		CancelledAt:        row.CancelledAt,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
	if row.Plan.ID != "" {
		sub.Plan = *row.Plan.ToDomain()
	}
	return sub
}

func AdvertiserSubscriptionRowFromDomain(sub *domain.AdvertiserSubscription) *AdvertiserSubscriptionRow {
	if sub == nil {
		return nil
	}
	return &AdvertiserSubscriptionRow{
		ID:                 sub.ID,
		AdvertiserID:       sub.AdvertiserID,
		PlanID:             sub.PlanID,
		Status:             sub.Status,
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		ImpressionsUsed:    sub.ImpressionsUsed,
		CancelledAt:        sub.CancelledAt,
		CreatedAt:          sub.CreatedAt,
		UpdatedAt:          sub.UpdatedAt,
	}
}
