package http

import (
	"skykin-platform/configs"
	adminHTTP "skykin-platform/internal/admin/interfaces/http"
	analyticsHTTP "skykin-platform/internal/analytics/interfaces/http"
	audienceHTTP "skykin-platform/internal/audience/interfaces/http"
	billingHTTP "skykin-platform/internal/billing/interfaces/http"
	campaignHTTP "skykin-platform/internal/campaigns/interfaces/http"
	platformMiddleware "skykin-platform/internal/platform/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts ad portal auth + campaign + audience + billing + admin routes under /api/v1/ad-portal.
func RegisterRoutes(r *gin.Engine, auth *AuthHandler, campaigns *campaignHTTP.Handler, audience *audienceHTTP.Handler, billing *billingHTTP.Handler, adminCampaigns *adminHTTP.CampaignHandler, adminCatalog *adminHTTP.CatalogHandler, analytics *analyticsHTTP.Handler, cfg *configs.Config) {
	g := r.Group("/api/v1/ad-portal")
	{
		g.POST("/register", auth.Register)
		g.POST("/login", auth.Login)

		protected := g.Group("/")
		protected.Use(platformMiddleware.AdPortalAuthMiddleware(cfg))
		{
			protected.GET("/me", platformMiddleware.RequirePortalRead(), auth.Me)

			protected.GET("/plans", platformMiddleware.RequirePortalRead(), billing.ListPlans)
			protected.GET("/channels", platformMiddleware.RequirePortalRead(), billing.ListChannels)
			protected.GET("/subscription", platformMiddleware.RequirePortalRead(), billing.GetSubscription)

			protected.GET("/audience/segments", platformMiddleware.RequirePortalRead(), audience.ListSegments)

			protected.GET("/campaigns", platformMiddleware.RequirePortalRead(), campaigns.ListCampaigns)
			protected.GET("/campaigns/:id", platformMiddleware.RequirePortalRead(), campaigns.GetCampaign)
			protected.GET("/campaigns/:id/preview", platformMiddleware.RequirePortalRead(), campaigns.PreviewCampaign)

			write := protected.Group("/")
			write.Use(platformMiddleware.RequirePortalWrite())
			{
				write.POST("/subscription", billing.Subscribe)
				write.POST("/campaigns", campaigns.CreateCampaign)
			}

			admin := protected.Group("/admin")
			admin.Use(platformMiddleware.RequirePortalRoles("operator_admin"))
			{
				admin.POST("/users", auth.CreateUser)
				admin.POST("/plans", adminCatalog.CreatePlan)
				admin.GET("/plans/:plan_id/billing-rates", adminCatalog.ListBillingRates)
				admin.PATCH("/billing-rates/:id", adminCatalog.UpdateBillingRate)
				admin.POST("/audience/segments", adminCatalog.CreateSegment)
				admin.GET("/campaigns/pending", adminCampaigns.ListPendingCampaigns)
				admin.POST("/campaigns/:id/validate", adminCampaigns.ValidateCampaign)
				admin.POST("/campaigns/:id/activate", adminCampaigns.ActivateCampaign)

				admin.GET("/analytics/overview", analytics.Overview)
				admin.GET("/analytics/campaigns", analytics.Campaigns)
				admin.GET("/analytics/campaigns/:id", analytics.CampaignDetail)
				admin.GET("/analytics/delivery", analytics.Delivery)
				admin.GET("/analytics/revenue", analytics.Revenue)
				admin.GET("/analytics/advertisers", analytics.Advertisers)
			}
		}
	}
}
