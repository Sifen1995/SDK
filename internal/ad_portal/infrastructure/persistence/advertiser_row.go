package persistence

import (
	"time"

	"skykin-platform/internal/ad_portal/domain"
)

type AdvertiserRow struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyName string    `gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}

func (AdvertiserRow) TableName() string { return "advertisers" }

func (row *AdvertiserRow) ToDomain() *domain.Advertiser {
	if row == nil {
		return nil
	}
	return &domain.Advertiser{
		ID:          row.ID,
		CompanyName: row.CompanyName,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func AdvertiserRowFromDomain(a *domain.Advertiser) *AdvertiserRow {
	if a == nil {
		return nil
	}
	return &AdvertiserRow{
		ID:          a.ID,
		CompanyName: a.CompanyName,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
