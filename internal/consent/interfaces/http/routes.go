package http

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts consent routes on the SDK API group.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/consent", h.CreateConsent)
}
