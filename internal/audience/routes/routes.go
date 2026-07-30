package routes

import (
	audienceApp "skykin-platform/internal/audience/application"
	audienceHTTP "skykin-platform/internal/audience/interfaces/http"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	billingdomain "skykin-platform/internal/billing/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module wires audience dependencies for the ad portal.
type Module struct {
	Segments  *audienceApp.ListService
	Purchases *audienceApp.PurchaseService
	Handler   *audienceHTTP.Handler
}

// Wire constructs the audience module.
func Wire(db *gorm.DB, subRepo billingdomain.SubscriptionRepository) *Module {
	segmentRepo := audienceInfra.NewSegmentRepository(db)
	purchaseRepo := audienceInfra.NewPurchaseRepository(db)
	segments := audienceApp.NewListService(segmentRepo, subRepo)
	return &Module{
		Segments:  segments,
		Purchases: audienceApp.NewPurchaseService(segmentRepo, purchaseRepo),
		Handler:   audienceHTTP.NewHandler(segments, nil),
	}
}

// AttachCandidates wires the segment-candidate list use case after analytics bootstrap.
func (m *Module) AttachCandidates(uc *audienceApp.ListSegmentCandidatesUseCase) {
	m.Handler = audienceHTTP.NewHandler(m.Segments, uc)
}

// RegisterRead mounts advertiser-facing audience routes.
func (m *Module) RegisterRead(g *gin.RouterGroup) {
	g.GET("/audience/segments", m.Handler.ListSegments)
	g.GET("/audience/segments/:segment_id", m.Handler.GetSegment)
}

// RegisterAdmin mounts operator admin audience routes.
func (m *Module) RegisterAdmin(g *gin.RouterGroup) {
	g.GET("/audience/segment-candidates", m.Handler.ListSegmentCandidates)
}
