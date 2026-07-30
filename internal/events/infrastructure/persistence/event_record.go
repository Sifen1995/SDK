package persistence

import (
	"time"

	"gorm.io/datatypes"
)

// EventRecord is the GORM persistence model for the events table.
type EventRecord struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventID        string         `gorm:"type:uuid;not null;uniqueIndex"`
	PseudonymousID string         `gorm:"column:pseudonymous_id;type:uuid;not null;index"`
	ApplicationID *string        `gorm:"type:uuid;index"`
	SessionID     *string        `gorm:"type:uuid;index"`
	EventType     string         `gorm:"type:varchar(100);not null;index"`
	Domain        string         `gorm:"type:varchar(100);index"`
	ScreenName    string         `gorm:"type:varchar(255)"`
	Metadata      datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	DeviceType    string         `gorm:"type:varchar(50)"`
	Platform      string         `gorm:"type:varchar(50)"`
	AppVersion    string         `gorm:"type:varchar(50)"`
	CreatedAt     time.Time      `gorm:"not null;index"`
}

func (EventRecord) TableName() string {
	return "events"
}
