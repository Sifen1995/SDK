package persistence

import (
	"encoding/json"
	"fmt"
	"time"

	"skykin-platform/internal/audience/domain"

	"gorm.io/datatypes"
)

// AudienceSegmentRow is the GORM persistence model for audience_segments.
type AudienceSegmentRow struct {
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

func (AudienceSegmentRow) TableName() string { return "audience_segments" }

// converts database row to Go domain object
func (row *AudienceSegmentRow) ToDomain() (*domain.AudienceSegment, error) {
	if row == nil {
		return nil, nil
	}
	var signals []string
	if len(row.TopIntentSignals) > 0 {
		if err := json.Unmarshal(row.TopIntentSignals, &signals); err != nil {
			return nil, fmt.Errorf("unmarshal top_intent_signals: %w", err)
		}
	}
	return &domain.AudienceSegment{
		ID:               row.ID,
		Name:             row.Name,
		Description:      row.Description,
		TopIntentSignals: signals,
		ApproximateSize:  row.ApproximateSize,
		EstimatedCPM:     row.EstimatedCPM,
		AvailableFrom:    row.AvailableFrom,
		AvailableUntil:   row.AvailableUntil,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}, nil
}

func AudienceSegmentRowFromDomain(seg *domain.AudienceSegment) (*AudienceSegmentRow, error) {
	if seg == nil {
		return nil, nil
	}
	signals := seg.TopIntentSignals
	if signals == nil {
		signals = []string{}
	}
	raw, err := json.Marshal(signals)
	if err != nil {
		return nil, fmt.Errorf("marshal top_intent_signals: %w", err)
	}
	return &AudienceSegmentRow{
		ID:               seg.ID,
		Name:             seg.Name,
		Description:      seg.Description,
		TopIntentSignals: raw,
		ApproximateSize:  seg.ApproximateSize,
		EstimatedCPM:     seg.EstimatedCPM,
		AvailableFrom:    seg.AvailableFrom,
		AvailableUntil:   seg.AvailableUntil,
		IsActive:         seg.IsActive,
		CreatedAt:        seg.CreatedAt,
		UpdatedAt:        seg.UpdatedAt,
	}, nil
}
