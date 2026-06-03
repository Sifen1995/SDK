package model

import "time"

type Users struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExternalUserID string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"external_user_id"`
	CreatedAt      time.Time `json:"created_at"`
}

func (Users) TableName() string {
	return "users"
}
