package controller

import (
	"net/http"
	"skykin-platform/internal/auth/dto"
	"skykin-platform/internal/auth/service"
	"skykin-platform/internal/common/response"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService service.AuthService
}

func NewAuthController(s service.AuthService) *AuthController {
	return &AuthController{authService: s}
}

func handleControllerError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	response.HandleError(c, err)
}

func (ctrl *AuthController) CreateApplication(c *gin.Context) {
	// Extract developer ID from authentication token context (set by platform developer JWT login)
	devID, exists := c.Get("developer_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthenticated developer context", nil)
		return
	}

	var req dto.ApplicationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	appRes, credentials, err := ctrl.authService.RegisterApplication(c.Request.Context(), devID.(string), req)
	if err != nil {
		handleControllerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"application": appRes,
		"credentials": credentials,
	})

}

func (ctrl *AuthController) RegisterDeveloper(c *gin.Context) {
	var req dto.DeveloperRegisterRequest

	// 1. Bind and automatically validate incoming JSON against our DTO struct tags
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// 2. Hand off the raw data to the service layer for business rules and hashing
	dev, err := ctrl.authService.RegisterDeveloper(c.Request.Context(), req)
	if err != nil {
		handleControllerError(c, err)
		return
	}

	// 3. Return a clean, successful tracking response (GORM struct tags automatically hide PasswordHash)
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Developer registered successfully",
		"data": gin.H{
			"developer": gin.H{
				"id":         dev.ID.String(),
				"name":       dev.Name,
				"email":      dev.Email,
				"created_at": dev.CreatedAt,
			},
		},
	})
}

func (ctrl *AuthController) LoginDeveloper(c *gin.Context) {
	var req dto.DeveloperLoginRequest

	// Bind input validations
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body structure", err.Error())
		return
	}

	// Handle login execution flow
	res, err := ctrl.authService.LoginDeveloper(c.Request.Context(), req)
	if err != nil {
		// Keep errors ambiguous for security, avoiding telling hackers if an email exists
		response.Error(c, http.StatusUnauthorized, "Authentication failed", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Session authorized successfully", res)
}
