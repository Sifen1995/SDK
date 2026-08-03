package domain

import "time"

const (
	StatusActive  = "active"
	StatusRevoked = "revoked"
)

type BlockedDomain struct {
	Domain     string
	ThreatType string
	Severity   string
	Source     string
	Status     string
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	UpdatedAt  time.Time
}

type BlockedSender struct {
	SenderHash string
	ThreatType string
	Severity   string
	Source     string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ScamPattern struct {
	ID             string
	PatternType    string
	PatternValue   string
	ThreatCategory string
	Confidence     float64
	Language       string
	IsActive       bool
	UpdatedAt      time.Time
}

// SyncSnapshot is a point-in-time fraud intelligence response. For deltas,
// inactive/revoked entries are tombstones that clients remove from local cache.
type SyncSnapshot struct {
	BlockedDomains []BlockedDomain
	BlockedSenders []BlockedSender
	ScamPatterns   []ScamPattern
	NextCursor     time.Time
	IsDelta        bool
}

type ThreatReport struct {
	ID              string
	ThreatType      string
	Severity        string
	SenderHash      *string
	URLDomain       *string
	DetectionSource string
	SDKVersion      string
	ReportedAt      time.Time
}
