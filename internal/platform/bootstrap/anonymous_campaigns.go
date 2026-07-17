package bootstrap

import (
	"log/slog"
	"strings"

	"skykin-platform/configs"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	deliveryHTTP "skykin-platform/internal/delivery/http"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

// NewAnonymousCampaignSystem wires the non-consented campaign master-list endpoint.
func NewAnonymousCampaignSystem(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) *deliveryHTTP.CampaignHandler {
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
			logger.Info("anonymous campaigns: redis enabled", "addr", addr)
		} else {
			logger.Warn("anonymous campaigns: redis unavailable", "error", err)
		}
	}

	cached := campaignInfra.NewCachedCampaignRepository(campaignRepo, redisCampaign, platformRDB)
	svc := campaignApp.NewAnonymousCampaignService(cached)
	return deliveryHTTP.NewCampaignHandler(svc)
}
