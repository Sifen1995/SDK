package routes

import (
	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsHTTP "skykin-platform/internal/analytics/interfaces/http"
	analyticsInfra "skykin-platform/internal/analytics/infrastructure"
	permApp "skykin-platform/internal/permissions/application"
	platformMiddleware "skykin-platform/internal/platform/middleware"

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
func (m *Module) RegisterAdmin(g *gin.RouterGroup, checker *permApp.PermissionChecker) {
	readPerm := platformMiddleware.RequirePermission(checker, "analytics:read")
	g.GET("/analytics/overview", readPerm, m.Handler.Overview)
	g.GET("/analytics/campaigns", readPerm, m.Handler.Campaigns)
	g.GET("/analytics/campaigns/:id", readPerm, m.Handler.CampaignDetail)
	g.GET("/analytics/delivery", readPerm, m.Handler.Delivery)
	g.GET("/analytics/revenue", readPerm, m.Handler.Revenue)
	g.GET("/analytics/advertisers", readPerm, m.Handler.Advertisers)
}
