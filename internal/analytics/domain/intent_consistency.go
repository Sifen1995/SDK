package domain

import (
	"time"

	"github.com/google/uuid"
)

// TopicIntentConsistencyFinding is published when sustained intent patterns are detected.
const TopicIntentConsistencyFinding = "IntentConsistencyFinding"

// ConsistentUser is a pseudonymous identity exhibiting sustained interest in an intent class.
type ConsistentUser struct {
	PseudonymousID string
	Confidence     float64
	DaysActive     int
	LastSeenAt     time.Time
}

// IntentConsistencyFinding is an analytics insight — not a persisted business entity.
type IntentConsistencyFinding struct {
	FindingID     uuid.UUID
	IntentName    string
	Users         []*ConsistentUser
	UserCount     int
	AvgConfidence float64
	AvgDaysActive float64
	MinDaysActive int
	LookbackDays  int
	ScannedAt     time.Time
}
