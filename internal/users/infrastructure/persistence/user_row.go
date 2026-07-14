package persistence

import (
	"time"

	"skykin-platform/internal/users/domain"
)

// UserRow is the GORM persistence model for users.
type UserRow struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (UserRow) TableName() string { return "users" }

func (row *UserRow) ToDomain() *domain.User {
	if row == nil {
		return nil
	}
	return &domain.User{
		ID:        row.ID,
		CreatedAt: row.CreatedAt,
	}
}

func UserRowFromDomain(u *domain.User) *UserRow {
	if u == nil {
		return nil
	}
	return &UserRow{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
	}
}
