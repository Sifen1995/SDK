package model

import "time"

// Advertiser is a campaign portal account (table: advertisers).
type Advertiser struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyName  string    `gorm:"type:varchar(255);not null"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	APIKey       string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	Role         string    `gorm:"type:varchar(50);not null;default:advertiser;index"`
	ContactName  string    `gorm:"type:varchar(255)"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"`
}

func (Advertiser) TableName() string { return "advertisers" }
