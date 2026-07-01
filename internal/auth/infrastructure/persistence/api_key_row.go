package persistence

import (
	"time"

	"skykin-platform/internal/auth/domain"

	"github.com/google/uuid"
)

type APIKeyRow struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ApplicationID  uuid.UUID  `gorm:"type:uuid;not null"`
	KeyValue       string     `gorm:"type:varchar(255);not null;unique"`
	SecretKeyValue string     `gorm:"type:varchar(255);not null;unique"`
	IsActive       bool       `gorm:"type:boolean;not null;default:true"`
	RateLimit      int        `gorm:"type:integer;not null;default:60"`
	CreatedAt      time.Time  `gorm:"not null;default:now()"`
	ExpiresAt      *time.Time `gorm:"default:null"`
}

func (APIKeyRow) TableName() string { return "api_keys" }

func (row *APIKeyRow) ToDomain() *domain.APIKey {
	if row == nil {
		return nil
	}
	return &domain.APIKey{
		ID:             row.ID,
		ApplicationID:  row.ApplicationID,
		KeyValue:       row.KeyValue,
		SecretKeyValue: row.SecretKeyValue,
		IsActive:       row.IsActive,
		RateLimit:      row.RateLimit,
		CreatedAt:      row.CreatedAt,
		ExpiresAt:      row.ExpiresAt,
	}
}

func APIKeyRowFromDomain(k *domain.APIKey) *APIKeyRow {
	if k == nil {
		return nil
	}
	return &APIKeyRow{
		ID:             k.ID,
		ApplicationID:  k.ApplicationID,
		KeyValue:       k.KeyValue,
		SecretKeyValue: k.SecretKeyValue,
		IsActive:       k.IsActive,
		RateLimit:      k.RateLimit,
		CreatedAt:      k.CreatedAt,
		ExpiresAt:      k.ExpiresAt,
	}
}
