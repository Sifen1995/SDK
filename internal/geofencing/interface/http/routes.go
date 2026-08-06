package http

import (
	permApp "skykin-platform/internal/permissions/application"
	platformMiddleware "skykin-platform/internal/platform/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterSDK mounts SDK geofencing routes on the authenticated /api/v1 group.
func RegisterSDK(r *gin.RouterGroup, h *Handler) {
	r.PATCH("/geofences/location-consent", h.SetLocationConsent)
	r.GET("/geofences/sync", h.Sync)
	r.POST("/geofence/event", h.Event)
}

// RegisterPortalRead mounts advertiser geofence read routes.
func RegisterPortalRead(r *gin.RouterGroup, h *Handler, checker *permApp.PermissionChecker) {
	manage := platformMiddleware.RequirePermission(checker, "geofences:manage")
	r.GET("/geofences", manage, h.ListZones)
	r.GET("/campaigns/:id/geofences", manage, h.ListCampaignZones)
}

// RegisterPortalWrite mounts advertiser geofence mutation routes.
func RegisterPortalWrite(r *gin.RouterGroup, h *Handler, checker *permApp.PermissionChecker) {
	manage := platformMiddleware.RequirePermission(checker, "geofences:manage")
	r.POST("/geofences", manage, h.CreateZone)
	r.POST("/campaigns/:id/geofences", manage, h.LinkCampaignZones)
}

// RegisterAdmin mounts operator geofence approval routes.
func RegisterAdmin(r *gin.RouterGroup, h *Handler) {
	r.GET("/geofences/pending", h.ListPendingZones)
	r.POST("/geofences/:id/activate", h.ActivateZone)
	r.POST("/campaigns/:id/geofences/activate", h.ActivateCampaignZones)
}

// Register is kept for backward compatibility with the placeholder signature.
func Register(_ *gin.RouterGroup) {}
