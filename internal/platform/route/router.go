package route

import (
	"log/slog"

	"skykin-platform/configs"
	adportalRoutes "skykin-platform/internal/ad_portal/routes"
	authRoutes "skykin-platform/internal/auth/routes"
	consentHTTP "skykin-platform/internal/consent/interfaces/http"
	intentHTTP "skykin-platform/internal/intents/interfaces/http"
	permApp "skykin-platform/internal/permissions/application"
	permHTTP "skykin-platform/internal/permissions/interfaces/http"
	"skykin-platform/internal/platform/bootstrap"
	"skykin-platform/internal/platform/messaging"
	platformMiddleware "skykin-platform/internal/platform/middleware"
	platformWS "skykin-platform/internal/platform/websocket"
	wsRoutes "skykin-platform/internal/websocket/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRouter(
	r *gin.Engine,
	db *gorm.DB,
	cfg *configs.Config,
	hub *platformWS.Hub,
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
	bootstrap.RegisterDownstreamConsumers(db, cfg, bus, hub)
	consentHandler := bootstrap.NewConsentSystem(db, bus, slog.Default())
	intentHandler := bootstrap.NewIntentSystem(db, cfg, slog.Default())

	sdkGroup := r.Group("/api/v1")
	sdkGroup.Use(sdkAuthMiddleware)
	{
		consentHTTP.RegisterRoutes(sdkGroup, consentHandler)
		intentHTTP.RegisterRoutes(sdkGroup, intentHandler)
		wsRoutes.RegisterRoutes(sdkGroup, hub)
	}

	return intentJobs
}
