package bootstrap

import (
	"context"
	"log/slog"
	"strings"

	"skykin-platform/configs"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	geoapp "skykin-platform/internal/geofencing/application"
	geoinfra "skykin-platform/internal/geofencing/infrastructure"
	geoHTTP "skykin-platform/internal/geofencing/interface/http"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

type campaignOwnerAdapter struct {
	repo *campaignInfra.Repository
}

func (a campaignOwnerAdapter) GetByID(ctx context.Context, id string) (*campaigndomain.Campaign, error) {
	return a.repo.Get(ctx, id)
}

// NewGeofencingSystem wires portal + SDK geofencing handlers.
func NewGeofencingSystem(
	db *gorm.DB,
	cfg *configs.Config,
	logger *slog.Logger,
) *geoHTTP.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	repo := geoinfra.NewGeofenceRepository(db)
	visits := geoinfra.NewStoreVisitRepository(db)
	campaigns := campaignInfra.NewRepository(db)

	var redisFreq *campaignInfra.RedisCampaignRepository
	if cfg != nil {
		if addr := strings.TrimSpace(cfg.RedisAddr); addr != "" {
			if rdb, err := platformredis.NewRedisClient(addr); err == nil {
				redisFreq = campaignInfra.NewRedisCampaignRepository(rdb)
				logger.Info("geofencing: redis frequency capping enabled", "addr", addr)
			} else {
				logger.Warn("geofencing: redis unavailable", "error", err)
			}
		}
	}

	svc := geoapp.NewGeofencingService(
		repo,
		repo,
		visits,
		repo,
		redisFreq,
		redisFreq,
		campaignOwnerAdapter{repo: campaigns},
	)
	return geoHTTP.NewHandler(svc)
}
