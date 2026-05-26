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

// Helper function to hash keys back to their DB footprint format
func sha256Hash(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func SDKAuthMiddleware(authRepo repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract the plaintext Publishable Key from the header
		pubKeyPlain := c.GetHeader("X-API-Key")
		if pubKeyPlain == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-API-Key header"})
			c.Abort()
			return
		}

		// 2. Hash it to match how it's stored in the database
		hashedPubKey := sha256Hash(pubKeyPlain)

		// 3. Look up the key state and parent application info in one database pass
		apiKeyRecord, appRecord, err := authRepo.VerifyAPIKey(c.Request.Context(), hashedPubKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or suspended api credentials"})
			c.Abort()
			return
		}

		// 4. CHECK FOR HMAC SIGNATURE (For secure backend data transfers)
		signature := c.GetHeader("X-Signature")
		if signature != "" || c.Request.Method == http.MethodPost {
			// Read the request body safely without erasing it for the Controller later
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read payload body"})
				c.Abort()
				return
			}
			// Put the body back so the next handler can read it normally
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Calculate what the signature *should* be using the stored Hashed Secret Key
			mac := hmac.New(sha256.New, []byte(apiKeyRecord.SecretKeyValue))
			mac.Write(bodyBytes)
			expectedSignature := hex.EncodeToString(mac.Sum(nil))

			// Check if the calculated signature matches what the client passed in header
			if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "cryptographic payload signature mismatch"})
				c.Abort()
				return
			}
		}

		// 5. Attach the verified Application ID context to Gin so controllers can use it
		c.Set("application_id", appRecord.ID.String())
		c.Next()
	}
}

// PortalAuthMiddleware protects administrative developer dashboard actions using signed JWTs
func PortalAuthMiddleware(cfg *configs.Config) gin.HandlerFunc {
	var jwtSecret = []byte(cfg.JwtSecret)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Missing authorization header context"})
			c.Abort()
			return
		}

		// Look for standard bearer token formatting prefix structure: "Bearer <token_string>"
		var tokenStr string
		_, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenStr)
		if err != nil || tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid Authorization header format. Use 'Bearer <token>'"})
			c.Abort()
			return
		}

		// Parse and validate the token string
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// Ensure the token method matches HS256 algorithm
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

		// Extract the Custom Developer Claim payload data map structures safely
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			devID, exists := claims["developer_id"].(string)
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Malformed token session payload tracking metrics"})
				c.Abort()
				return
			}

			// Set the parsed developer UUID directly into Gin's running context lifecycle
			c.Set("developer_id", devID)
		}

		c.Next()
	}
}
