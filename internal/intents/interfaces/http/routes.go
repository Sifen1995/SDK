package http

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts intent SDK routes on the API group.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/intents/ingest-ad", h.IngestIntentAd)
}
