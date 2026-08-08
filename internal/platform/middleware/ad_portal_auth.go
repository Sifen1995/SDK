package middleware

import (
	"net/http"

	"skykin-platform/configs"
	"skykin-platform/internal/ad_portal/application"
	"skykin-platform/internal/ad_portal/domain"

	"github.com/gin-gonic/gin"
)

func AdPortalAuthMiddleware(cfg *configs.Config, auth *application.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}
		tokenStr, ok := bearerTokenFromHeader(authHeader)
		if !ok {
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
		u, err := auth.Me(c.Request.Context(), claims.UserID)
		if err != nil || u == nil || !u.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}
		c.Set("portal_user_id", claims.UserID)
		c.Set("portal_role", claims.Role)
		// Company scope only for advertisers — analysts/admins must not inherit advertiser_id.
		if claims.Role == domain.RoleAdvertiser {
			c.Set("advertiser_id", u.AccountAdvertiserID())
		} else {
			c.Set("advertiser_id", "")
		}
		if claims.Role == domain.RoleReadOnlyAnalyst {
			c.Set("analyst_id", u.AccountAnalystID())
		} else {
			c.Set("analyst_id", "")
		}
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

// AccountAdvertiserID returns the company scope for campaign APIs (loaded from DB in auth middleware).
// Empty for operator_admin and read_only_analyst.
func AccountAdvertiserID(c *gin.Context) string {
	id, _ := c.Get("advertiser_id")
	s, _ := id.(string)
	return s
}

// AccountAnalystID returns the analyst profile id when the session is a read_only_analyst.
func AccountAnalystID(c *gin.Context) string {
	id, _ := c.Get("analyst_id")
	s, _ := id.(string)
	return s
}
