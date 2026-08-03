package http

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts fraud intelligence routes on the authenticated SDK group.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/sync", h.Sync)
	r.POST("/reports", h.Report)
}
