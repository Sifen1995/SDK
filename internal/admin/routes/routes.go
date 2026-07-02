package routes

import (
	"log/slog"

	adminApp "skykin-platform/internal/admin/application"
	adminHTTP "skykin-platform/internal/admin/interfaces/http"
	adportalHTTP "skykin-platform/internal/ad_portal/interfaces/http"
	audienceApp "skykin-platform/internal/audience/application"
	billingApp "skykin-platform/internal/billing/application"
	campaignApp "skykin-platform/internal/campaigns/application"
	"skykin-platform/configs"
	"skykin-platform/internal/platform/bootstrap"
	"skykin-platform/internal/platform/messaging"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module wires admin operator dependencies for the ad portal.
type Module struct {
	Campaigns         *adminHTTP.CampaignHandler
	Catalog           *adminHTTP.CatalogHandler
	AnalyticsOps      *adminHTTP.AnalyticsHandler
	SegmentCandidates *adminHTTP.SegmentCandidateHandler
	Users             *adminHTTP.UsersHandler
}

// Wire constructs the admin module from application services owned by each bounded context.
func Wire(
	db *gorm.DB,
	cfg *configs.Config,
	jobs *bootstrap.IntentConsistencyJobs,
	bus *messaging.Bus,
	plans *billingApp.PlanService,
	billingAdmin *billingApp.BillingAdminService,
	segments *audienceApp.ListService,
	moderation *campaignApp.ModerationService,
) *Module {
	approveCandidate := adminApp.NewApproveCandidateUseCase(bus, slog.Default())
	rejectCandidate := adminApp.NewRejectCandidateUseCase(bus, slog.Default())
	getUsers := bootstrap.NewGetUsersWithIntentsUseCase(db, cfg, slog.Default())

	return &Module{
		Campaigns:         adminHTTP.NewCampaignHandler(moderation),
		Catalog:           adminHTTP.NewCatalogHandler(segments, billingAdmin, plans),
		AnalyticsOps:      adminHTTP.NewAnalyticsHandler(jobs.AnalyzeUC),
		SegmentCandidates: adminHTTP.NewSegmentCandidateHandler(approveCandidate, rejectCandidate),
		Users:             adminHTTP.NewUsersHandler(getUsers),
	}
}

// Register mounts operator admin routes on the admin group.
func (m *Module) Register(g *gin.RouterGroup, auth *adportalHTTP.AuthHandler) {
	g.POST("/users", auth.CreateUser)
	g.GET("/sdk-users", m.Users.ListUsers)
	g.GET("/plans", m.Catalog.ListPlans)
	g.POST("/plans", m.Catalog.CreatePlan)
	g.GET("/plans/:plan_id", m.Catalog.GetPlan)
	g.PATCH("/plans/:plan_id", m.Catalog.UpdatePlan)
	g.POST("/plans/:plan_id/suspend", m.Catalog.SuspendPlan)
	g.GET("/plans/:plan_id/billing-rates", m.Catalog.ListBillingRates)
	g.PATCH("/billing-rates/:id", m.Catalog.UpdateBillingRate)
	g.POST("/audience/segments", m.Catalog.CreateSegment)
	g.GET("/audience/segments", m.Catalog.ListSegments)
	g.GET("/audience/segments/:segment_id", m.Catalog.GetSegment)
	g.POST("/audience/segments/:segment_id/suspend", m.Catalog.SuspendSegment)
	g.POST("/audience/segment-candidates/:id/approve", m.SegmentCandidates.ApproveSegmentCandidate)
	g.POST("/audience/segment-candidates/:id/reject", m.SegmentCandidates.RejectSegmentCandidate)
	g.POST("/analytics/intent-consistency/run", m.AnalyticsOps.TriggerIntentConsistency)
	g.GET("/campaigns/pending", m.Campaigns.ListPendingCampaigns)
	g.POST("/campaigns/:id/validate", m.Campaigns.ValidateCampaign)
	g.POST("/campaigns/:id/activate", m.Campaigns.ActivateCampaign)
}
