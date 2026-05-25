package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware is a placeholder for API rate limiting.
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Placeholder logic for rate limiting
		c.Next()
	}
}

// Return429 Aborts request with status 429
func Return429(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error":   "rate_limit_exceeded",
		"message": "Too many requests. Please try again later.",
	})
}
