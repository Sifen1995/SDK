package persistence

import (
	"encoding/json"
	"fmt"
	"time"

	"skykin-platform/internal/campaigns/domain"

	"gorm.io/datatypes"
)

// CampaignRow is the GORM persistence model for campaigns.
type CampaignRow struct {
	ID           string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID string `gorm:"type:uuid;not null;index"`

	Name         string `gorm:"type:varchar(255);not null"`
	TargetIntent string `gorm:"type:varchar(100);not null;index"`

	ChannelID string  `gorm:"type:uuid;not null;index"`
	SegmentID *string `gorm:"type:uuid;index"`

	Title          string         `gorm:"type:varchar(255)"`
	BodyText       string         `gorm:"type:text"`
	ImageURL       string         `gorm:"type:text"`
	DestinationURL string         `gorm:"type:text"`
	CanvasJSON     datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`

	DailyBudgetCap     float64 `gorm:"type:numeric(12,2);not null;default:0"`
	TotalBudgetCap     float64 `gorm:"type:numeric(12,2);not null;default:0"`
	BudgetSpent        float64 `gorm:"type:numeric(12,2);not null;default:0"`
	FrequencyCapPerDay int     `gorm:"not null;default:3"`

	ScheduledStartAt *time.Time `gorm:"index"`
	ScheduledEndAt   *time.Time `gorm:"index"`

	IsActive         bool   `gorm:"not null;default:false;index"`
	ValidationStatus string `gorm:"type:varchar(20);not null;default:'pending'"`
	ValidationNotes  string `gorm:"type:text"`

	ModerationStatus string     `gorm:"type:varchar(20);not null;default:'pending';index"`
	ModerationNotes  string     `gorm:"type:text"`
	ModeratedAt      *time.Time `gorm:"index"`
	ModeratedBy      *string    `gorm:"type:uuid"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (CampaignRow) TableName() string { return "campaigns" }

func (row *CampaignRow) ToDomain() (*domain.Campaign, error) {
	if row == nil {
		return nil, nil
	}
	var canvas map[string]any
	if len(row.CanvasJSON) > 0 {
		if err := json.Unmarshal(row.CanvasJSON, &canvas); err != nil {
			return nil, fmt.Errorf("unmarshal canvas_json: %w", err)
		}
	}
	return &domain.Campaign{
		ID:                 row.ID,
		AdvertiserID:       row.AdvertiserID,
		Name:               row.Name,
		TargetIntent:       row.TargetIntent,
		ChannelID:          row.ChannelID,
		SegmentID:          row.SegmentID,
		Title:              row.Title,
		BodyText:           row.BodyText,
		ImageURL:           row.ImageURL,
		DestinationURL:     row.DestinationURL,
		CanvasJSON:         canvas,
		DailyBudgetCap:     row.DailyBudgetCap,
		TotalBudgetCap:     row.TotalBudgetCap,
		BudgetSpent:        row.BudgetSpent,
		FrequencyCapPerDay: row.FrequencyCapPerDay,
		ScheduledStartAt:   row.ScheduledStartAt,
		ScheduledEndAt:     row.ScheduledEndAt,
		IsActive:           row.IsActive,
		ValidationStatus:   row.ValidationStatus,
		ValidationNotes:    row.ValidationNotes,
		ModerationStatus:   row.ModerationStatus,
		ModerationNotes:    row.ModerationNotes,
		ModeratedAt:        row.ModeratedAt,
		ModeratedBy:        row.ModeratedBy,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}, nil
}

func CampaignRowFromDomain(c *domain.Campaign) (*CampaignRow, error) {
	if c == nil {
		return nil, nil
	}
	canvas := c.CanvasJSON
	if canvas == nil {
		canvas = map[string]any{}
	}
	raw, err := json.Marshal(canvas)
	if err != nil {
		return nil, fmt.Errorf("marshal canvas_json: %w", err)
	}
	return &CampaignRow{
		ID:                 c.ID,
		AdvertiserID:       c.AdvertiserID,
		Name:               c.Name,
		TargetIntent:       c.TargetIntent,
		ChannelID:          c.ChannelID,
		SegmentID:          c.SegmentID,
		Title:              c.Title,
		BodyText:           c.BodyText,
		ImageURL:           c.ImageURL,
		DestinationURL:     c.DestinationURL,
		CanvasJSON:         raw,
		DailyBudgetCap:     c.DailyBudgetCap,
		TotalBudgetCap:     c.TotalBudgetCap,
		BudgetSpent:        c.BudgetSpent,
		FrequencyCapPerDay: c.FrequencyCapPerDay,
		ScheduledStartAt:   c.ScheduledStartAt,
		ScheduledEndAt:     c.ScheduledEndAt,
		IsActive:           c.IsActive,
		ValidationStatus:   c.ValidationStatus,
		ValidationNotes:    c.ValidationNotes,
		ModerationStatus:   c.ModerationStatus,
		ModerationNotes:    c.ModerationNotes,
		ModeratedAt:        c.ModeratedAt,
		ModeratedBy:        c.ModeratedBy,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}, nil
}
