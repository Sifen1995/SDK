package application

import (
	"context"
	"log/slog"

	"skykin-platform/internal/platform/messaging"
	"skykin-platform/internal/users/domain"
)

type userWriter interface {
	Create(ctx context.Context, user *domain.User) error
}

// ProvisionUserForConsentUseCase creates a users row and continues the consent saga.
type ProvisionUserForConsentUseCase struct {
	users userWriter
	bus   interface {
		PublishSync(event messaging.Event)
	}
	logger *slog.Logger
}

func NewProvisionUserForConsentUseCase(
	users userWriter,
	bus interface {
		PublishSync(event messaging.Event)
	},
	logger *slog.Logger,
) *ProvisionUserForConsentUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProvisionUserForConsentUseCase{users: users, bus: bus, logger: logger}
}

func (uc *ProvisionUserForConsentUseCase) Execute(
	ctx context.Context,
	pseudonymousID, consentLevel, sdkVersion string,
) error {
	user := &domain.User{}
	if err := uc.users.Create(ctx, user); err != nil {
		return err
	}

	uc.bus.PublishSync(messaging.Event{
		Name: domain.EventUserProvisionedForConsent,
		Payload: domain.UserProvisionedForConsentPayload{
			UserID:         user.ID,
			PseudonymousID: pseudonymousID,
			ConsentLevel:   consentLevel,
			SDKVersion:     sdkVersion,
		},
		Ctx: ctx,
	})

	uc.logger.Info("user provisioned for consent",
		"user_id", user.ID,
		"pseudonymous_id", pseudonymousID,
		"consent_level", consentLevel,
	)
	return nil
}
