package routes

import (
	"log/slog"

	advertiserApp "skykin-platform/internal/ad_portal/application"
	advertiserInfra "skykin-platform/internal/ad_portal/infrastructure"
	adportalHTTP "skykin-platform/internal/ad_portal/interfaces/http"
	adminRoutes "skykin-platform/internal/admin/routes"
	analyticsRoutes "skykin-platform/internal/analytics/routes"
	audienceApp "skykin-platform/internal/audience/application"
	audienceRoutes "skykin-platform/internal/audience/routes"
	billingRoutes "skykin-platform/internal/billing/routes"
	campaignRoutes "skykin-platform/internal/campaigns/routes"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	"skykin-platform/configs"
	"skykin-platform/internal/platform/bootstrap"
	"skykin-platform/internal/platform/messaging"
	platformMiddleware "skykin-platform/internal/platform/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register wires ad portal modules and mounts routes under /api/v1/ad-portal.
func Register(r *gin.Engine, db *gorm.DB, cfg *configs.Config, bus *messaging.Bus) *bootstrap.IntentConsistencyJobs {
	adRepo := advertiserInfra.NewRepository(db)
	authHandler := adportalHTTP.NewAuthHandler(advertiserApp.NewAuthService(adRepo, cfg))

	billingMod := billingRoutes.Wire(db)
	segmentRepo := audienceInfra.NewSegmentRepository(db)
	audiencePurchases := audienceApp.NewPurchaseService(segmentRepo)
	audienceMod := audienceRoutes.Wire(db, billingMod.SubRepo)

	jobs := bootstrap.SetupIntentConsistency(db, cfg, bus, slog.Default())
	audienceMod.AttachCandidates(audienceApp.NewListSegmentCandidatesUseCase(jobs.CandidateRepo))

	campaignMod := campaignRoutes.Wire(db, billingMod.SubEnforcer, audiencePurchases, billingMod.ChannelRepo)
	adminMod := adminRoutes.Wire(db, jobs)
	analyticsMod := analyticsRoutes.Wire(db)

	g := r.Group("/api/v1/ad-portal")
	{
		g.POST("/register", authHandler.Register)
		g.POST("/login", authHandler.Login)

		protected := g.Group("/")
		protected.Use(platformMiddleware.AdPortalAuthMiddleware(cfg))
		{
			protected.GET("/me", platformMiddleware.RequirePortalRead(), authHandler.Me)

			read := protected.Group("/")
			read.Use(platformMiddleware.RequirePortalRead())
			{
				billingMod.RegisterRead(read)
				audienceMod.RegisterRead(read)
				campaignMod.RegisterRead(read)
			}

			write := protected.Group("/")
			write.Use(platformMiddleware.RequirePortalWrite())
			{
				billingMod.RegisterWrite(write)
				campaignMod.RegisterWrite(write)
			}

			admin := protected.Group("/admin")
			admin.Use(platformMiddleware.RequirePortalRoles("operator_admin"))
			{
				adminMod.Register(admin, authHandler)
				analyticsMod.RegisterAdmin(admin)
				audienceMod.RegisterAdmin(admin)
			}
		}
	}

	return jobs
}
