package http

import (
	"skykin-platform/configs"
	audienceHTTP "skykin-platform/internal/audience/interfaces/http"
	campaignHTTP "skykin-platform/internal/campaigns/interfaces/http"
	platformMiddleware "skykin-platform/internal/platform/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts ad portal auth + campaign + audience routes under /api/v1/ad-portal.
func RegisterRoutes(r *gin.Engine, auth *AuthHandler, campaigns *campaignHTTP.Handler, audience *audienceHTTP.Handler, cfg *configs.Config) {
	g := r.Group("/api/v1/ad-portal")
	{
		g.POST("/register", auth.Register)
		g.POST("/login", auth.Login)

		protected := g.Group("/")
		protected.Use(platformMiddleware.AdPortalAuthMiddleware(cfg))
		{
			protected.GET("/me", platformMiddleware.RequirePortalRead(), auth.Me)

			protected.GET("/audience/segments", platformMiddleware.RequirePortalRead(), audience.ListSegments)

			protected.GET("/campaigns", platformMiddleware.RequirePortalRead(), campaigns.ListCampaigns)
			protected.GET("/campaigns/:id", platformMiddleware.RequirePortalRead(), campaigns.GetCampaign)
			protected.GET("/campaigns/:id/preview", platformMiddleware.RequirePortalRead(), campaigns.PreviewCampaign)

			write := protected.Group("/")
			write.Use(platformMiddleware.RequirePortalWrite())
			{
				write.POST("/campaigns", campaigns.CreateCampaign)
				write.POST("/campaigns/:id/activate", campaigns.ActivateCampaign)
			}

			admin := protected.Group("/admin")
			admin.Use(platformMiddleware.RequirePortalRoles("operator_admin"))
			{
				admin.POST("/users", auth.CreateUser)
			}
		}
	}
}
