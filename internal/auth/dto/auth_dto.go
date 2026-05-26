package dto

import (
	"time"
)

type DeveloperRegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100" example:"Jane Developer"`
	Email    string `json:"email" binding:"required,email" example:"jane@company.com"`
	Password string `json:"password" binding:"required,min=8" example:"securepass123"`
}

type ApplicationCreateRequest struct {
	AppName  string `json:"app_name" binding:"required,min=2,max=100" example:"My Shopping App"`
	Platform string `json:"platform" binding:"required" example:"flutter" enums:"flutter,android,ios,web"`
	BundleID string `json:"bundle_id" binding:"required" example:"com.company.app"`
}

type ApplicationResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	AppName   string    `json:"app_name" example:"My Shopping App"`
	Platform  string    `json:"platform" example:"flutter"`
	BundleID  string    `json:"bundle_id" example:"com.company.app"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type APIKeyCredentialResponse struct {
	ApplicationID  string `json:"application_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PublishableKey string `json:"publishable_key" example:"pk_live_a1b2c3d4e5f6..."`
	RawSecretKey   string `json:"secret_key" example:"sk_secret_x9y8z7w6v5u4..."`
	RateLimit      int    `json:"rate_limit" example:"120"`
}

type DeveloperLoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"jane@company.com"`
	Password string `json:"password" binding:"required" example:"securepass123"`
}

type LoginResponse struct {
	Token     string                 `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	Developer map[string]interface{} `json:"developer"`
}
