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
	permApp "skykin-platform/internal/permissions/application"
	permHTTP "skykin-platform/internal/permissions/interfaces/http"
	"skykin-platform/configs"
	"skykin-platform/internal/platform/bootstrap"
	"skykin-platform/internal/platform/messaging"
	platformMiddleware "skykin-platform/internal/platform/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register wires ad portal modules and mounts routes under /api/v1/ad-portal.
func Register(
	r *gin.Engine,
	db *gorm.DB,
	cfg *configs.Config,
	bus *messaging.Bus,
	checker *permApp.PermissionChecker,
	permHandler *permHTTP.Handler,
) *bootstrap.IntentConsistencyJobs {
	adRepo := advertiserInfra.NewRepository(db)
	authService := advertiserApp.NewAuthService(adRepo, cfg)
	authHandler := adportalHTTP.NewAuthHandler(authService)

	billingMod := billingRoutes.Wire(db, bus)
	audienceMod := audienceRoutes.Wire(db, billingMod.SubRepo)

	bootstrap.RegisterAdminEventConsumers(db, bus, slog.Default())
	jobs := bootstrap.SetupIntentConsistency(db, cfg, slog.Default())
	audienceMod.AttachCandidates(audienceApp.NewListSegmentCandidatesUseCase(jobs.CandidateRepo))

	campaignMod := campaignRoutes.Wire(db, billingMod.SubEnforcer, audienceMod.Purchases, billingMod.ChannelRepo, bus)
	adminMod := adminRoutes.Wire(
		db, cfg,
		jobs, bus,
		billingMod.PlanService,
		billingMod.BillingAdmin,
		audienceMod.Segments,
		campaignMod.Moderation,
	)
	analyticsMod := analyticsRoutes.Wire(db)

	g := r.Group("/api/v1/ad-portal")
	{
		g.POST("/register", authHandler.Register)
		g.POST("/login", authHandler.Login)

		protected := g.Group("/")
		protected.Use(platformMiddleware.AdPortalAuthMiddleware(cfg, authService))
		{
			protected.GET("/me", platformMiddleware.RequirePortalRead(), authHandler.Me)

			read := protected.Group("/")
			read.Use(platformMiddleware.RequirePortalRead())
			{
				billingMod.RegisterRead(read)
				audienceMod.RegisterRead(read)
				campaignMod.RegisterRead(read, checker)
			}

			write := protected.Group("/")
			write.Use(platformMiddleware.RequirePortalWrite())
			{
				billingMod.RegisterWrite(write)
				campaignMod.RegisterWrite(write, checker)
			}

			admin := protected.Group("/admin")
			admin.Use(platformMiddleware.RequirePortalRoles("operator_admin"))
			{
				adminMod.Register(admin, authHandler)
				analyticsMod.RegisterAdmin(admin, checker)
				audienceMod.RegisterAdmin(admin)
				permHTTP.RegisterRoutes(admin, permHandler)
			}
		}
	}

	return jobs
}
