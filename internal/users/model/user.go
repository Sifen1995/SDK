package model

import "time"

type Users struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExternalUserID string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"external_user_id"`
	PhoneNumber    *string   `gorm:"type:varchar(20);null;index:idx_users_phone_number"`
	CreatedAt      time.Time `json:"created_at"`
}

func (Users) TableName() string {
	return "users"
}
