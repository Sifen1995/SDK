package routes

import (
	billingHTTP "skykin-platform/internal/billing/interfaces/http"
	billingApp "skykin-platform/internal/billing/application"
	billingInfra "skykin-platform/internal/billing/infrastructure"
	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/platform/messaging"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module wires billing dependencies for the ad portal.
type Module struct {
	Handler      *billingHTTP.Handler
	SubEnforcer  *billingApp.SubscriptionEnforcer
	SubRepo      *billingInfra.SubscriptionRepository
	ChannelRepo  *billingInfra.ChannelRepository
	RateRepo     *billingInfra.BillingRateRepository
	PlanService  *billingApp.PlanService
	BillingAdmin *billingApp.BillingAdminService
}

// Wire constructs the billing module. Campaign quota reads are provided by the
// composition root so billing never imports campaign infrastructure.
func Wire(db *gorm.DB, bus *messaging.Bus, campaignQuota billingdomain.CampaignQuotaReader) *Module {
	subRepo := billingInfra.NewSubscriptionRepository(db)
	rateRepo := billingInfra.NewBillingRateRepository(db)
	channelRepo := billingInfra.NewChannelRepository(db)

	subEnforcer := billingApp.NewSubscriptionEnforcer(subRepo, channelRepo, campaignQuota)
	subscriptionSvc := billingApp.NewSubscriptionService(subRepo, channelRepo)
	planSvc := billingApp.NewPlanService(subRepo, bus)
	billingAdmin := billingApp.NewBillingAdminService(subRepo, rateRepo)

	return &Module{
		Handler:      billingHTTP.NewHandler(subscriptionSvc),
		SubEnforcer:  subEnforcer,
		SubRepo:      subRepo,
		ChannelRepo:  channelRepo,
		RateRepo:     rateRepo,
		PlanService:  planSvc,
		BillingAdmin: billingAdmin,
	}
}

// RegisterRead mounts read-only billing routes on the protected ad portal group.
func (m *Module) RegisterRead(g *gin.RouterGroup) {
	g.GET("/plans", m.Handler.ListPlans)
	g.GET("/channels", m.Handler.ListChannels)
	g.GET("/subscription", m.Handler.GetSubscription)
}

// RegisterWrite mounts billing mutation routes on the write-enabled group.
func (m *Module) RegisterWrite(g *gin.RouterGroup) {
	g.POST("/subscription", m.Handler.Subscribe)
}
