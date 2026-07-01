package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"skykin-platform/configs"
	"skykin-platform/internal/auth/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func SDKAuthMiddleware(authRepo repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		pubKeyPlain := c.GetHeader("X-API-Key")
		if pubKeyPlain == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-API-Key header"})
			c.Abort()
			return
		}

		hashedPubKey := sha256Hash(pubKeyPlain)
		apiKeyRecord, appRecord, err := authRepo.VerifyAPIKey(c.Request.Context(), hashedPubKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or suspended api credentials"})
			c.Abort()
			return
		}

		signature := c.GetHeader("X-Signature")
		if signature != "" || c.Request.Method == http.MethodPost {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read payload body"})
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			mac := hmac.New(sha256.New, []byte(apiKeyRecord.SecretKeyValue))
			mac.Write(bodyBytes)
			expectedSignature := hex.EncodeToString(mac.Sum(nil))
			if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "cryptographic payload signature mismatch"})
				c.Abort()
				return
			}
		}

		c.Set("application_id", appRecord.ID.String())
		c.Next()
	}
}

func PortalAuthMiddleware(cfg *configs.Config) gin.HandlerFunc {
	var jwtSecret = []byte(cfg.JwtSecret)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Missing authorization header context"})
			c.Abort()
			return
		}

		var tokenStr string
		var ok bool
		if tokenStr, ok = bearerTokenFromHeader(authHeader); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid Authorization header format. Use 'Bearer <token>'"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected token signing algorithm method: %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			log.Printf("[Auth Debug] Token validation failed structural check: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Session expired or invalid token authentication"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			devID, exists := claims["developer_id"].(string)
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Malformed token session payload tracking metrics"})
				c.Abort()
				return
			}
			c.Set("developer_id", devID)
		}

		c.Next()
	}
}

func sha256Hash(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

