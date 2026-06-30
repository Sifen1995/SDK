package application

import (
	"context"
	"errors"
	"log/slog"

	adminEvents "skykin-platform/internal/admin/events"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
)

// RejectCandidateUseCase publishes a candidate rejection for async audience processing.
type RejectCandidateUseCase struct {
	bus    *messaging.Bus
	logger *slog.Logger
}

func NewRejectCandidateUseCase(bus *messaging.Bus, logger *slog.Logger) *RejectCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &RejectCandidateUseCase{bus: bus, logger: logger}
}

func (uc *RejectCandidateUseCase) Execute(ctx context.Context, candidateID, adminID uuid.UUID, notes string) error {
	if uc.bus == nil {
		return errors.New("event bus not configured")
	}
	uc.bus.Publish(messaging.Event{
		Name: adminEvents.TopicCandidateRejected,
		Ctx:  ctx,
		Payload: adminEvents.CandidateRejectedEvent{
			CandidateID: candidateID,
			AdminID:     adminID,
			Notes:       notes,
		},
	})
	uc.logger.Info("candidate rejection requested", "candidate_id", candidateID)
	return nil
}
