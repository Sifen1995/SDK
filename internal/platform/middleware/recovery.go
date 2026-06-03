package middleware

import (
	"fmt"
	"net/http"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

func GlobalRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				errStr := fmt.Sprintf("%v", err)
				platformHTTP.Error(
					c,
					http.StatusInternalServerError,
					"An unexpected internal server error occurred",
					errStr,
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}


