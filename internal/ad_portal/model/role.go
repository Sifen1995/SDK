package model

import "time"

type Role struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Slug        string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	DisplayName string    `gorm:"type:varchar(100);not null"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
}

func (Role) TableName() string { return "roles" }
