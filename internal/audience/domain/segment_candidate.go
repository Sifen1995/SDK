package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CandidateStatus string

const (
	CandidateStatusPending  CandidateStatus = "pending"
	CandidateStatusApproved CandidateStatus = "approved"
	CandidateStatusRejected CandidateStatus = "rejected"
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

type UserInCandidate struct {
	UserID     uuid.UUID
	Confidence float64
	DaysActive int
	LastSeenAt time.Time
}

type CandidateRepository interface {
	Save(ctx context.Context, c *SegmentCandidate, users []*UserInCandidate) error
	FindByStatus(ctx context.Context, status CandidateStatus) ([]*SegmentCandidate, error)
	FindByID(ctx context.Context, id uuid.UUID) (*SegmentCandidate, error)
	FindPendingByIntentName(ctx context.Context, intentName string) (*SegmentCandidate, error)
	UpdateFromFinding(ctx context.Context, id uuid.UUID, c *SegmentCandidate, users []*UserInCandidate) error
	GetUsers(ctx context.Context, candidateID uuid.UUID) ([]*UserInCandidate, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status CandidateStatus, reviewedBy uuid.UUID, notes string) error
	LinkToSegment(ctx context.Context, candidateID, segmentID uuid.UUID) error
}

type MembershipRepository interface {
	BulkInsert(ctx context.Context, segmentID uuid.UUID, users []*UserInCandidate) error
	FindUsersInSegment(ctx context.Context, segmentID uuid.UUID) ([]uuid.UUID, error)
	CountMembers(ctx context.Context, segmentID uuid.UUID) (int, error)
}
