package persistence

import (
	"errors"
	"fmt"
	"time"

	"skykin-platform/internal/consent/domain"
)

// ConsentRow is the GORM model for consents.
type ConsentRow struct {
	ID           string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       int64      `gorm:"type:bigint;not null;index"`
	ConsentLevel string     `gorm:"type:varchar(20);not null"`
	IsActive     bool       `gorm:"not null;default:true;index"`
	GrantedAt    *time.Time `gorm:"type:timestamptz"`
	RevokedAt    *time.Time `gorm:"type:timestamptz"`
	SDKVersion   string     `gorm:"type:varchar(20);not null"`
	CreatedAt    time.Time  `gorm:"not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"not null;default:now()"`
}

func (ConsentRow) TableName() string { return "consents" }

func (c *ConsentRow) ToDomain() *domain.Consent {
	return &domain.Consent{
		ID:           c.ID,
		UserID:       c.UserID,
		ConsentLevel: c.ConsentLevel,
		IsActive:     c.IsActive,
		GrantedAt:    c.GrantedAt,
		RevokedAt:    c.RevokedAt,
		SDKVersion:   c.SDKVersion,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

func (c *ConsentRow) FromDomain(consent *domain.Consent) {
	c.ID = consent.ID
	c.UserID = consent.UserID
	c.ConsentLevel = consent.ConsentLevel
	c.IsActive = consent.IsActive
	c.GrantedAt = consent.GrantedAt
	c.RevokedAt = consent.RevokedAt
	c.SDKVersion = consent.SDKVersion
	c.CreatedAt = consent.CreatedAt
	c.UpdatedAt = consent.UpdatedAt
}

func (c *ConsentRow) Validate() error {
	if c.UserID == 0 {
		return errors.New("user ID cannot be empty")
	}
	if c.ConsentLevel != "individual" && c.ConsentLevel != "aggregate" && c.ConsentLevel != "none" {
		return fmt.Errorf("invalid consent level: %s", c.ConsentLevel)
	}
	return nil
}
