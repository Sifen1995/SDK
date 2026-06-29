package routes

import (
	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsHTTP "skykin-platform/internal/analytics/interfaces/http"
	analyticsInfra "skykin-platform/internal/analytics/infrastructure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module wires analytics dependencies for the ad portal.
type Module struct {
	Handler *analyticsHTTP.Handler
}

// Wire constructs the analytics module.
func Wire(db *gorm.DB) *Module {
	svc := analyticsApp.NewService(analyticsInfra.NewRepository(db))
	return &Module{Handler: analyticsHTTP.NewHandler(svc)}
}

// RegisterAdmin mounts operator analytics dashboard routes.
func (m *Module) RegisterAdmin(g *gin.RouterGroup) {
	g.GET("/analytics/overview", m.Handler.Overview)
	g.GET("/analytics/campaigns", m.Handler.Campaigns)
	g.GET("/analytics/campaigns/:id", m.Handler.CampaignDetail)
	g.GET("/analytics/delivery", m.Handler.Delivery)
	g.GET("/analytics/revenue", m.Handler.Revenue)
	g.GET("/analytics/advertisers", m.Handler.Advertisers)
}
