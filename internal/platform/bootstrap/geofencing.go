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
	intentdomain "skykin-platform/internal/intents/domain"
	intentInfra "skykin-platform/internal/intents/infrastructure"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

type campaignOwnerAdapter struct {
	repo *campaignInfra.Repository
}

func (a campaignOwnerAdapter) GetByID(ctx context.Context, id string) (*campaigndomain.Campaign, error) {
	return a.repo.Get(ctx, id)
}

// activeIntentAdapter resolves Redis user_intent first, then latest intents DB row.
type activeIntentAdapter struct {
	redis *intentInfra.RedisIntentRepository
	repo  intentdomain.IntentRepository
}

func (a activeIntentAdapter) CurrentIntent(ctx context.Context, pseudonymousID string) (string, error) {
	if a.redis != nil {
		name, err := a.redis.GetUserIntent(ctx, pseudonymousID)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(name) != "" {
			return name, nil
		}
	}
	if a.repo == nil || pseudonymousID == "" {
		return "", nil
	}
	latest, err := a.repo.FindLatestByPseudonymousIDs(ctx, []string{pseudonymousID})
	if err != nil {
		return "", err
	}
	if intent, ok := latest[pseudonymousID]; ok && intent != nil {
		return intent.IntentName, nil
	}
	return "", nil
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
	intentRepo := intentInfra.NewIntentRepository(db, cfg)

	var redisFreq *campaignInfra.RedisCampaignRepository
	var redisIntents *intentInfra.RedisIntentRepository
	if cfg != nil {
		if addr := strings.TrimSpace(cfg.RedisAddr); addr != "" {
			if rdb, err := platformredis.NewRedisClient(addr); err == nil {
				redisFreq = campaignInfra.NewRedisCampaignRepository(rdb)
				redisIntents = intentInfra.NewRedisIntentRepository(rdb)
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
		activeIntentAdapter{redis: redisIntents, repo: intentRepo},
	)
	return geoHTTP.NewHandler(svc)
}
