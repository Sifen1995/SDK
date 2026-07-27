package http

import "github.com/gin-gonic/gin"

// RegisterSDKRoutes mounts anonymous campaign delivery and telemetry routes on the SDK API group.
func RegisterSDKRoutes(r *gin.RouterGroup, campaigns *CampaignHandler, telemetry *TelemetryHandler, cpc *CPCClickHandler) {
	if campaigns != nil {
		r.GET("/campaigns/anonymous", campaigns.ListAnonymousCampaigns)
	}
	if telemetry != nil {
		r.POST("/telemetry/track", telemetry.Track)
		r.POST("/telemetry/track-anonymous", telemetry.TrackAnonymous)
	}
	if cpc != nil {
		r.POST("/telemetry/anonymous-click", func(c *gin.Context) {
			cpc.TrackAnonymousClick(c.Writer, c.Request)
		})
	}
}
