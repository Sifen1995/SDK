package bootstrap

import (
	"log/slog"

	consentApp "skykin-platform/internal/consent/application"
	consentConsumers "skykin-platform/internal/consent/consumers"
	consentInfra "skykin-platform/internal/consent/infrastructure"
	consentHTTP "skykin-platform/internal/consent/interfaces/http"
	"skykin-platform/internal/platform/messaging"
	usersApp "skykin-platform/internal/users/application"
	usersConsumers "skykin-platform/internal/users/consumers"
	usersInfra "skykin-platform/internal/users/infrastructure"

	"gorm.io/gorm"
)

// NewConsentSystem wires the consent + users registration saga (composition root).
// Consent never writes to users; users never write to consents/mappings.
func NewConsentSystem(
	db *gorm.DB,
	bus *messaging.Bus,
	logger *slog.Logger,
) *consentHTTP.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	userRepo := usersInfra.NewUserRepository(db)
	consentRepo := consentInfra.NewConsentRepository(db)
	mappingRepo := consentInfra.NewPseudonymousMappingRepository(db)

	provisionUC := usersApp.NewProvisionUserForConsentUseCase(userRepo, bus, logger)
	completeUC := consentApp.NewCompleteConsentRegistrationUseCase(consentRepo, mappingRepo, bus, logger)

	usersConsumers.NewConsentRegistrationConsumer(provisionUC, logger).Register(bus)
	consentConsumers.NewUserProvisionedConsumer(completeUC, logger).Register(bus)

	createUC := consentApp.NewCreateConsentUseCase(bus, logger)
	return consentHTTP.NewHandler(createUC)
}
