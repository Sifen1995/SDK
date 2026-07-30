package persistence

import (
	"time"

	"skykin-platform/internal/intents/domain"
)

// IntentRow is the GORM persistence model for intents.
type IntentRow struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PseudonymousID string    `gorm:"column:pseudonymous_id;type:uuid;not null;index"`
	IntentName     string    `gorm:"type:varchar(100);not null;index"`
	Confidence     float64   `gorm:"type:numeric(4,3);not null"`
	CreatedAt      time.Time `gorm:"not null;default:now()"`
}

func (IntentRow) TableName() string { return "intents" }

func (row *IntentRow) ToDomain() *domain.Intent {
	if row == nil {
		return nil
	}
	return &domain.Intent{
		ID:             row.ID,
		PseudonymousID: row.PseudonymousID,
		IntentName:     row.IntentName,
		Confidence:     row.Confidence,
		CreatedAt:      row.CreatedAt,
	}
}

func IntentRowFromDomain(i *domain.Intent) *IntentRow {
	if i == nil {
		return nil
	}
	return &IntentRow{
		ID:             i.ID,
		PseudonymousID: i.PseudonymousID,
		IntentName:     i.IntentName,
		Confidence:     i.Confidence,
		CreatedAt:      i.CreatedAt,
	}
}
