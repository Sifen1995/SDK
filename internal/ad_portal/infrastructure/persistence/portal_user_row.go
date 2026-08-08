package persistence

import (
	"time"

	"skykin-platform/internal/ad_portal/domain"
)

type PortalUserRow struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Name         string    `gorm:"type:varchar(255);not null"`
	RoleID       string    `gorm:"type:uuid;not null;index"`
	AdvertiserID *string   `gorm:"type:uuid;index"`
	AnalystID    *string   `gorm:"type:uuid;index"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"`

	Role       *RoleRow       `gorm:"foreignKey:RoleID"`
	Advertiser *AdvertiserRow `gorm:"foreignKey:AdvertiserID"`
	Analyst    *AnalystRow    `gorm:"foreignKey:AnalystID"`
}

func (PortalUserRow) TableName() string { return "portal_users" }

func (row *PortalUserRow) ToDomain() *domain.PortalUser {
	if row == nil {
		return nil
	}
	u := &domain.PortalUser{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Name:         row.Name,
		RoleID:       row.RoleID,
		AdvertiserID: row.AdvertiserID,
		AnalystID:    row.AnalystID,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.Role != nil {
		u.Role = row.Role.ToDomain()
	}
	if row.Advertiser != nil {
		u.Advertiser = row.Advertiser.ToDomain()
	}
	if row.Analyst != nil {
		u.Analyst = row.Analyst.ToDomain()
	}
	return u
}

func PortalUserRowFromDomain(u *domain.PortalUser) *PortalUserRow {
	if u == nil {
		return nil
	}
	return &PortalUserRow{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Name:         u.Name,
		RoleID:       u.RoleID,
		AdvertiserID: u.AdvertiserID,
		AnalystID:    u.AnalystID,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
