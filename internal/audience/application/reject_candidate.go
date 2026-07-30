package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"skykin-platform/internal/audience/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RejectCandidateUseCase marks a segment candidate as rejected.
type RejectCandidateUseCase struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
}

func NewRejectCandidateUseCase(uow domain.UnitOfWork, logger *slog.Logger) *RejectCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &RejectCandidateUseCase{uow: uow, logger: logger}
}

func (uc *RejectCandidateUseCase) Execute(ctx context.Context, candidateID, adminID uuid.UUID, notes string) error {
	err := uc.uow.Do(ctx, func(r domain.Repositories) error {
		if _, err := r.Candidates.LockPending(ctx, candidateID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCandidateNotPending
			}
			return fmt.Errorf("lock candidate: %w", err)
		}
		if err := r.Candidates.UpdateStatus(ctx, candidateID, domain.CandidateStatusRejected, adminID, notes); err != nil {
			return fmt.Errorf("update candidate status: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	uc.logger.Info("candidate rejected", "candidate_id", candidateID)
	return nil
}
