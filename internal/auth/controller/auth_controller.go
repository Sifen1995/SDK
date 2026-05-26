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

// CreateApplication godoc
// @Summary      Create a new application
// @Description  Registers a new application under the authenticated developer and generates SDK API key credentials. The secret key is only shown once.
// @Tags         Portal - Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.ApplicationCreateRequest  true  "Application details"
// @Success      201   {object}  response.JSONResponse
// @Failure      400   {object}  response.APIError
// @Failure      401   {object}  response.APIError
// @Router       /portal/applications [post]
func (ctrl *AuthController) CreateApplication(c *gin.Context) {
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

	response.Success(c, http.StatusCreated, "Application registered successfully", gin.H{
		"application": appRes,
		"credentials": credentials,
	})
}

// RegisterDeveloper godoc
// @Summary      Register a new developer
// @Description  Creates a new developer account for the Skykin portal
// @Tags         Portal - Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.DeveloperRegisterRequest  true  "Developer registration details"
// @Success      201   {object}  response.JSONResponse
// @Failure      400   {object}  response.APIError
// @Failure      409   {object}  response.APIError
// @Router       /portal/register [post]
func (ctrl *AuthController) RegisterDeveloper(c *gin.Context) {
	var req dto.DeveloperRegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	dev, err := ctrl.authService.RegisterDeveloper(c.Request.Context(), req)
	if err != nil {
		handleControllerError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Developer registered successfully", gin.H{
		"developer": gin.H{
			"id":         dev.ID.String(),
			"name":       dev.Name,
			"email":      dev.Email,
			"created_at": dev.CreatedAt,
		},
	})
}

// GetApplications godoc
// @Summary      List applications
// @Description  Returns all applications belonging to the authenticated developer
// @Tags         Portal - Applications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.JSONResponse
// @Failure      401  {object}  response.APIError
// @Router       /portal/applications [get]
func (ctrl *AuthController) GetApplications(c *gin.Context) {
	devID, exists := c.Get("developer_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthenticated developer context", nil)
		return
	}

	apps, err := ctrl.authService.GetApplications(c.Request.Context(), devID.(string))
	if err != nil {
		handleControllerError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Applications retrieved", gin.H{
		"applications": apps,
	})
}

// LoginDeveloper godoc
// @Summary      Login developer
// @Description  Authenticates a developer and returns a JWT token
// @Tags         Portal - Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.DeveloperLoginRequest  true  "Login credentials"
// @Success      200   {object}  response.JSONResponse
// @Failure      400   {object}  response.APIError
// @Failure      401   {object}  response.APIError
// @Router       /portal/login [post]
func (ctrl *AuthController) LoginDeveloper(c *gin.Context) {
	var req dto.DeveloperLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body structure", err.Error())
		return
	}

	res, err := ctrl.authService.LoginDeveloper(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Authentication failed", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Session authorized successfully", res)
}
