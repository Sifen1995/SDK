package bootstrap

import (
	"log/slog"
	"strings"

	"skykin-platform/configs"
	billingWorker "skykin-platform/internal/billing/worker"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	deliveryHTTP "skykin-platform/internal/delivery/http"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

// DeliverySDKHandlers holds SDK delivery + telemetry HTTP handlers.
type DeliverySDKHandlers struct {
	Campaigns *deliveryHTTP.CampaignHandler
	Telemetry *deliveryHTTP.TelemetryHandler
}

// NewDeliverySDKSystem wires anonymous campaigns and telemetry track ingest.
func NewDeliverySDKSystem(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) *DeliverySDKHandlers {
	if logger == nil {
		logger = slog.Default()
	}

	campaignRepo := campaignInfra.NewRepository(db)

	var platformRDB *platformredis.RedisClient
	var redisCampaign *campaignInfra.RedisCampaignRepository
	if addr := strings.TrimSpace(cfg.RedisAddr); addr != "" {
		if rdb, err := platformredis.NewRedisClient(addr); err == nil {
			platformRDB = rdb
			redisCampaign = campaignInfra.NewRedisCampaignRepository(rdb)
			logger.Info("delivery sdk: redis enabled", "addr", addr)
		} else {
			logger.Warn("delivery sdk: redis unavailable", "error", err)
		}
	}

	cached := campaignInfra.NewCachedCampaignRepository(campaignRepo, redisCampaign, platformRDB)
	anonSvc := campaignApp.NewAnonymousCampaignService(cached)

	out := &DeliverySDKHandlers{
		Campaigns: deliveryHTTP.NewCampaignHandler(anonSvc),
	}
	if platformRDB != nil {
		out.Telemetry = deliveryHTTP.NewTelemetryHandler(platformRDB)
	} else {
		logger.Warn("delivery sdk: telemetry track disabled (redis required)")
	}
	return out
}

// StartBillingStreamWorker launches the Redis Streams write-behind billing consumer.
func StartBillingStreamWorker(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	addr := strings.TrimSpace(cfg.RedisAddr)
	if addr == "" {
		return
	}
	rdb, err := platformredis.NewRedisClient(addr)
	if err != nil {
		logger.Warn("billing stream worker: redis unavailable", "error", err)
		return
	}
	billingWorker.StartBillingConsumer(db, rdb, logger)
}
