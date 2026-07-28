package bootstrap

import (
	"log/slog"
	"strings"

	"skykin-platform/configs"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	campaignEvents "skykin-platform/internal/campaigns/interfaces/events"
	"skykin-platform/internal/platform/messaging"

	"gorm.io/gorm"
)

// RegisterDownstreamConsumers wires campaign delivery logging.
// Intent-reward bus consumers are not used in the current SDK ingest-ad / TFLite path.
func RegisterDownstreamConsumers(db *gorm.DB, cfg *configs.Config, bus *messaging.Bus) {
	_ = cfg

	campaignRepo := campaignInfra.NewRepository(db)
	linkBuilder := campaignInfra.NewPlayLinkBuilder(strings.TrimSpace(cfg.ClickTokenSecret))
	adDelivery := campaignApp.NewAdDeliveryService(campaignRepo, linkBuilder)

	campaignEvents.NewDeliveryConsumer(adDelivery, slog.Default()).Register(bus)
}
