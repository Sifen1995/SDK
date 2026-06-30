package routes

import (
	audienceApp "skykin-platform/internal/audience/application"
	billingApp "skykin-platform/internal/billing/application"
	billingdomain "skykin-platform/internal/billing/domain"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignHTTP "skykin-platform/internal/campaigns/interfaces/http"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	"skykin-platform/internal/platform/messaging"

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
	moderation := campaignApp.NewModerationService(repo, channels, bus)
	return &Module{
		Handler:    campaignHTTP.NewHandler(svc),
		Moderation: moderation,
	}
}

// RegisterRead mounts read-only campaign routes.
func (m *Module) RegisterRead(g *gin.RouterGroup) {
	g.GET("/campaigns", m.Handler.ListCampaigns)
	g.GET("/campaigns/:id", m.Handler.GetCampaign)
	g.GET("/campaigns/:id/preview", m.Handler.PreviewCampaign)
}

// RegisterWrite mounts campaign mutation routes.
func (m *Module) RegisterWrite(g *gin.RouterGroup) {
	g.POST("/campaigns", m.Handler.CreateCampaign)
}
