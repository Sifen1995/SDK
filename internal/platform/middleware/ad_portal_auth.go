package middleware

import (
	"fmt"
	"net/http"

	"skykin-platform/configs"
	"skykin-platform/internal/advertisers/application"
	"skykin-platform/internal/advertisers/domain"

	"github.com/gin-gonic/gin"
)

func AdPortalAuthMiddleware(cfg *configs.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}
		var tokenStr string
		if _, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenStr); err != nil || tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "use Bearer <token>"})
			c.Abort()
			return
		}
		claims, err := application.ParsePortalToken(cfg, tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}
		c.Set("portal_user_id", claims.PortalUserID)
		c.Set("advertiser_id", claims.AdvertiserID)
		c.Set("portal_role", claims.Role)
		c.Set("portal_email", claims.Email)
		c.Next()
	}
}

func RequirePortalRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role, _ := c.Get("portal_role")
		roleStr, _ := role.(string)
		if _, ok := allowed[roleStr]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequirePortalWrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("portal_role")
		if !domain.CanWrite(role.(string)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "read-only access"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequirePortalRead() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("portal_role")
		if !domain.CanRead(role.(string)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AccountAdvertiserID returns the company scope for campaign APIs (from JWT advertiser_id).
func AccountAdvertiserID(c *gin.Context) string {
	id, _ := c.Get("advertiser_id")
	s, _ := id.(string)
	return s
}
