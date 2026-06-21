package route

import (
	"skykin-platform/configs"
	advertiserApp "skykin-platform/internal/ad_portal/application"
	advertiserInfra "skykin-platform/internal/ad_portal/infrastructure"
	advertiserHTTP "skykin-platform/internal/ad_portal/interfaces/http"
	adminApp "skykin-platform/internal/admin/application"
	adminHTTP "skykin-platform/internal/admin/interfaces/http"
	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsInfra "skykin-platform/internal/analytics/infrastructure"
	analyticsHTTP "skykin-platform/internal/analytics/interfaces/http"
	audienceApp "skykin-platform/internal/audience/application"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	audienceHTTP "skykin-platform/internal/audience/interfaces/http"
	authRoutes "skykin-platform/internal/auth/routes"
	billingApp "skykin-platform/internal/billing/application"
	billingInfra "skykin-platform/internal/billing/infrastructure"
	billingHTTP "skykin-platform/internal/billing/interfaces/http"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	campaignHTTP "skykin-platform/internal/campaigns/interfaces/http"
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
	subRepo := billingInfra.NewSubscriptionRepository(db)
	rateRepo := billingInfra.NewBillingRateRepository(db)
	channelRepo := billingInfra.NewChannelRepository(db)
	segmentRepo := audienceInfra.NewSegmentRepository(db)

	subEnforcer := billingApp.NewSubscriptionEnforcer(subRepo, channelRepo, campaignRepo)
	audiencePurchases := audienceApp.NewPurchaseService(segmentRepo)
	audienceList := audienceApp.NewListService(segmentRepo, subRepo)
	subscriptionSvc := billingApp.NewSubscriptionService(subRepo, channelRepo)

	campaignSvc := campaignApp.NewCampaignService(campaignRepo, subEnforcer, audiencePurchases, channelRepo)
	adminModeration := adminApp.NewCampaignModerationService(campaignRepo, channelRepo)
	adminCatalog := adminApp.NewPlanAndSegmentService(subRepo, rateRepo, segmentRepo)
	adminBilling := adminApp.NewBillingAdminService(subRepo, rateRepo)
	analyticsSvc := analyticsApp.NewService(analyticsInfra.NewRepository(db))

	advertiserHTTP.RegisterRoutes(r,
		advertiserHTTP.NewAuthHandler(advertiserApp.NewAuthService(adRepo, cfg)),
		campaignHTTP.NewHandler(campaignSvc),
		audienceHTTP.NewHandler(audienceList),
		billingHTTP.NewHandler(subscriptionSvc),
		adminHTTP.NewCampaignHandler(adminModeration),
		adminHTTP.NewCatalogHandler(adminCatalog, adminBilling),
		analyticsHTTP.NewHandler(analyticsSvc),
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
