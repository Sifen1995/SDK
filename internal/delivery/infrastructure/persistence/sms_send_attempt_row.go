package persistence

import "time"

type SMSSendAttemptRow struct {
	ID                string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SendKey           string     `gorm:"column:send_key;type:text;not null;uniqueIndex"`
	CampaignID        string     `gorm:"column:campaign_id;type:uuid;not null;index"`
	PseudonymousID    string     `gorm:"column:pseudonymous_id;type:uuid;not null;index"`
	UserID            int64      `gorm:"column:user_id;type:bigint;not null;index"`
	PhoneE164         string     `gorm:"column:phone_e164;type:text;not null"`
	Provider          string     `gorm:"column:provider;type:varchar(50);not null"`
	ProviderMessageID string     `gorm:"column:provider_message_id;type:text"`
	Status            string     `gorm:"column:status;type:varchar(50);not null;index"`
	MessageBody       string     `gorm:"column:message_body;type:text;not null"`
	TrackingToken     string     `gorm:"column:tracking_token;type:text;not null;uniqueIndex"`
	DestinationURL    string     `gorm:"column:destination_url;type:text;not null"`
	ErrorMessage      string     `gorm:"column:error_message;type:text"`
	SentAt            *time.Time `gorm:"column:sent_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (SMSSendAttemptRow) TableName() string { return "sms_send_attempts" }
