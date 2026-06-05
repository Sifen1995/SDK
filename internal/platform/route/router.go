package route

import (
	"skykin-platform/configs"
	advertiserApp "skykin-platform/internal/advertisers/application"
	advertiserHTTP "skykin-platform/internal/advertisers/interfaces/http"
	advertiserInfra "skykin-platform/internal/advertisers/infrastructure"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignHTTP "skykin-platform/internal/campaigns/interfaces/http"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	authRoutes "skykin-platform/internal/auth/routes"
	eventHTTP "skykin-platform/internal/events/interfaces/http"
	intentHTTP "skykin-platform/internal/intents/interfaces/http"
	"skykin-platform/internal/platform/bootstrap"
	"skykin-platform/internal/platform/messaging"
	platformMiddleware "skykin-platform/internal/platform/middleware"
	platformWS "skykin-platform/internal/platform/websocket"
	wsRoutes "skykin-platform/internal/websocket/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRouter(r *gin.Engine, db *gorm.DB, cfg *configs.Config, hub *platformWS.Hub, bus *messaging.Bus) {
	r.Use(platformMiddleware.CORS())
	r.Use(gin.Logger())
	r.Use(platformMiddleware.GlobalRecovery())

	sdkAuthMiddleware := authRoutes.RegisterRoutes(r, db, cfg)

	adRepo := advertiserInfra.NewRepository(db)
	campaignRepo := campaignInfra.NewRepository(db)
	advertiserHTTP.RegisterRoutes(r,
		advertiserHTTP.NewAuthHandler(advertiserApp.NewAuthService(adRepo, cfg)),
		campaignHTTP.NewHandler(campaignApp.NewCampaignService(campaignRepo)),
		cfg,
	)

	eventsModule := eventHTTP.NewModule(db, cfg, bus)
	downstream := bootstrap.RegisterDownstreamConsumers(db, cfg, eventsModule.Bus, hub)
	intentHandler := intentHTTP.NewHandler(downstream.Predict)

	sdkGroup := r.Group("/api/v1")
	sdkGroup.Use(sdkAuthMiddleware)
	{
		eventHTTP.RegisterRoutes(sdkGroup, eventsModule)
		intentHTTP.RegisterRoutes(sdkGroup, intentHandler)
		wsRoutes.RegisterRoutes(sdkGroup, hub)
	}
}
