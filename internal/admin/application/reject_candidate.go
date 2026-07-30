package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

// CandidateRejecter marks a segment candidate rejected. Implemented by the audience module.
type CandidateRejecter interface {
	RejectCandidate(ctx context.Context, candidateID, adminID uuid.UUID, notes string) error
}

// RejectCandidateUseCase records an operator rejection.
type RejectCandidateUseCase struct {
	rejecter CandidateRejecter
	logger   *slog.Logger
}

func NewRejectCandidateUseCase(rejecter CandidateRejecter, logger *slog.Logger) *RejectCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &RejectCandidateUseCase{rejecter: rejecter, logger: logger}
}

func (uc *RejectCandidateUseCase) Execute(ctx context.Context, candidateID, adminID uuid.UUID, notes string) error {
	if uc.rejecter == nil {
		return errors.New("candidate rejecter not configured")
	}
	if err := uc.rejecter.RejectCandidate(ctx, candidateID, adminID, notes); err != nil {
		return err
	}
	uc.logger.Info("segment candidate rejected", "candidate_id", candidateID)
	return nil
}
