package route

import (
	"skykin-platform/configs"
	authController "skykin-platform/internal/auth/controller"
	authRepository "skykin-platform/internal/auth/repository"
	authRoutes "skykin-platform/internal/auth/routes"
	authService "skykin-platform/internal/auth/service"
	"skykin-platform/internal/common/middleware"
	"skykin-platform/internal/common/websocket"
	eventRoutes "skykin-platform/internal/events/routes"
	wsRoutes "skykin-platform/internal/websocket/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitRouter coordinates all module routes away from main.go
func InitRouter(r *gin.Engine, db *gorm.DB, cfg *configs.Config, hub *websocket.Hub) {

	// 1. Initialize Global Interceptors
	r.Use(middleware.GlobalRecovery())

	// 2. Initialize Core Infrastructure & Shared Middleware
	authRepo := authRepository.NewAuthRepository(db, cfg)
	authServ := authService.NewAuthService(authRepo, cfg)

	sdkAuthMiddleware := middleware.SDKAuthMiddleware(authRepo)

	// 3. Initialize Domain Controllers (Single unified instance)
	authCtrl := authController.NewAuthController(authServ)

	// ==========================================
	// DOMAIN ROUTE REGISTRATION REGIONS
	// ==========================================

	// Mount Portal Auth Engine Module (Passing only the unified controller)
	authRoutes.RegisterAuthRoutes(r, authCtrl)

	// Mount Protected SDK Ingestion Stream Group (/api/v1/...)
	sdkGroup := r.Group("/api/v1")
	sdkGroup.Use(sdkAuthMiddleware)
	{
		// Event domain routes
		eventRoutes.RegisterRoutes(sdkGroup, db, cfg, hub)

		// WebSocket domain routes
		wsRoutes.RegisterRoutes(sdkGroup, hub)
	}
}
