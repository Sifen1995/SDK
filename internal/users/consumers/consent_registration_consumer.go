package consumers

import (
	"log/slog"

	consentDomain "skykin-platform/internal/consent/domain"
	"skykin-platform/internal/platform/messaging"
	"skykin-platform/internal/users/application"
)

// ConsentRegistrationConsumer reacts to ConsentRegistrationRequested by
// provisioning a user in the users bounded context.
type ConsentRegistrationConsumer struct {
	uc     *application.ProvisionUserForConsentUseCase
	logger *slog.Logger
}

func NewConsentRegistrationConsumer(
	uc *application.ProvisionUserForConsentUseCase,
	logger *slog.Logger,
) *ConsentRegistrationConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &ConsentRegistrationConsumer{uc: uc, logger: logger}
}

func (c *ConsentRegistrationConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, consentDomain.EventConsentRegistrationRequested, c.handle)
}

func (c *ConsentRegistrationConsumer) handle(e messaging.Event) {
	payload, ok := e.Payload.(consentDomain.ConsentRegistrationRequestedPayload)
	if !ok {
		c.logger.Error("invalid ConsentRegistrationRequested payload")
		return
	}
	if err := c.uc.Execute(e.Ctx, payload.PseudonymousID, payload.ConsentLevel, payload.SDKVersion); err != nil {
		c.logger.Error("provision user for consent failed",
			"pseudonymous_id", payload.PseudonymousID,
			"error", err,
		)
	}
}
