package bootstrap

import (
	"log/slog"

	audienceApp "skykin-platform/internal/audience/application"
	audienceEvents "skykin-platform/internal/audience/interfaces/events"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	billingEvents "skykin-platform/internal/billing/interfaces/events"
	billingInfra "skykin-platform/internal/billing/infrastructure"
	campaignConsumers "skykin-platform/internal/campaigns/consumers"
	"skykin-platform/internal/platform/messaging"

	"gorm.io/gorm"
)

// RegisterAdminEventConsumers wires async handlers for admin-emitted domain events.
func RegisterAdminEventConsumers(db *gorm.DB, bus *messaging.Bus, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}

	candidateRepo := audienceInfra.NewCandidateRepository(db)
	segmentRepo := audienceInfra.NewSegmentRepository(db)
	membershipRepo := audienceInfra.NewMembershipRepository(db)
	purchaseRepo := audienceInfra.NewPurchaseRepository(db)

	processApproved := audienceApp.NewProcessApprovedCandidateUseCase(
		segmentRepo, membershipRepo, candidateRepo, logger,
	)
	rejectCandidate := audienceApp.NewRejectCandidateUseCase(candidateRepo, logger)
	recordPurchase := audienceApp.NewRecordSegmentPurchaseUseCase(purchaseRepo)

	audienceEvents.NewCandidateConsumer(processApproved, rejectCandidate, recordPurchase, logger).Register(bus)

	rateRepo := billingInfra.NewBillingRateRepository(db)
	billingEvents.NewPlanConsumer(rateRepo, logger).Register(bus)

	campaignConsumers.NewModerationConsumer(logger).Register(bus)
}
