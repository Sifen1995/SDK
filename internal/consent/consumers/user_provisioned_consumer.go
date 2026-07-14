package consumers

import (
	"log/slog"

	"skykin-platform/internal/consent/application"
	"skykin-platform/internal/platform/messaging"
	usersDomain "skykin-platform/internal/users/domain"
)

// UserProvisionedConsumer completes consent registration after users publishes
// UserProvisionedForConsent.
type UserProvisionedConsumer struct {
	uc     *application.CompleteConsentRegistrationUseCase
	logger *slog.Logger
}

func NewUserProvisionedConsumer(
	uc *application.CompleteConsentRegistrationUseCase,
	logger *slog.Logger,
) *UserProvisionedConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &UserProvisionedConsumer{uc: uc, logger: logger}
}

func (c *UserProvisionedConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, usersDomain.EventUserProvisionedForConsent, c.handle)
}

func (c *UserProvisionedConsumer) handle(e messaging.Event) {
	payload, ok := e.Payload.(usersDomain.UserProvisionedForConsentPayload)
	if !ok {
		c.logger.Error("invalid UserProvisionedForConsent payload")
		return
	}
	if err := c.uc.Execute(
		e.Ctx,
		payload.UserID,
		payload.PseudonymousID,
		payload.ConsentLevel,
		payload.SDKVersion,
	); err != nil {
		c.logger.Error("complete consent registration failed",
			"user_id", payload.UserID,
			"pseudonymous_id", payload.PseudonymousID,
			"error", err,
		)
	}
}
