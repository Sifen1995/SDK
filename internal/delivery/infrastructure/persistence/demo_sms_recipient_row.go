package persistence

import (
	"time"

	"github.com/google/uuid"
)

type DemoSMSRecipientRow struct {
	UserID             int64      `gorm:"column:user_id;primaryKey;autoIncrement:false"`
	PhoneE164          string     `gorm:"column:phone_e164;type:text;not null;uniqueIndex"`
	DisplayName        string     `gorm:"column:display_name;type:text"`
	PseudonymousID     *uuid.UUID `gorm:"column:pseudonymous_id;type:uuid"`
	IsActive           bool       `gorm:"column:is_active;not null;default:true"`
	IsMock             bool       `gorm:"column:is_mock;not null;default:true"`
	ProviderExternalID string     `gorm:"column:provider_external_id;type:text"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (DemoSMSRecipientRow) TableName() string { return "demo_sms_recipients" }
