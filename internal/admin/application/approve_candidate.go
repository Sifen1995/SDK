package application

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	adminEvents "skykin-platform/internal/admin/events"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
)

// ApproveCandidateUseCase publishes a candidate approval for async audience processing.
type ApproveCandidateUseCase struct {
	bus    *messaging.Bus
	logger *slog.Logger
}

func NewApproveCandidateUseCase(bus *messaging.Bus, logger *slog.Logger) *ApproveCandidateUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ApproveCandidateUseCase{bus: bus, logger: logger}
}

func (uc *ApproveCandidateUseCase) Execute(
	ctx context.Context,
	candidateID uuid.UUID,
	adminID uuid.UUID,
	name, description string,
	estimatedCPM float64,
) error {
	if uc.bus == nil {
		return errors.New("event bus not configured")
	}
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if estimatedCPM <= 0 {
		return errors.New("estimated_cpm must be > 0")
	}

	uc.bus.Publish(messaging.Event{
		Name: adminEvents.TopicCandidateApproved,
		Ctx:  ctx,
		Payload: adminEvents.CandidateApprovedEvent{
			CandidateID:  candidateID,
			AdminID:      adminID,
			Name:         name,
			Description:  description,
			EstimatedCPM: estimatedCPM,
		},
	})

	uc.logger.Info("candidate approval requested", "candidate_id", candidateID)
	return nil
}
