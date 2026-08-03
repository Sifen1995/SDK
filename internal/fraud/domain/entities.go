package domain

import "time"

type BlockedDomain struct {
	Domain     string
	ThreatType string
	Severity   string
	Source     string
	Status     string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type BlockedSender struct {
	SenderHash string
	ThreatType string
	Severity   string
	Source     string
	CreatedAt  time.Time
}

type ScamPattern struct {
	PatternType    string
	PatternValue   string
	ThreatCategory string
	Confidence     float64
	Language       string
	IsActive       bool
	UpdatedAt      time.Time
}

type ThreatReport struct {
	ID              string
	ThreatType      string
	SenderHash      string
	Domain          string
	DetectionSource string
	SDKVersion      string
	ReportedAt      time.Time
}
