package application

import (
	"context"
	"errors"
	"log/slog"

	"skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/model"

	"github.com/google/uuid"
)

type ApproveCandidateUseCase struct {
	candidateRepo  domain.CandidateRepository
	membershipRepo domain.MembershipRepository
	segments       SegmentPublisher
	logger         *slog.Logger
}

// SegmentPublisher creates audience segments (implemented by PlanAndSegmentService).
type SegmentPublisher interface {
	CreateSegment(ctx context.Context, cmd CreateSegmentCmd) (*model.AudienceSegment, error)
}

func NewApproveCandidateUseCase(
	candidateRepo domain.CandidateRepository,
	membershipRepo domain.MembershipRepository,
	segments SegmentPublisher,
	logger *slog.Logger,
) *ApproveCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ApproveCandidateUseCase{
		candidateRepo: candidateRepo, membershipRepo: membershipRepo,
		segments: segments, logger: logger,
	}
}

func (uc *ApproveCandidateUseCase) Execute(
	ctx context.Context,
	candidateID uuid.UUID,
	adminID uuid.UUID,
	name, description string,
	estimatedCPM float64,
) (*model.AudienceSegment, error) {
	candidate, err := uc.candidateRepo.FindByID(ctx, candidateID)
	if err != nil {
		return nil, errors.New("candidate not found")
	}
	if candidate.Status != domain.CandidateStatusPending {
		return nil, errors.New("candidate is not pending")
	}
	segment, err := uc.segments.CreateSegment(ctx, CreateSegmentCmd{
		Name: name, Description: description,
		TopIntentSignals: []string{candidate.IntentName},
		ApproximateSize: candidate.UserCount, EstimatedCPM: estimatedCPM, IsActive: true,
	})
	if err != nil {
		return nil, err
	}
	segID, err := uuid.Parse(segment.ID)
	if err != nil {
		return nil, err
	}
	users, err := uc.candidateRepo.GetUsers(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	if err := uc.membershipRepo.BulkInsert(ctx, segID, users); err != nil {
		return nil, err
	}
	if err := uc.candidateRepo.UpdateStatus(ctx, candidateID, domain.CandidateStatusApproved, adminID, ""); err != nil {
		return nil, err
	}
	if err := uc.candidateRepo.LinkToSegment(ctx, candidateID, segID); err != nil {
		return nil, err
	}
	uc.logger.Info("candidate approved", "candidate_id", candidateID, "segment_id", segID)
	return segment, nil
}
