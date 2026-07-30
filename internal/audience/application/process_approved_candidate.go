package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"skykin-platform/internal/audience/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrCandidateNotPending is returned when a candidate has already been reviewed.
var ErrCandidateNotPending = errors.New("candidate is not pending")

// ApproveCandidateCmd carries the operator-supplied segment definition.
type ApproveCandidateCmd struct {
	CandidateID  uuid.UUID
	AdminID      uuid.UUID
	Name         string
	Description  string
	EstimatedCPM float64
}

// PublishedSegment describes the segment materialised from an approved candidate.
type PublishedSegment struct {
	SegmentID   string
	MemberCount int
}

// ProcessApprovedCandidateUseCase materializes an audience segment from an approved
// candidate. Every write happens in one transaction: a candidate is never left marked
// approved without its segment and memberships.
type ProcessApprovedCandidateUseCase struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
}

func NewProcessApprovedCandidateUseCase(uow domain.UnitOfWork, logger *slog.Logger) *ProcessApprovedCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProcessApprovedCandidateUseCase{uow: uow, logger: logger}
}

func (uc *ProcessApprovedCandidateUseCase) Execute(ctx context.Context, cmd ApproveCandidateCmd) (*PublishedSegment, error) {
	var published PublishedSegment

	err := uc.uow.Do(ctx, func(r domain.Repositories) error {
		candidate, err := r.Candidates.LockPending(ctx, cmd.CandidateID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCandidateNotPending
			}
			return fmt.Errorf("lock candidate: %w", err)
		}

		users, err := r.Candidates.GetUsers(ctx, cmd.CandidateID)
		if err != nil {
			return fmt.Errorf("load candidate users: %w", err)
		}
		if len(users) == 0 {
			return errors.New("candidate has no captured members; rerun the intent consistency scan")
		}

		seg := &domain.AudienceSegment{
			ID:               uuid.New().String(),
			Name:             cmd.Name,
			Description:      cmd.Description,
			TopIntentSignals: []string{candidate.IntentName},
			ApproximateSize:  len(users),
			EstimatedCPM:     cmd.EstimatedCPM,
			AvailableFrom:    time.Now().UTC(),
			IsActive:         true,
		}
		if err := r.Segments.Create(ctx, seg); err != nil {
			return fmt.Errorf("create segment: %w", err)
		}
		segID, err := uuid.Parse(seg.ID)
		if err != nil {
			return fmt.Errorf("parse segment id: %w", err)
		}

		if err := r.Membership.BulkInsert(ctx, segID, users); err != nil {
			return fmt.Errorf("bulk insert memberships: %w", err)
		}
		if err := r.Candidates.UpdateStatus(ctx, cmd.CandidateID, domain.CandidateStatusApproved, cmd.AdminID, ""); err != nil {
			return fmt.Errorf("update candidate status: %w", err)
		}
		if err := r.Candidates.LinkToSegment(ctx, cmd.CandidateID, segID); err != nil {
			return fmt.Errorf("link candidate to segment: %w", err)
		}

		published = PublishedSegment{SegmentID: seg.ID, MemberCount: len(users)}
		return nil
	})
	if err != nil {
		return nil, err
	}

	uc.logger.Info("segment published from approved candidate",
		"candidate_id", cmd.CandidateID,
		"segment_id", published.SegmentID,
		"member_count", published.MemberCount,
	)
	return &published, nil
}
