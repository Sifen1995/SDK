package model

import "time"

// Advertiser is a company account (table: advertisers). Auth is via portal_users.
type Advertiser struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyName string    `gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}

func (Advertiser) TableName() string { return "advertisers" }
