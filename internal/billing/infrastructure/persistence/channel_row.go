package persistence

import (
	"time"

	"skykin-platform/internal/billing/domain"
)

// ChannelRow is the GORM persistence model for channels.
type ChannelRow struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code        string    `gorm:"type:varchar(50);not null;unique"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:text"`
	IsPremium   bool      `gorm:"not null;default:false"`
	IsActive    bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
}

func (ChannelRow) TableName() string { return "channels" }

func (row *ChannelRow) ToDomain() *domain.Channel {
	if row == nil {
		return nil
	}
	return &domain.Channel{
		ID:          row.ID,
		Code:        row.Code,
		Name:        row.Name,
		Description: row.Description,
		IsPremium:   row.IsPremium,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt,
	}
}

func ChannelRowFromDomain(ch *domain.Channel) *ChannelRow {
	if ch == nil {
		return nil
	}
	return &ChannelRow{
		ID:          ch.ID,
		Code:        ch.Code,
		Name:        ch.Name,
		Description: ch.Description,
		IsPremium:   ch.IsPremium,
		IsActive:    ch.IsActive,
		CreatedAt:   ch.CreatedAt,
	}
}
