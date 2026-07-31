package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"skykin-platform/configs"
	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsInfra "skykin-platform/internal/analytics/infrastructure"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	deliveryApp "skykin-platform/internal/delivery/application"
	intentApp "skykin-platform/internal/intents/application"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	intentHTTP "skykin-platform/internal/intents/interfaces/http"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

// NewIntentSystem wires the intents ingest + ad fetch + anonymous aggregate flow.
// smsDispatch may be nil when SMS click secret is unset; SMS_PLUS then fails on dispatch.
func NewIntentSystem(
	db *gorm.DB,
	cfg *configs.Config,
	logger *slog.Logger,
	smsDispatch *deliveryApp.SMSDispatchService,
) *intentHTTP.Handler {
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
	adSelector := campaignApp.NewIntentAdSelector(cachedCampaigns)

	var smsPort intentApp.SMSAdDispatcher
	if smsDispatch != nil {
		smsPort = &smsAdDispatcherAdapter{svc: smsDispatch}
	}

	svc := intentApp.NewIntentService(profileRepo, cache, adSelector, smsPort)
	return intentHTTP.NewHandler(svc, aggregateIngest)
}

type smsAdDispatcherAdapter struct {
	svc *deliveryApp.SMSDispatchService
}

func (a *smsAdDispatcherAdapter) Dispatch(
	ctx context.Context,
	campaign *campaigndomain.Campaign,
	pseudonymousID string,
) error {
	if a == nil || a.svc == nil {
		return fmt.Errorf("sms dispatch is not configured")
	}
	if campaign == nil {
		return fmt.Errorf("campaign is required")
	}
	return a.svc.DispatchSMSCampaign(ctx, &deliveryApp.SMSCampaign{
		ID:             campaign.ID,
		Title:          campaign.Title,
		BodyText:       campaign.BodyText,
		DestinationURL: campaign.DestinationURL,
	}, pseudonymousID)
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
