package model

import (
	"time"
)

// Users represents the users table in PostgreSQL
type Users struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExternalUserID string    `gorm:"type:varchar(255);unique;not null" json:"external_user_id"`
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`
}

// TableName explicitly tells GORM to use the "users" table
func (Users) TableName() string {
	return "users"
}
