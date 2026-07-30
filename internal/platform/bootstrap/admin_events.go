package bootstrap

import (
	"log/slog"

	billingEvents "skykin-platform/internal/billing/interfaces/events"
	billingInfra "skykin-platform/internal/billing/infrastructure"
	campaignConsumers "skykin-platform/internal/campaigns/consumers"
	"skykin-platform/internal/platform/messaging"

	"gorm.io/gorm"
)

// RegisterAdminEventConsumers wires async handlers for admin-emitted domain events.
// Segment candidate review and segment purchases are synchronous and transactional,
// so they are deliberately absent here.
func RegisterAdminEventConsumers(db *gorm.DB, bus *messaging.Bus, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}

	rateRepo := billingInfra.NewBillingRateRepository(db)
	billingEvents.NewPlanConsumer(rateRepo, logger).Register(bus)

	campaignConsumers.NewModerationConsumer(logger).Register(bus)
}
