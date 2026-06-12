package model

import (
	"time"

	"gorm.io/datatypes"
)

// AudienceSegment is a purchasable Audiencemart cohort definition (rules, not user rows).
type AudienceSegment struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name             string         `gorm:"type:varchar(255);not null"`
	Description      string         `gorm:"type:text"`
	TopIntentSignals datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
	ApproximateSize  int            `gorm:"not null;default:0"`
	EstimatedCPM     float64        `gorm:"type:numeric(10,2);not null"`
	AvailableFrom    time.Time      `gorm:"not null"`
	AvailableUntil   *time.Time
	IsActive         bool      `gorm:"not null;default:true"`
	CreatedAt        time.Time `gorm:"not null;default:now()"`
	UpdatedAt        time.Time `gorm:"not null;default:now()"`
}

func (AudienceSegment) TableName() string { return "audience_segments" }
