package middleware

import (
	"net/http"

	permApp "skykin-platform/internal/permissions/application"

	"github.com/gin-gonic/gin"
)

func RequirePermission(checker *permApp.PermissionChecker, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, ok := c.Get("portal_role")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		role, _ := roleVal.(string)
		if role == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		if !checker.HasPermission(c.Request.Context(), role, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "required": permission})
			c.Abort()
			return
		}
		c.Next()
	}
}
