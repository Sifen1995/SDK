package persistence

import (
	"time"

	"skykin-platform/internal/consent/domain"

	"github.com/google/uuid"
)

// PseudonymousMappingRow is the GORM model for pseudonymous_mappings.
type PseudonymousMappingRow struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         int64     `gorm:"type:bigint;not null;uniqueIndex"`
	PseudonymousID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	CreatedAt      time.Time `gorm:"not null;default:now()"`
}

func (PseudonymousMappingRow) TableName() string { return "pseudonymous_mappings" }

func (r *PseudonymousMappingRow) ToDomain() *domain.PseudonymousMapping {
	return &domain.PseudonymousMapping{
		ID:             r.ID,
		UserID:         r.UserID,
		PseudonymousID: r.PseudonymousID,
		CreatedAt:      r.CreatedAt,
	}
}

func MappingRowFromDomain(m *domain.PseudonymousMapping) *PseudonymousMappingRow {
	return &PseudonymousMappingRow{
		ID:             m.ID,
		UserID:         m.UserID,
		PseudonymousID: m.PseudonymousID,
		CreatedAt:      m.CreatedAt,
	}
}
