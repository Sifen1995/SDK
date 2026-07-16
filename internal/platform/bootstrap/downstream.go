package bootstrap

import (
	"log/slog"

	campaignApp "skykin-platform/internal/campaigns/application"
	campaignConsumers "skykin-platform/internal/campaigns/consumers"
	campaignEvents "skykin-platform/internal/campaigns/interfaces/events"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	deliveryConsumers "skykin-platform/internal/delivery/consumers"
	"skykin-platform/configs"
	"skykin-platform/internal/platform/messaging"
	platformWS "skykin-platform/internal/platform/websocket"
	rewardsConsumers "skykin-platform/internal/rewards/consumers"
	rewardsInfra "skykin-platform/internal/rewards/infrastructure"
	wsConsumers "skykin-platform/internal/websocket/consumers"

	"gorm.io/gorm"
)

// RegisterDownstreamConsumers wires campaign / reward / websocket consumers.
// SDK event ingestion and server-side intent prediction from events are not wired;
// that code remains under internal/events and internal/intents for later reuse.
func RegisterDownstreamConsumers(db *gorm.DB, cfg *configs.Config, bus *messaging.Bus, hub *platformWS.Hub) {
	_ = cfg

	rewardRepo := rewardsInfra.NewRewardRepository(db)
	campaignRepo := campaignInfra.NewRepository(db)
	adDelivery := campaignApp.NewAdDeliveryService(campaignRepo)

	wsConsumers.NewIntentPredictedConsumer(hub).Register(bus)
	wsConsumers.NewRewardCreatedConsumer(bus, hub).Register()
	campaignConsumers.NewCampaignAdConsumer(hub).Register(bus)
	campaignEvents.NewDeliveryConsumer(adDelivery, slog.Default()).Register(bus)
	rewardsConsumers.NewIntentRewardConsumer(rewardRepo, bus, slog.Default()).Register(bus)
	deliveryConsumers.NewDispatchConsumer(db, adDelivery, hub).Register(bus)
}
