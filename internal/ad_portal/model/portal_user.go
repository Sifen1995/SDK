package model

import "time"

// PortalUser is an ad portal login (table: portal_users).
type PortalUser struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Name         string    `gorm:"type:varchar(255);not null"`
	RoleID       string    `gorm:"type:uuid;not null;index"`
	AdvertiserID *string   `gorm:"type:uuid;index"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"`

	Role       *Role       `gorm:"foreignKey:RoleID"`
	Advertiser *Advertiser `gorm:"foreignKey:AdvertiserID"`
}

func (PortalUser) TableName() string { return "portal_users" }

func (u *PortalUser) RoleSlug() string {
	if u.Role != nil {
		return u.Role.Slug
	}
	return ""
}

func (u *PortalUser) AccountAdvertiserID() string {
	if u.AdvertiserID != nil {
		return *u.AdvertiserID
	}
	return ""
}
