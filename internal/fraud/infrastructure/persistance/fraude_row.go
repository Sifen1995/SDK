package persistance

import (
	"database/sql"
	"time"
)

// BlockedDomainsRow maps the blocked_domains table.
type BlockedDomainsRow struct {
	Domain     string       `gorm:"column:domain;type:varchar(255);primaryKey"`
	ThreatType string       `gorm:"column:threat_type;type:varchar(64);not null"`
	Severity   string       `gorm:"column:severity;type:varchar(32);not null"`
	Source     string       `gorm:"column:source;type:varchar(64);not null"`
	Status     string       `gorm:"column:status;type:varchar(32);not null;default:active;index"`
	CreatedAt  time.Time    `gorm:"column:created_at;not null;default:now()"`
	ExpiresAt  sql.NullTime `gorm:"column:expires_at"`
}

// BlockedSendersRow maps the blocked_senders table.
type BlockedSendersRow struct {
	SenderHash string    `gorm:"column:sender_hash;type:varchar(64);primaryKey"`
	ThreatType string    `gorm:"column:threat_type;type:varchar(64);not null"`
	Severity   string    `gorm:"column:severity;type:varchar(32);not null"`
	Source     string    `gorm:"column:source;type:varchar(64);not null"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:now()"`
}

// ScamPatternsRow maps the scam_patterns table.
type ScamPatternsRow struct {
	ID             string    `gorm:"column:id;type:varchar(64);primaryKey"`
	PatternType    string    `gorm:"column:pattern_type;type:varchar(32);not null"`
	PatternValue   string    `gorm:"column:pattern_value;type:text;not null"`
	ThreatCategory string    `gorm:"column:threat_category;type:varchar(64);not null"`
	Confidence     float64   `gorm:"column:confidence;type:numeric(3,2);not null"`
	Language       string    `gorm:"column:language;type:varchar(8);not null;default:any"`
	IsActive       bool      `gorm:"column:is_active;not null;default:true;index"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:now()"`
}

// ThreatReportsRow maps the threat_reports table. SenderHash and URLDomain are
// nullable because a report can describe a sender, a URL, or both.
type ThreatReportsRow struct {
	ID              string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	ThreatType      string         `gorm:"column:threat_type;type:varchar(64);not null"`
	Severity        string         `gorm:"column:severity;type:varchar(32);not null"`
	SenderHash      sql.NullString `gorm:"column:sender_hash;type:varchar(64);index"`
	URLDomain       sql.NullString `gorm:"column:url_domain;type:varchar(255);index"`
	DetectionSource string         `gorm:"column:detection_source;type:varchar(32);not null"`
	SDKVersion      string         `gorm:"column:sdk_version;type:varchar(32);not null"`
	ReportedAt      time.Time      `gorm:"column:reported_at;not null;default:now();index"`
}

func (BlockedDomainsRow) TableName() string { return "blocked_domains" }

func (BlockedSendersRow) TableName() string { return "blocked_senders" }

func (ScamPatternsRow) TableName() string { return "scam_patterns" }

func (ThreatReportsRow) TableName() string { return "threat_reports" }
