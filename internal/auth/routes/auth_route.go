package routes

import (
	"skykin-platform/configs"
	"skykin-platform/internal/auth/controller"
	"skykin-platform/internal/auth/repository"
	"skykin-platform/internal/auth/service"
	"skykin-platform/internal/common/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes wires up the auth module and mounts its endpoints.
// Returns the SDK auth middleware so the central router can protect other groups.
func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *configs.Config) gin.HandlerFunc {
	repo := repository.NewAuthRepository(db, cfg)
	serv := service.NewAuthService(repo, cfg)
	ctrl := controller.NewAuthController(serv)

	portalGroup := r.Group("/api/v1/portal")
	{
		portalGroup.POST("/register", ctrl.RegisterDeveloper)
		portalGroup.POST("/login", ctrl.LoginDeveloper)

		protectedPortal := portalGroup.Group("/")
		protectedPortal.Use(middleware.PortalAuthMiddleware(cfg))
		{
			protectedPortal.POST("/applications", ctrl.CreateApplication)
			protectedPortal.GET("/applications", ctrl.GetApplications)
		}
	}

	return middleware.SDKAuthMiddleware(repo)
}
