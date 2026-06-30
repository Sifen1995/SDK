package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	adminEvents "skykin-platform/internal/admin/events"
	"skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/model"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ProcessApprovedCandidateUseCase materializes an audience segment from an approved candidate.
type ProcessApprovedCandidateUseCase struct {
	segments   domain.SegmentRepository
	membership domain.MembershipRepository
	candidates domain.CandidateRepository
	logger     *slog.Logger
}

func NewProcessApprovedCandidateUseCase(
	segments domain.SegmentRepository,
	membership domain.MembershipRepository,
	candidates domain.CandidateRepository,
	logger *slog.Logger,
) *ProcessApprovedCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProcessApprovedCandidateUseCase{
		segments:   segments,
		membership: membership,
		candidates: candidates,
		logger:     logger,
	}
}

func (uc *ProcessApprovedCandidateUseCase) Execute(ctx context.Context, evt adminEvents.CandidateApprovedEvent) error {
	candidate, err := uc.candidates.FindByID(ctx, evt.CandidateID)
	if err != nil {
		return errors.New("candidate not found")
	}
	if candidate.Status != domain.CandidateStatusPending {
		return errors.New("candidate is not pending")
	}
	if err := uc.candidates.UpdateStatus(ctx, evt.CandidateID, domain.CandidateStatusApproved, evt.AdminID, ""); err != nil {
		return fmt.Errorf("update candidate status: %w", err)
	}

	intentName := evt.IntentName
	if intentName == "" {
		intentName = candidate.IntentName
	}
	userCount := evt.UserCount
	if userCount == 0 {
		userCount = candidate.UserCount
	}

	signalsJSON, err := json.Marshal([]string{intentName})
	if err != nil {
		return fmt.Errorf("marshal intent signals: %w", err)
	}

	seg := &model.AudienceSegment{
		ID:               uuid.New().String(),
		Name:             evt.Name,
		Description:      evt.Description,
		TopIntentSignals: datatypes.JSON(signalsJSON),
		ApproximateSize:  userCount,
		EstimatedCPM:     evt.EstimatedCPM,
		AvailableFrom:    time.Now().UTC(),
		IsActive:         true,
	}
	if err := uc.segments.Create(ctx, seg); err != nil {
		return fmt.Errorf("create segment: %w", err)
	}

	segID, err := uuid.Parse(seg.ID)
	if err != nil {
		return fmt.Errorf("parse segment id: %w", err)
	}

	users, err := uc.candidates.GetUsers(ctx, evt.CandidateID)
	if err != nil {
		return fmt.Errorf("load candidate users: %w", err)
	}
	if err := uc.membership.BulkInsert(ctx, segID, users); err != nil {
		return fmt.Errorf("bulk insert memberships: %w", err)
	}
	if err := uc.candidates.LinkToSegment(ctx, evt.CandidateID, segID); err != nil {
		return fmt.Errorf("link candidate to segment: %w", err)
	}

	uc.logger.Info("segment published from approved candidate",
		"candidate_id", evt.CandidateID,
		"segment_id", segID,
		"member_count", len(users),
	)
	return nil
}
