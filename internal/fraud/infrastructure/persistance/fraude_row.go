package persistance

import (
	"database/sql"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"
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
	UpdatedAt  time.Time    `gorm:"column:updated_at;not null;default:now();index"`
}

// BlockedSendersRow maps the blocked_senders table.
type BlockedSendersRow struct {
	SenderHash string    `gorm:"column:sender_hash;type:varchar(64);primaryKey"`
	ThreatType string    `gorm:"column:threat_type;type:varchar(64);not null"`
	Severity   string    `gorm:"column:severity;type:varchar(32);not null"`
	Source     string    `gorm:"column:source;type:varchar(64);not null"`
	Status     string    `gorm:"column:status;type:varchar(32);not null;default:active;index"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:now();index"`
}

// ScamPatternsRow maps the scam_patterns table.
type ScamPatternsRow struct {
	ID             string    `gorm:"column:id;type:varchar(64);primaryKey"`
	PatternType    string    `gorm:"column:pattern_type;type:varchar(32);not null"`
	PatternValue   string    `gorm:"column:pattern_value;type:text;not null"`
	ThreatCategory string    `gorm:"column:threat_category;type:varchar(64);not null"`
	Confidence     float64   `gorm:"column:confidence;type:numeric(3,2);not null"`
	Language       string    `gorm:"column:language;type:varchar(8);not null;default:any"`
	IsActive       bool      `gorm:"column:is_active;not null;index"`
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

func (r BlockedDomainsRow) ToDomain() frauddomain.BlockedDomain {
	var expiresAt *time.Time
	if r.ExpiresAt.Valid {
		value := r.ExpiresAt.Time
		expiresAt = &value
	}
	return frauddomain.BlockedDomain{
		Domain:     r.Domain,
		ThreatType: r.ThreatType,
		Severity:   r.Severity,
		Source:     r.Source,
		Status:     r.Status,
		CreatedAt:  r.CreatedAt,
		ExpiresAt:  expiresAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func BlockedDomainsRowFromDomain(value frauddomain.BlockedDomain) BlockedDomainsRow {
	row := BlockedDomainsRow{
		Domain:     value.Domain,
		ThreatType: value.ThreatType,
		Severity:   value.Severity,
		Source:     value.Source,
		Status:     value.Status,
		CreatedAt:  value.CreatedAt,
		UpdatedAt:  value.UpdatedAt,
	}
	if value.ExpiresAt != nil {
		row.ExpiresAt = sql.NullTime{Time: *value.ExpiresAt, Valid: true}
	}
	return row
}

func (r BlockedSendersRow) ToDomain() frauddomain.BlockedSender {
	return frauddomain.BlockedSender{
		SenderHash: r.SenderHash,
		ThreatType: r.ThreatType,
		Severity:   r.Severity,
		Source:     r.Source,
		Status:     r.Status,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func BlockedSendersRowFromDomain(value frauddomain.BlockedSender) BlockedSendersRow {
	return BlockedSendersRow{
		SenderHash: value.SenderHash,
		ThreatType: value.ThreatType,
		Severity:   value.Severity,
		Source:     value.Source,
		Status:     value.Status,
		CreatedAt:  value.CreatedAt,
		UpdatedAt:  value.UpdatedAt,
	}
}

func (r ScamPatternsRow) ToDomain() frauddomain.ScamPattern {
	return frauddomain.ScamPattern{
		ID:             r.ID,
		PatternType:    r.PatternType,
		PatternValue:   r.PatternValue,
		ThreatCategory: r.ThreatCategory,
		Confidence:     r.Confidence,
		Language:       r.Language,
		IsActive:       r.IsActive,
		UpdatedAt:      r.UpdatedAt,
	}
}

func ScamPatternsRowFromDomain(value frauddomain.ScamPattern) ScamPatternsRow {
	return ScamPatternsRow{
		ID:             value.ID,
		PatternType:    value.PatternType,
		PatternValue:   value.PatternValue,
		ThreatCategory: value.ThreatCategory,
		Confidence:     value.Confidence,
		Language:       value.Language,
		IsActive:       value.IsActive,
		UpdatedAt:      value.UpdatedAt,
	}
}

func (r ThreatReportsRow) ToDomain() frauddomain.ThreatReport {
	var senderHash, urlDomain *string
	if r.SenderHash.Valid {
		value := r.SenderHash.String
		senderHash = &value
	}
	if r.URLDomain.Valid {
		value := r.URLDomain.String
		urlDomain = &value
	}
	return frauddomain.ThreatReport{
		ID:              r.ID,
		ThreatType:      r.ThreatType,
		Severity:        r.Severity,
		SenderHash:      senderHash,
		URLDomain:       urlDomain,
		DetectionSource: r.DetectionSource,
		SDKVersion:      r.SDKVersion,
		ReportedAt:      r.ReportedAt,
	}
}

func ThreatReportsRowFromDomain(value frauddomain.ThreatReport) ThreatReportsRow {
	row := ThreatReportsRow{
		ID:              value.ID,
		ThreatType:      value.ThreatType,
		Severity:        value.Severity,
		DetectionSource: value.DetectionSource,
		SDKVersion:      value.SDKVersion,
		ReportedAt:      value.ReportedAt,
	}
	if value.SenderHash != nil {
		row.SenderHash = sql.NullString{String: *value.SenderHash, Valid: true}
	}
	if value.URLDomain != nil {
		row.URLDomain = sql.NullString{String: *value.URLDomain, Valid: true}
	}
	return row
}
