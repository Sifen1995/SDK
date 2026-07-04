package http

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts permission management routes on an existing admin router group.
func RegisterRoutes(admin *gin.RouterGroup, h *Handler) {
	if h == nil {
		return
	}
	h.RegisterRoutes(admin)
}
