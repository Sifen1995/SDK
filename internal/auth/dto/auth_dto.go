package dto

import (
	"time"

	"github.com/gin-gonic/gin"
)

// DeveloperRegisterRequest handles incoming portal registration data
type DeveloperRegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// ApplicationCreateRequest isolates platform settings for the SDK environment
type ApplicationCreateRequest struct {
	AppName  string `json:"app_name" binding:"required,min=2,max=100"`
	Platform string `json:"platform" binding:"required"`  // e.g., "flutter", "ios"
	BundleID string `json:"bundle_id" binding:"required"` // e.g., "com.company.app"
}

// ApplicationResponse returns details back to the portal interface
type ApplicationResponse struct {
	ID        string    `json:"id"`
	AppName   string    `json:"app_name"`
	Platform  string    `json:"platform"`
	BundleID  string    `json:"bundle_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// APIKeyCredentialResponse shows the raw keys ONCE during creation
type APIKeyCredentialResponse struct {
	ApplicationID  string `json:"application_id"`
	PublishableKey string `json:"publishable_key"` // Sent plain text in mobile apps (X-API-Key)
	RawSecretKey   string `json:"secret_key"`      // Shown ONCE, used to sign payloads via HMAC
	RateLimit      int    `json:"rate_limit"`
}

// DeveloperLoginRequest handles incoming portal credential validation check payloads
type DeveloperLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse returns the signed access token and developer details
type LoginResponse struct {
	Token     string `json:"token"`
	Developer gin.H  `json:"developer"`
}
