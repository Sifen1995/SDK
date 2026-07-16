package bootstrap

import (
	"log/slog"
	"strings"

	"skykin-platform/configs"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	intentApp "skykin-platform/internal/intents/application"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	intentHTTP "skykin-platform/internal/intents/interfaces/http"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

// NewIntentSystem wires the intents ingest + ad fetch flow (composition root).
// Campaign selection stays behind intents/application.AdSelector; no cross-repo access here.
func NewIntentSystem(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) *intentHTTP.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	intentRepo := intentsInfra.NewIntentRepository(db, cfg)
	profileRepo := intentsInfra.NewProfileRepository(intentRepo)

	var cache intentApp.ActiveIntentCache
	if addr := strings.TrimSpace(cfg.RedisAddr); addr != "" {
		if rdb, err := platformredis.NewRedisClient(addr); err == nil {
			cache = intentsInfra.NewIntentCacheAdapter(intentsInfra.NewRedisIntentRepository(rdb))
			logger.Info("intent cache: redis", "addr", addr)
		} else {
			logger.Warn("intent cache unavailable", "error", err)
		}
	}

	campaignRepo := campaignInfra.NewRepository(db)
	adSelector := campaignApp.NewIntentAdSelector(campaignRepo)

	svc := intentApp.NewIntentService(profileRepo, cache, adSelector)
	return intentHTTP.NewHandler(svc)
}
