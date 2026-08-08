package persistence

import (
	"time"

	"skykin-platform/internal/ad_portal/domain"
)

type AnalystRow struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DisplayName string    `gorm:"column:display_name;type:varchar(255);not null"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}

func (AnalystRow) TableName() string { return "analysts" }

func (row *AnalystRow) ToDomain() *domain.Analyst {
	if row == nil {
		return nil
	}
	return &domain.Analyst{
		ID:          row.ID,
		DisplayName: row.DisplayName,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func AnalystRowFromDomain(a *domain.Analyst) *AnalystRow {
	if a == nil {
		return nil
	}
	return &AnalystRow{
		ID:          a.ID,
		DisplayName: a.DisplayName,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
