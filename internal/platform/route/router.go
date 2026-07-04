package route

import (
	"skykin-platform/configs"
	adportalRoutes "skykin-platform/internal/ad_portal/routes"
	authRoutes "skykin-platform/internal/auth/routes"
	eventHTTP "skykin-platform/internal/events/interfaces/http"
	intentRoutes "skykin-platform/internal/intents/routes"
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

	eventsModule := eventHTTP.NewModule(db, cfg, bus)
	downstream := bootstrap.RegisterDownstreamConsumers(db, cfg, eventsModule.Bus, hub)
	intentModule := intentRoutes.Wire(downstream.Predict)

	sdkGroup := r.Group("/api/v1")
	sdkGroup.Use(sdkAuthMiddleware)
	{
		eventHTTP.RegisterRoutes(sdkGroup, eventsModule)
		intentModule.Register(sdkGroup)
		wsRoutes.RegisterRoutes(sdkGroup, hub)
	}

	return intentJobs
}
