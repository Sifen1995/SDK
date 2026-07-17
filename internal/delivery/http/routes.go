package http

import "github.com/gin-gonic/gin"

// RegisterSDKRoutes mounts anonymous campaign delivery routes on the SDK API group.
func RegisterSDKRoutes(r *gin.RouterGroup, h *CampaignHandler) {
	if h == nil {
		return
	}
	r.GET("/campaigns/anonymous", h.ListAnonymousCampaigns)
}
