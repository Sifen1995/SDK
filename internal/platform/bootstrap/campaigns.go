package bootstrap

import (
	"context"
	"log/slog"
	"time"

	campaignApp "skykin-platform/internal/campaigns/application"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	deliveryInfra "skykin-platform/internal/delivery/infrastructure"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	"skykin-platform/configs"
	"skykin-platform/internal/platform/messaging"

	"gorm.io/gorm"
)

func StartTargetingJob(db *gorm.DB, bus *messaging.Bus, logger *slog.Logger, interval time.Duration) {
	cfg := &configs.Config{}
	campaignRepo := campaignInfra.NewRepository(db)
	intentRepo := intentsInfra.NewIntentRepository(db, cfg)
	deliveryRepo := deliveryInfra.NewDeliveryRepository(db)
	job := campaignApp.NewTargetingJob(campaignRepo, intentRepo, deliveryRepo, bus, logger)

	job.Run(context.Background())
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			logger.Info("targeting job tick")
			job.Run(context.Background())
		}
	}()
}
