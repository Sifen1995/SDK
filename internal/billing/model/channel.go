package model

import "time"

type Channel struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	// Unique on code is enforced by SQL migration (channels_code_key); avoid GORM constraint rename drift.
	Code        string    `gorm:"type:varchar(50);not null;unique"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:text"`
	IsPremium   bool      `gorm:"not null;default:false"`
	IsActive    bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
}

func (Channel) TableName() string { return "channels" }
