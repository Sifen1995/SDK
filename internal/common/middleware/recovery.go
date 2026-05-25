package middleware

import (
	"fmt"
	"net/http"
	"skykin-platform/internal/common/response"
	"github.com/gin-gonic/gin"
)

// GlobalRecovery catches any unexpected system panics and maps them to a 500 error response
func GlobalRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Turn the interface panic into a string or error type safely
				errStr := fmt.Sprintf("%v", err)
				
				// Optional: Log the exact raw error trace locally here using your configs/logger.go
				// logger.Error("CRITICAL RUNTIME PANIC: " + errStr)

				// Use your standardized response structure to safely respond to the client
				response.Error(
					c, 
					http.StatusInternalServerError, 
					"An unexpected internal server error occurred", 
					errStr, // Hidden or shown depending on your environment profile (prod/dev)
				)
				
				// Stop execution immediately for this request context
				c.Abort()
			}
		}()
		
		// Continue processing the request down the chain if no panic occurs
		c.Next()
	}
}