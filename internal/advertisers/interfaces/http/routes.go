package http

import (
	"skykin-platform/configs"
	adminHTTP "skykin-platform/internal/admin/interfaces/http"
	audienceHTTP "skykin-platform/internal/audience/interfaces/http"
	campaignHTTP "skykin-platform/internal/campaigns/interfaces/http"
	platformMiddleware "skykin-platform/internal/platform/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts ad portal auth + campaign + audience + admin routes under /api/v1/ad-portal.
func RegisterRoutes(r *gin.Engine, auth *AuthHandler, campaigns *campaignHTTP.Handler, audience *audienceHTTP.Handler, adminCampaigns *adminHTTP.CampaignHandler, cfg *configs.Config) {
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
			}

			admin := protected.Group("/admin")
			admin.Use(platformMiddleware.RequirePortalRoles("operator_admin"))
			{
				admin.POST("/users", auth.CreateUser)
				admin.GET("/campaigns/pending", adminCampaigns.ListPendingCampaigns)
				admin.POST("/campaigns/:id/validate", adminCampaigns.ValidateCampaign)
				admin.POST("/campaigns/:id/activate", adminCampaigns.ActivateCampaign)
			}
		}
	}
}
