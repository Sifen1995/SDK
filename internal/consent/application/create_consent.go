package application

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"skykin-platform/internal/consent/domain"
	"skykin-platform/internal/consent/validation"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
)

// CreateConsentCommand is the input for consent registration from the Flutter SDK.
type CreateConsentCommand struct {
	ConsentLevel string
	SDKVersion   string
}

// CreateConsentResult is returned to the HTTP layer after a successful create.
type CreateConsentResult struct {
	PseudonymousID string
	ConsentLevel   string
}

// CreateConsentUseCase starts the consent registration saga via the event bus.
// It does not touch the users repository — users provision asynchronously in their BC.
type CreateConsentUseCase struct {
	bus interface {
		PublishSync(event messaging.Event)
	}
	logger *slog.Logger
}

func NewCreateConsentUseCase(
	bus interface {
		PublishSync(event messaging.Event)
	},
	logger *slog.Logger,
) *CreateConsentUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &CreateConsentUseCase{bus: bus, logger: logger}
}

func (uc *CreateConsentUseCase) Execute(
	ctx context.Context,
	cmd CreateConsentCommand,
) (*CreateConsentResult, error) {
	cmd.ConsentLevel = strings.TrimSpace(strings.ToLower(cmd.ConsentLevel))
	cmd.SDKVersion = strings.TrimSpace(cmd.SDKVersion)

	if cmd.SDKVersion == "" {
		return nil, errors.New("sdk_version is required")
	}
	if !validation.ValidateConsentLevel(cmd.ConsentLevel) {
		return nil, errors.New("invalid consent level: " + cmd.ConsentLevel)
	}

	pseudonymousID := uuid.New()

	uc.bus.PublishSync(messaging.Event{
		Name: domain.EventConsentRegistrationRequested,
		Payload: domain.ConsentRegistrationRequestedPayload{
			PseudonymousID: pseudonymousID.String(),
			ConsentLevel:   cmd.ConsentLevel,
			SDKVersion:     cmd.SDKVersion,
		},
		Ctx: ctx,
	})

	uc.logger.Info("consent registration requested",
		"pseudonymous_id", pseudonymousID.String(),
		"consent_level", cmd.ConsentLevel,
	)

	return &CreateConsentResult{
		PseudonymousID: pseudonymousID.String(),
		ConsentLevel:   cmd.ConsentLevel,
	}, nil
}
