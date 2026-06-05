package bootstrap

import (
	"log/slog"
	"strings"

	"skykin-platform/configs"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignConsumers "skykin-platform/internal/campaigns/consumers"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	deliveryConsumers "skykin-platform/internal/delivery/consumers"
	eventsInfra "skykin-platform/internal/events/infrastructure"
	intentsApp "skykin-platform/internal/intents/application"
	intentsConsumers "skykin-platform/internal/intents/consumers"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	"skykin-platform/internal/platform/messaging"
	platformredis "skykin-platform/internal/platform/redis"
	platformWS "skykin-platform/internal/platform/websocket"
	rewardsInfra "skykin-platform/internal/rewards/infrastructure"
	usersInfra "skykin-platform/internal/users/infrastructure"
	wsConsumers "skykin-platform/internal/websocket/consumers"

	"gorm.io/gorm"
)

// Downstream holds async consumers and shared prediction use case.
type Downstream struct {
	Predict *intentsApp.PredictIntentUseCase
}

// RegisterDownstreamConsumers wires intent prediction worker and websocket consumers.
func RegisterDownstreamConsumers(db *gorm.DB, cfg *configs.Config, bus *messaging.Bus, hub *platformWS.Hub) *Downstream {
	eventRepo := eventsInfra.NewPostgresRepository(db)
	userRepo := usersInfra.NewUserRepository(db)
	intentRepo := intentsInfra.NewIntentRepository(db, cfg)
	rewardRepo := rewardsInfra.NewRewardRepository(db)
	campaignRepo := campaignInfra.NewRepository(db)
	adDelivery := campaignApp.NewAdDeliveryService(campaignRepo)

	mlURL := strings.TrimSpace(cfg.MLServiceURL)
	if mlURL == "" {
		mlURL = "http://localhost:8000"
	}
	mlClient := intentsInfra.NewMLClient(strings.TrimSuffix(mlURL, "/"))

	var redisClient *platformredis.RedisClient
	if addr := strings.TrimSpace(cfg.RedisAddr); addr != "" {
		if c, err := platformredis.NewRedisClient(addr); err == nil {
			redisClient = c
		} else {
			slog.Warn("redis unavailable for prediction pipeline", "error", err)
		}
	}

	predictUC := intentsApp.NewPredictIntentUseCase(
		eventRepo,
		userRepo,
		mlClient,
		intentRepo,
		rewardRepo,
		adDelivery,
		redisClient,
		bus,
		slog.Default(),
	)

	intentsConsumers.NewPredictIntentConsumer(predictUC, redisClient).Register(bus)
	wsConsumers.NewRewardCreatedConsumer(bus, hub).Register()
	campaignConsumers.NewCampaignAdConsumer(hub).Register(bus)
	deliveryConsumers.NewDispatchConsumer(db, adDelivery, hub).Register(bus)

	return &Downstream{Predict: predictUC}
}
