package persistence

import (
	"time"

	"skykin-platform/internal/auth/domain"

	"github.com/google/uuid"
)

type DeveloperRow struct {
	ID           uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name         string        `gorm:"type:varchar(100);not null"`
	Email        string        `gorm:"type:varchar(150);not null;unique"`
	PasswordHash string        `gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time     `gorm:"not null;default:now()"`
	UpdatedAt    time.Time     `gorm:"not null;default:now()"`
	Applications []ApplicationRow `gorm:"foreignKey:DeveloperID;constraint:OnDelete:CASCADE"`
}

func (DeveloperRow) TableName() string { return "developers" }

func (row *DeveloperRow) ToDomain() *domain.Developer {
	if row == nil {
		return nil
	}
	return &domain.Developer{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func DeveloperRowFromDomain(d *domain.Developer) *DeveloperRow {
	if d == nil {
		return nil
	}
	return &DeveloperRow{
		ID:           d.ID,
		Name:         d.Name,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
