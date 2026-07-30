package application

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"
)

// ErrCandidateNotPending means the candidate was already approved or rejected.
var ErrCandidateNotPending = errors.New("candidate is not pending")

// ApprovedSegment is the admin-facing result of publishing a candidate.
type ApprovedSegment struct {
	SegmentID   string
	MemberCount int
}

// SegmentPublisher provisions the audience segment for an approved candidate.
// The audience module implements it and must commit the segment row, its
// memberships and the candidate status change together.
type SegmentPublisher interface {
	PublishFromCandidate(
		ctx context.Context,
		candidateID, adminID uuid.UUID,
		name, description string,
		estimatedCPM float64,
	) (ApprovedSegment, error)
}

// ApproveCandidateUseCase validates operator input and publishes the segment.
type ApproveCandidateUseCase struct {
	publisher SegmentPublisher
	logger    *slog.Logger
}

func NewApproveCandidateUseCase(publisher SegmentPublisher, logger *slog.Logger) *ApproveCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ApproveCandidateUseCase{publisher: publisher, logger: logger}
}

func (uc *ApproveCandidateUseCase) Execute(
	ctx context.Context,
	candidateID, adminID uuid.UUID,
	name, description string,
	estimatedCPM float64,
) (ApprovedSegment, error) {
	if uc.publisher == nil {
		return ApprovedSegment{}, errors.New("segment publisher not configured")
	}
	if strings.TrimSpace(name) == "" {
		return ApprovedSegment{}, errors.New("name is required")
	}
	if estimatedCPM <= 0 {
		return ApprovedSegment{}, errors.New("estimated_cpm must be > 0")
	}

	segment, err := uc.publisher.PublishFromCandidate(ctx, candidateID, adminID, name, description, estimatedCPM)
	if err != nil {
		return ApprovedSegment{}, err
	}

	uc.logger.Info("segment candidate approved",
		"candidate_id", candidateID,
		"segment_id", segment.SegmentID,
		"member_count", segment.MemberCount,
	)
	return segment, nil
}
