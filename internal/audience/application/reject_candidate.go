package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	adminEvents "skykin-platform/internal/admin/events"
	"skykin-platform/internal/audience/domain"
)

// RejectCandidateUseCase marks a segment candidate as rejected.
type RejectCandidateUseCase struct {
	candidates domain.CandidateRepository
	logger     *slog.Logger
}

func NewRejectCandidateUseCase(candidates domain.CandidateRepository, logger *slog.Logger) *RejectCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &RejectCandidateUseCase{candidates: candidates, logger: logger}
}

func (uc *RejectCandidateUseCase) Execute(ctx context.Context, evt adminEvents.CandidateRejectedEvent) error {
	candidate, err := uc.candidates.FindByID(ctx, evt.CandidateID)
	if err != nil {
		return errors.New("candidate not found")
	}
	if candidate.Status != domain.CandidateStatusPending {
		return errors.New("candidate is not pending")
	}
	if err := uc.candidates.UpdateStatus(ctx, evt.CandidateID, domain.CandidateStatusRejected, evt.AdminID, evt.Notes); err != nil {
		return fmt.Errorf("update candidate status: %w", err)
	}
	uc.logger.Info("candidate rejected", "candidate_id", evt.CandidateID)
	return nil
}
