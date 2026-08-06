package route

import (
	"log/slog"

	"skykin-platform/configs"
	adportalRoutes "skykin-platform/internal/ad_portal/routes"
	authRoutes "skykin-platform/internal/auth/routes"
	consentHTTP "skykin-platform/internal/consent/interfaces/http"
	deliveryHTTP "skykin-platform/internal/delivery/http"
	fraudHTTP "skykin-platform/internal/fraud/interfaces/http"
	geoHTTP "skykin-platform/internal/geofencing/interface/http"
	intentHTTP "skykin-platform/internal/intents/interfaces/http"
	permApp "skykin-platform/internal/permissions/application"
	permHTTP "skykin-platform/internal/permissions/interfaces/http"
	"skykin-platform/internal/platform/bootstrap"
	"skykin-platform/internal/platform/messaging"
	platformMiddleware "skykin-platform/internal/platform/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRouter(
	r *gin.Engine,
	db *gorm.DB,
	cfg *configs.Config,
	bus *messaging.Bus,
	checker *permApp.PermissionChecker,
	permHandler *permHTTP.Handler,
) *bootstrap.IntentConsistencyJobs {
	r.Use(platformMiddleware.CORS())
	r.Use(gin.Logger())
	r.Use(platformMiddleware.GlobalRecovery())

	sdkAuthMiddleware := authRoutes.RegisterRoutes(r, db, cfg)
	intentJobs := adportalRoutes.Register(r, db, cfg, bus, checker, permHandler)

	// Event ingestion HTTP + bus publishing intentionally not mounted.
	// Package internal/events is retained for later reactivation.
	consentHandler := bootstrap.NewConsentSystem(db, bus, slog.Default())
	deliverySDK := bootstrap.NewDeliverySDKSystem(db, cfg, slog.Default())
	intentHandler := bootstrap.NewIntentSystem(db, cfg, slog.Default(), deliverySDK.SMSDispatch)
	fraudHandler := bootstrap.NewFraudSystem(db, cfg, slog.Default())
	geofenceHandler := bootstrap.NewGeofencingSystem(db, cfg, slog.Default())
	bootstrap.RegisterDeliveryEventConsumers(bus, deliverySDK.SMSDispatch, slog.Default())

	// Stream write-behind: billing owns billing_events, delivery owns campaign_delivery_logs.
	bootstrap.StartBillingStreamWorker(db, cfg, slog.Default())
	bootstrap.StartDeliveryLogStreamWorker(db, cfg, slog.Default())
	bootstrap.StartAnalyticsAggregateWorker(db, cfg, slog.Default())
	bootstrap.StartIntentLogWorker(db, cfg, slog.Default())

	sdkGroup := r.Group("/api/v1")
	deliveryHTTP.RegisterPublicRoutes(r, deliverySDK.SMSClick, deliverySDK.Twilio)
	sdkGroup.Use(sdkAuthMiddleware)
	{
		consentHTTP.RegisterRoutes(sdkGroup, consentHandler)
		intentHTTP.RegisterRoutes(sdkGroup, intentHandler)
		fraudHTTP.RegisterRoutes(sdkGroup, fraudHandler)
		geoHTTP.RegisterSDK(sdkGroup, geofenceHandler)
		deliveryHTTP.RegisterSDKRoutes(sdkGroup, deliverySDK.Campaigns, deliverySDK.Telemetry, deliverySDK.CPC, deliverySDK.SMSDebug)
	}

	return intentJobs
}
