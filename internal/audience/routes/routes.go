package routes

import (
	audienceApp "skykin-platform/internal/audience/application"
	audienceHTTP "skykin-platform/internal/audience/interfaces/http"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	billingInfra "skykin-platform/internal/billing/infrastructure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module wires audience dependencies for the ad portal.
type Module struct {
	listSvc *audienceApp.ListService
	Handler *audienceHTTP.Handler
}

// Wire constructs the audience module.
func Wire(db *gorm.DB, subRepo *billingInfra.SubscriptionRepository) *Module {
	segmentRepo := audienceInfra.NewSegmentRepository(db)
	listSvc := audienceApp.NewListService(segmentRepo, subRepo)
	return &Module{
		listSvc: listSvc,
		Handler: audienceHTTP.NewHandler(listSvc, nil),
	}
}

// AttachCandidates wires the segment-candidate list use case after analytics bootstrap.
func (m *Module) AttachCandidates(uc *audienceApp.ListSegmentCandidatesUseCase) {
	m.Handler = audienceHTTP.NewHandler(m.listSvc, uc)
}

// RegisterRead mounts advertiser-facing audience routes.
func (m *Module) RegisterRead(g *gin.RouterGroup) {
	g.GET("/audience/segments", m.Handler.ListSegments)
}

// RegisterAdmin mounts operator admin audience routes.
func (m *Module) RegisterAdmin(g *gin.RouterGroup) {
	g.GET("/audience/segment-candidates", m.Handler.ListSegmentCandidates)
}
