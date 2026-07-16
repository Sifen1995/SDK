package bootstrap

import (
	"log/slog"

	"skykin-platform/configs"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	campaignEvents "skykin-platform/internal/campaigns/interfaces/events"
	"skykin-platform/internal/platform/messaging"
	rewardsConsumers "skykin-platform/internal/rewards/consumers"
	rewardsInfra "skykin-platform/internal/rewards/infrastructure"

	"gorm.io/gorm"
)

// RegisterDownstreamConsumers wires campaign delivery logging and reward consumers.
// SDK event ingestion and server-side intent prediction from events are not wired;
// that code remains under internal/events and internal/intents for later reuse.
func RegisterDownstreamConsumers(db *gorm.DB, cfg *configs.Config, bus *messaging.Bus) {
	_ = cfg

	rewardRepo := rewardsInfra.NewRewardRepository(db)
	campaignRepo := campaignInfra.NewRepository(db)
	adDelivery := campaignApp.NewAdDeliveryService(campaignRepo)

	campaignEvents.NewDeliveryConsumer(adDelivery, slog.Default()).Register(bus)
	rewardsConsumers.NewIntentRewardConsumer(rewardRepo, bus, slog.Default()).Register(bus)
}
