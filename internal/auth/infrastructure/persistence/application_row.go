package persistence

import (
	"time"

	"skykin-platform/internal/auth/domain"

	"github.com/google/uuid"
)

type ApplicationRow struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	DeveloperID uuid.UUID `gorm:"type:uuid;not null"`
	AppName     string    `gorm:"type:varchar(100);not null"`
	Platform    string    `gorm:"type:varchar(50);not null"`
	BundleID    string    `gorm:"type:varchar(150);not null"`
	Status      string    `gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
	APIKeys     []APIKeyRow `gorm:"foreignKey:ApplicationID;constraint:OnDelete:CASCADE"`
}

func (ApplicationRow) TableName() string { return "applications" }

func (row *ApplicationRow) ToDomain() *domain.Application {
	if row == nil {
		return nil
	}
	keys := make([]domain.APIKey, len(row.APIKeys))
	for i := range row.APIKeys {
		keys[i] = *row.APIKeys[i].ToDomain()
	}
	return &domain.Application{
		ID:          row.ID,
		DeveloperID: row.DeveloperID,
		AppName:     row.AppName,
		Platform:    row.Platform,
		BundleID:    row.BundleID,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		APIKeys:     keys,
	}
}

func ApplicationRowFromDomain(a *domain.Application) *ApplicationRow {
	if a == nil {
		return nil
	}
	return &ApplicationRow{
		ID:          a.ID,
		DeveloperID: a.DeveloperID,
		AppName:     a.AppName,
		Platform:    a.Platform,
		BundleID:    a.BundleID,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
