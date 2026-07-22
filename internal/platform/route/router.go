package route

import (
	"log/slog"

	"skykin-platform/configs"
	adportalRoutes "skykin-platform/internal/ad_portal/routes"
	authRoutes "skykin-platform/internal/auth/routes"
	consentHTTP "skykin-platform/internal/consent/interfaces/http"
	deliveryHTTP "skykin-platform/internal/delivery/http"
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
	bootstrap.RegisterDownstreamConsumers(db, cfg, bus)
	consentHandler := bootstrap.NewConsentSystem(db, bus, slog.Default())
	intentHandler := bootstrap.NewIntentSystem(db, cfg, slog.Default())
	deliverySDK := bootstrap.NewDeliverySDKSystem(db, cfg, slog.Default())

	sdkGroup := r.Group("/api/v1")
	sdkGroup.Use(sdkAuthMiddleware)
	{
		consentHTTP.RegisterRoutes(sdkGroup, consentHandler)
		intentHTTP.RegisterRoutes(sdkGroup, intentHandler)
		deliveryHTTP.RegisterSDKRoutes(sdkGroup, deliverySDK.Campaigns, deliverySDK.Telemetry)
	}

	return intentJobs
}
