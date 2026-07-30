package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CandidateStatus string

const (
	CandidateStatusPending    CandidateStatus = "pending"
	CandidateStatusApproved   CandidateStatus = "approved"
	CandidateStatusRejected   CandidateStatus = "rejected"
	CandidateStatusSuperseded CandidateStatus = "superseded"
)

type SegmentCandidate struct {
	ID                 uuid.UUID
	IntentName         string
	UserCount          int
	AvgConfidence      float64
	AvgDaysActive      float64
	MinDaysActive      int
	LookbackDays       int
	Status             CandidateStatus
	ScannedAt          time.Time
	ReviewedBy         *uuid.UUID
	ReviewedAt         *time.Time
	ReviewNotes        string
	PublishedSegmentID *uuid.UUID
}

// UserInCandidate is one pseudonymous member captured by a candidate scan.
type UserInCandidate struct {
	PseudonymousID string
	Confidence     float64
	DaysActive     int
	LastSeenAt     time.Time
}

// UpsertOutcome reports whether a scan created a new pending candidate or refreshed one.
type UpsertOutcome struct {
	CandidateID uuid.UUID
	Created     bool
}

type CandidateRepository interface {
	// UpsertPending atomically creates or refreshes the single pending candidate for
	// an intent and replaces its captured member list.
	UpsertPending(ctx context.Context, c *SegmentCandidate, users []*UserInCandidate) (UpsertOutcome, error)
	FindByStatus(ctx context.Context, status CandidateStatus) ([]*SegmentCandidate, error)
	FindByID(ctx context.Context, id uuid.UUID) (*SegmentCandidate, error)
	// LockPending selects a pending candidate FOR UPDATE so concurrent reviews serialise.
	LockPending(ctx context.Context, id uuid.UUID) (*SegmentCandidate, error)
	GetUsers(ctx context.Context, candidateID uuid.UUID) ([]*UserInCandidate, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status CandidateStatus, reviewedBy uuid.UUID, notes string) error
	LinkToSegment(ctx context.Context, candidateID, segmentID uuid.UUID) error
}

type MembershipRepository interface {
	BulkInsert(ctx context.Context, segmentID uuid.UUID, users []*UserInCandidate) error
	FindPseudonymousIDsInSegment(ctx context.Context, segmentID uuid.UUID) ([]string, error)
	CountMembers(ctx context.Context, segmentID uuid.UUID) (int, error)
}

// Repositories is the transaction-scoped repository set handed to a UnitOfWork.
type Repositories struct {
	Segments   SegmentRepository
	Membership MembershipRepository
	Candidates CandidateRepository
}

// UnitOfWork runs audience writes that must commit or roll back together.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(r Repositories) error) error
}
