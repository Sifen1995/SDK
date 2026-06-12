package model

import (
	"time"

	"gorm.io/datatypes"
)

// Campaign stores targeting, budget, creative payload, and delivery channel in one row.
type Campaign struct {
	ID           string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID string `gorm:"type:uuid;not null;index"`

	Name         string `gorm:"type:varchar(255);not null"`
	TargetIntent string `gorm:"type:varchar(100);not null;index"`

	// ChannelID references channels.id (IN_APP_BANNER, PUSH, SMS_PLUS, …).
	ChannelID string  `gorm:"type:uuid;not null;index"`
	// SegmentID is optional; nil = free intent-only targeting via the targeting job.
	SegmentID *string `gorm:"type:uuid;index"`

	Title          string         `gorm:"type:varchar(255)"`
	BodyText       string         `gorm:"type:text"`
	ImageURL       string         `gorm:"type:text"`
	DestinationURL string         `gorm:"type:text"`
	CanvasJSON     datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`

	BillingModel string  `gorm:"type:varchar(20);not null;default:'CPC'"`
	DailyBudgetCap float64 `gorm:"type:numeric(12,2);not null;default:0"`
	TotalBudgetCap float64 `gorm:"type:numeric(12,2);not null;default:0"`
	BudgetSpent    float64 `gorm:"type:numeric(12,2);not null;default:0"`

	FrequencyCapPerDay int `gorm:"not null;default:3"`

	ScheduledStartAt *time.Time `gorm:"index"`
	ScheduledEndAt   *time.Time `gorm:"index"`

	IsActive         bool   `gorm:"not null;default:false;index"`
	ValidationStatus string `gorm:"type:varchar(20);not null;default:'pending'"`
	ValidationNotes  string `gorm:"type:text"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (Campaign) TableName() string { return "campaigns" }
