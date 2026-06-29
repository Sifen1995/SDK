package routes

import (
	"log/slog"

	adminApp "skykin-platform/internal/admin/application"
	adminHTTP "skykin-platform/internal/admin/interfaces/http"
	adportalHTTP "skykin-platform/internal/ad_portal/interfaces/http"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	billingApp "skykin-platform/internal/billing/application"
	billingInfra "skykin-platform/internal/billing/infrastructure"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	"skykin-platform/internal/platform/bootstrap"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module wires admin operator dependencies for the ad portal.
type Module struct {
	Campaigns         *adminHTTP.CampaignHandler
	Catalog           *adminHTTP.CatalogHandler
	AnalyticsOps      *adminHTTP.AnalyticsHandler
	SegmentCandidates *adminHTTP.SegmentCandidateHandler
}

// Wire constructs the admin module.
func Wire(db *gorm.DB, jobs *bootstrap.IntentConsistencyJobs) *Module {
	campaignRepo := campaignInfra.NewRepository(db)
	subRepo := billingInfra.NewSubscriptionRepository(db)
	rateRepo := billingInfra.NewBillingRateRepository(db)
	channelRepo := billingInfra.NewChannelRepository(db)
	segmentRepo := audienceInfra.NewSegmentRepository(db)
	membershipRepo := audienceInfra.NewMembershipRepository(db)

	adminModeration := adminApp.NewCampaignModerationService(campaignRepo, channelRepo)
	adminCatalog := adminApp.NewPlanAndSegmentService(subRepo, rateRepo, segmentRepo)
	adminBilling := adminApp.NewBillingAdminService(subRepo, rateRepo)
	planSvc := billingApp.NewPlanService(subRepo)
	approveCandidate := adminApp.NewApproveCandidateUseCase(jobs.CandidateRepo, membershipRepo, adminCatalog, slog.Default())
	rejectCandidate := adminApp.NewRejectCandidateUseCase(jobs.CandidateRepo, slog.Default())

	return &Module{
		Campaigns:         adminHTTP.NewCampaignHandler(adminModeration),
		Catalog:           adminHTTP.NewCatalogHandler(adminCatalog, adminBilling, planSvc),
		AnalyticsOps:      adminHTTP.NewAnalyticsHandler(jobs.AnalyzeUC),
		SegmentCandidates: adminHTTP.NewSegmentCandidateHandler(approveCandidate, rejectCandidate),
	}
}

// Register mounts operator admin routes on the admin group.
func (m *Module) Register(g *gin.RouterGroup, auth *adportalHTTP.AuthHandler) {
	g.POST("/users", auth.CreateUser)
	g.GET("/plans", m.Catalog.ListPlans)
	g.POST("/plans", m.Catalog.CreatePlan)
	g.GET("/plans/:plan_id", m.Catalog.GetPlan)
	g.PATCH("/plans/:plan_id", m.Catalog.UpdatePlan)
	g.POST("/plans/:plan_id/suspend", m.Catalog.SuspendPlan)
	g.GET("/plans/:plan_id/billing-rates", m.Catalog.ListBillingRates)
	g.PATCH("/billing-rates/:id", m.Catalog.UpdateBillingRate)
	g.POST("/audience/segments", m.Catalog.CreateSegment)
	g.GET("/audience/segments", m.Catalog.ListSegments)
	g.POST("/audience/segment-candidates/:id/approve", m.SegmentCandidates.ApproveSegmentCandidate)
	g.POST("/audience/segment-candidates/:id/reject", m.SegmentCandidates.RejectSegmentCandidate)
	g.POST("/analytics/intent-consistency/run", m.AnalyticsOps.TriggerIntentConsistency)
	g.GET("/campaigns/pending", m.Campaigns.ListPendingCampaigns)
	g.POST("/campaigns/:id/validate", m.Campaigns.ValidateCampaign)
	g.POST("/campaigns/:id/activate", m.Campaigns.ActivateCampaign)
}
