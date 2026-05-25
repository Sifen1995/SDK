package model

import (
	"time"
	"github.com/google/uuid"
)

type Developer struct {
	ID           uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name         string        `gorm:"type:varchar(100);not null" json:"name"`
	Email        string        `gorm:"type:varchar(150);not null;unique" json:"email"`
	PasswordHash string        `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time     `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"not null;default:now()" json:"updated_at"`
	Applications []Application `gorm:"foreignKey:DeveloperID;constraint:OnDelete:CASCADE" json:"applications,omitempty"`
}

type Application struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	DeveloperID uuid.UUID `gorm:"type:uuid;not null" json:"developer_id"`
	AppName     string    `gorm:"type:varchar(100);not null" json:"app_name"`
	Platform    string    `gorm:"type:varchar(50);not null" json:"platform"`   // e.g., "flutter"
	BundleID    string    `gorm:"type:varchar(150);not null" json:"bundle_id"` // e.g., "com.shega.app"
	Status      string    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
	APIKeys     []APIKey  `gorm:"foreignKey:ApplicationID;constraint:OnDelete:CASCADE" json:"api_keys,omitempty"`
}

type APIKey struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ApplicationID uuid.UUID  `gorm:"type:uuid;not null" json:"application_id"`
	KeyValue      string     `gorm:"type:varchar(255);not null;unique" json:"key_value"`
	SecretKeyValue  string     `gorm:"type:varchar(255);not null;unique" json:"secret_key_value"` // SHA-256 Hash of Secret Key (NEW)
	IsActive      bool       `gorm:"type:boolean;not null;default:true" json:"is_active"`
	RateLimit     int        `gorm:"type:integer;not null;default:60" json:"rate_limit"`
	CreatedAt     time.Time  `gorm:"not null;default:now()" json:"created_at"`
	ExpiresAt     *time.Time `gorm:"default:null" json:"expires_at,omitempty"`
}