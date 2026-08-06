package routes

import (
	audienceApp "skykin-platform/internal/audience/application"
	billingApp "skykin-platform/internal/billing/application"
	billingdomain "skykin-platform/internal/billing/domain"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	campaignHTTP "skykin-platform/internal/campaigns/interfaces/http"
	geoinfra "skykin-platform/internal/geofencing/infrastructure"
	permApp "skykin-platform/internal/permissions/application"
	"skykin-platform/internal/platform/messaging"
	platformMiddleware "skykin-platform/internal/platform/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module wires campaign dependencies for the ad portal.
type Module struct {
	Handler    *campaignHTTP.Handler
	Moderation *campaignApp.ModerationService
}

// Wire constructs the campaigns module.
func Wire(
	db *gorm.DB,
	subEnforcer *billingApp.SubscriptionEnforcer,
	audiencePurchases *audienceApp.PurchaseService,
	channels billingdomain.ChannelRepository,
	bus *messaging.Bus,
) *Module {
	repo := campaignInfra.NewRepository(db)
	svc := campaignApp.NewCampaignService(repo, subEnforcer, audiencePurchases, channels, bus)
	moderation := campaignApp.NewModerationService(
		repo,
		channels,
		bus,
		geoinfra.NewGeofenceRepository(db),
	)
	return &Module{
		Handler:    campaignHTTP.NewHandler(svc),
		Moderation: moderation,
	}
}

// RegisterRead mounts read-only campaign routes.
func (m *Module) RegisterRead(g *gin.RouterGroup, checker *permApp.PermissionChecker) {
	readPerm := platformMiddleware.RequirePermission(checker, "campaigns:read")
	g.GET("/campaigns", readPerm, m.Handler.ListCampaigns)
	g.GET("/campaigns/:id", m.Handler.GetCampaign)
	g.GET("/campaigns/:id/preview", m.Handler.PreviewCampaign)
}

// RegisterWrite mounts campaign mutation routes.
func (m *Module) RegisterWrite(g *gin.RouterGroup, checker *permApp.PermissionChecker) {
	createPerm := platformMiddleware.RequirePermission(checker, "campaigns:create")
	g.POST("/campaigns", createPerm, m.Handler.CreateCampaign)
}
