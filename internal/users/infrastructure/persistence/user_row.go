package persistence

import (
	"time"

	"skykin-platform/internal/users/domain"
)

// UserRow is the GORM persistence model for users.
type UserRow struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExternalUserID string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	PhoneNumber    *string   `gorm:"type:varchar(20);null;index:idx_users_phone_number"`
	CreatedAt      time.Time `gorm:"not null;default:now()"`
}

func (UserRow) TableName() string { return "users" }

func (row *UserRow) ToDomain() *domain.User {
	if row == nil {
		return nil
	}
	phone := ""
	if row.PhoneNumber != nil {
		phone = *row.PhoneNumber
	}
	return &domain.User{
		ID:             row.ID,
		ExternalUserID: row.ExternalUserID,
		PhoneNumber:    phone,
		CreatedAt:      row.CreatedAt,
	}
}

func UserRowFromDomain(u *domain.User) *UserRow {
	if u == nil {
		return nil
	}
	row := &UserRow{
		ID:             u.ID,
		ExternalUserID: u.ExternalUserID,
		CreatedAt:      u.CreatedAt,
	}
	if u.PhoneNumber != "" {
		phone := u.PhoneNumber
		row.PhoneNumber = &phone
	}
	return row
}
