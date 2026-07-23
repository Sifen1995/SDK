package bootstrap

import (
	"log/slog"
	"strings"

	"skykin-platform/configs"
	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsInfra "skykin-platform/internal/analytics/infrastructure"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	deliveryInfra "skykin-platform/internal/delivery/infrastructure"
	intentApp "skykin-platform/internal/intents/application"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	intentHTTP "skykin-platform/internal/intents/interfaces/http"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

// NewIntentSystem wires the intents ingest + ad fetch + anonymous aggregate flow.
func NewIntentSystem(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) *intentHTTP.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	intentRepo := intentsInfra.NewIntentRepository(db, cfg)

	var platformRDB *platformredis.RedisClient
	var cache intentApp.ActiveIntentCache
	var logQueue *intentsInfra.IntentLogQueue
	var redisCampaign *campaignInfra.RedisCampaignRepository
	var aggregateIngest *analyticsApp.AggregateIngestService

	if addr := strings.TrimSpace(cfg.RedisAddr); addr != "" {
		if rdb, err := platformredis.NewRedisClient(addr); err == nil {
			platformRDB = rdb
			cache = intentsInfra.NewIntentCacheAdapter(intentsInfra.NewRedisIntentRepository(rdb))
			logQueue = intentsInfra.NewIntentLogQueue(rdb)
			redisCampaign = campaignInfra.NewRedisCampaignRepository(rdb)
			aggregateIngest = analyticsApp.NewAggregateIngestService(analyticsInfra.NewAnalyticsAggregateQueue(rdb))
			logger.Info("intent system: redis enabled", "addr", addr)
		} else {
			logger.Warn("intent system: redis unavailable", "error", err)
		}
	}

	profileRepo := intentsInfra.NewProfileRepository(intentRepo, logQueue)

	campaignRepo := campaignInfra.NewRepository(db)
	cachedCampaigns := campaignInfra.NewCachedCampaignRepository(campaignRepo, redisCampaign, platformRDB)
	deliveryJobs := deliveryInfra.NewDeliveryRepository(db)
	adSelector := campaignApp.NewIntentAdSelector(cachedCampaigns, campaignRepo, deliveryJobs, logger)

	svc := intentApp.NewIntentService(profileRepo, cache, adSelector)
	return intentHTTP.NewHandler(svc, aggregateIngest)
}

// StartIntentLogWorker launches the background BRPop → Postgres batch flusher.
func StartIntentLogWorker(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	addr := strings.TrimSpace(cfg.RedisAddr)
	if addr == "" {
		return
	}
	rdb, err := platformredis.NewRedisClient(addr)
	if err != nil {
		logger.Warn("intent log worker: redis unavailable", "error", err)
		return
	}
	intentsInfra.StartIntentLogWorker(db, rdb, logger)
}
