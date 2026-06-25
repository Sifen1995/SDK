package application

import (
	"context"
	"errors"
	"log/slog"

	"skykin-platform/internal/audience/domain"

	"github.com/google/uuid"
)

type RejectCandidateUseCase struct {
	candidateRepo domain.CandidateRepository
	logger        *slog.Logger
}

func NewRejectCandidateUseCase(candidateRepo domain.CandidateRepository, logger *slog.Logger) *RejectCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &RejectCandidateUseCase{candidateRepo: candidateRepo, logger: logger}
}

func (uc *RejectCandidateUseCase) Execute(ctx context.Context, candidateID, adminID uuid.UUID, notes string) error {
	candidate, err := uc.candidateRepo.FindByID(ctx, candidateID)
	if err != nil {
		return errors.New("candidate not found")
	}
	if candidate.Status != domain.CandidateStatusPending {
		return errors.New("candidate is not pending")
	}
	if err := uc.candidateRepo.UpdateStatus(ctx, candidateID, domain.CandidateStatusRejected, adminID, notes); err != nil {
		return err
	}
	uc.logger.Info("candidate rejected", "candidate_id", candidateID)
	return nil
}
