package model

import (
	"time"

	"gorm.io/datatypes"
)

// Campaign stores targeting, budget, and creative payload in one row (table: campaigns).
type Campaign struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID   string         `gorm:"type:uuid;not null;index"`
	Name           string         `gorm:"type:varchar(255);not null"`
	TargetIntent   string         `gorm:"type:varchar(100);not null;index"`
	ApplicationID  string         `gorm:"type:varchar(255);not null;index"`
	CreativeFormat string         `gorm:"type:varchar(50);not null"`
	Title          string         `gorm:"type:varchar(255)"`
	BodyText       string         `gorm:"type:text"`
	ImageURL       string         `gorm:"type:text"`
	DestinationURL string         `gorm:"type:text;not null"`
	CanvasJSON     datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	DailyBudgetCap float64        `gorm:"type:numeric(12,2);not null;default:0"`
	TotalBudgetCap float64        `gorm:"type:numeric(12,2);not null;default:0"`
	IsActive       bool           `gorm:"not null;default:false;index"`
	ValidationStatus string       `gorm:"type:varchar(20);not null;default:pending"`
	ValidationNotes  string       `gorm:"type:text"`
	CreatedAt      time.Time      `gorm:"not null;default:now()"`
	UpdatedAt      time.Time      `gorm:"not null;default:now()"`
}

func (Campaign) TableName() string { return "campaigns" }
