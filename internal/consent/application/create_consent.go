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
	"gorm.io/gorm"
)

// CreateConsentCommand is the input for consent registration from the Flutter SDK.
type CreateConsentCommand struct {
	ConsentLevel string
	SMSConsented bool
	SDKVersion   string
}

// CreateConsentResult is returned to the HTTP layer after a successful create.
type CreateConsentResult struct {
	PseudonymousID string
	ConsentLevel   string
	SMSConsented   bool
}

type demoPseudonymousLookup interface {
	FindOneDemoPseudonymousID(ctx context.Context) (string, error)
}

// CreateConsentUseCase starts the consent registration saga via the event bus.
// When SMSConsented is true it returns an existing demo user's pseudonymous_id
// instead of provisioning a new user (Flutter SMS+ demo path).
type CreateConsentUseCase struct {
	bus interface {
		PublishSync(event messaging.Event)
	}
	demo   demoPseudonymousLookup
	logger *slog.Logger
}

func NewCreateConsentUseCase(
	bus interface {
		PublishSync(event messaging.Event)
	},
	demo demoPseudonymousLookup,
	logger *slog.Logger,
) *CreateConsentUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &CreateConsentUseCase{bus: bus, demo: demo, logger: logger}
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

	if cmd.SMSConsented {
		return uc.returnDemoConsent(ctx, cmd)
	}

	pseudonymousID := uuid.New()

	uc.bus.PublishSync(messaging.Event{
		Name: domain.EventConsentRegistrationRequested,
		Payload: domain.ConsentRegistrationRequestedPayload{
			PseudonymousID: pseudonymousID.String(),
			ConsentLevel:   cmd.ConsentLevel,
			SMSConsented:   false,
			SDKVersion:     cmd.SDKVersion,
		},
		Ctx: ctx,
	})

	uc.logger.Info("consent registration requested",
		"pseudonymous_id", pseudonymousID.String(),
		"consent_level", cmd.ConsentLevel,
		"sms_consented", false,
	)

	return &CreateConsentResult{
		PseudonymousID: pseudonymousID.String(),
		ConsentLevel:   cmd.ConsentLevel,
		SMSConsented:   false,
	}, nil
}

func (uc *CreateConsentUseCase) returnDemoConsent(
	ctx context.Context,
	cmd CreateConsentCommand,
) (*CreateConsentResult, error) {
	if uc.demo == nil {
		return nil, errors.New("demo consent lookup is not configured")
	}
	pseudoID, err := uc.demo.FindOneDemoPseudonymousID(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no demo users available for sms_consented consent")
		}
		return nil, err
	}

	uc.logger.Info("consent returned existing demo mapping",
		"pseudonymous_id", pseudoID,
		"consent_level", cmd.ConsentLevel,
		"sms_consented", true,
	)

	return &CreateConsentResult{
		PseudonymousID: pseudoID,
		ConsentLevel:   cmd.ConsentLevel,
		SMSConsented:   true,
	}, nil
}
