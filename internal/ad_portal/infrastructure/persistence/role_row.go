package persistence

import (
	"time"

	"skykin-platform/internal/ad_portal/domain"
)

type RoleRow struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Slug        string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	DisplayName string    `gorm:"type:varchar(100);not null"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
}

func (RoleRow) TableName() string { return "roles" }

func (row *RoleRow) ToDomain() *domain.Role {
	if row == nil {
		return nil
	}
	return &domain.Role{
		ID:          row.ID,
		Slug:        row.Slug,
		DisplayName: row.DisplayName,
		CreatedAt:   row.CreatedAt,
	}
}

func RoleRowFromDomain(r *domain.Role) *RoleRow {
	if r == nil {
		return nil
	}
	return &RoleRow{
		ID:          r.ID,
		Slug:        r.Slug,
		DisplayName: r.DisplayName,
		CreatedAt:   r.CreatedAt,
	}
}
