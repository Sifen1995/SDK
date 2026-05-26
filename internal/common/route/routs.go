package route

import (
	"skykin-platform/configs"
	authRoutes "skykin-platform/internal/auth/routes"
	"skykin-platform/internal/common/middleware"
	"skykin-platform/internal/common/websocket"
	eventRoutes "skykin-platform/internal/events/routes"
	wsRoutes "skykin-platform/internal/websocket/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRouter(r *gin.Engine, db *gorm.DB, cfg *configs.Config, hub *websocket.Hub) {
	r.Use(middleware.CORS())
	r.Use(gin.Logger())
	r.Use(middleware.GlobalRecovery())

	sdkAuthMiddleware := authRoutes.RegisterRoutes(r, db, cfg)

	sdkGroup := r.Group("/api/v1")
	sdkGroup.Use(sdkAuthMiddleware)
	{
		eventRoutes.RegisterRoutes(sdkGroup, db, cfg, hub)
		wsRoutes.RegisterRoutes(sdkGroup, hub)
	}
}
