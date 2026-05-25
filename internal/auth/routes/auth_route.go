package routes

import (
	"skykin-platform/internal/auth/controller"
	"skykin-platform/internal/common/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes handles the endpoint mappings for the auth domain module using a single controller
func RegisterAuthRoutes(r *gin.Engine, authCtrl *controller.AuthController) {

	// Create a portal-specific base group for developer actions
	portalGroup := r.Group("/api/v1/portal")
	{
		// ==========================================
		// PUBLIC ENDPOINTS
		// ==========================================

		// Open endpoint for new web portal platform signups
		portalGroup.POST("/register", authCtrl.RegisterDeveloper)
		portalGroup.POST("/login", authCtrl.LoginDeveloper)

		// ==========================================
		// PROTECTED PORTAL ENDPOINTS
		// ==========================================

		// Create a separate subgroup that strictly requires administrative developer authorization
		protectedPortal := portalGroup.Group("/")
		protectedPortal.Use(middleware.PortalAuthMiddleware())
		{
			// Endpoint to create apps, generating publishable/secret key pairs
			protectedPortal.POST("/applications", authCtrl.CreateApplication)
		}
	}
}
