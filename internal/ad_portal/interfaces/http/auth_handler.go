package http

import (
	"net/http"

	"skykin-platform/internal/ad_portal/application"
	adportalvalidation "skykin-platform/internal/ad_portal/validation"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *application.AuthService
}

func NewAuthHandler(auth *application.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register godoc
// @Summary      Register advertiser
// @Tags         Ad Portal - Auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest   true  "Registration"
// @Success      201   {object}  RegisterResponse
// @Failure      400   {object}  platformHTTP.APIError
// @Router       /ad-portal/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	email, err := adportalvalidation.Register(req.Name, req.Email, req.Password, req.CompanyName, req.Role)
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "validation failed", err.Error())
		return
	}
	u, err := h.auth.Register(c.Request.Context(), req.Name, email, req.Password, req.CompanyName, req.Role)
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "registration failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": application.UserResponse(u)})
}

// Login godoc
// @Summary      Login to ad portal
// @Tags         Ad Portal - Auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Credentials"
// @Success      200   {object}  AdPortalLoginResponse
// @Failure      401   {object}  platformHTTP.APIError
// @Router       /ad-portal/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	email, err := adportalvalidation.Login(req.Email, req.Password)
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "validation failed", err.Error())
		return
	}
	token, u, err := h.auth.Login(c.Request.Context(), email, req.Password)
	if err != nil {
		platformHTTP.Error(c, http.StatusUnauthorized, "login failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": application.UserResponse(u)})
}

// Me godoc
// @Summary      Current advertiser profile
// @Tags         Ad Portal - Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  MeResponse
// @Failure      404  {object}  platformHTTP.APIError
// @Router       /ad-portal/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	id, _ := c.Get("portal_user_id")
	u, err := h.auth.Me(c.Request.Context(), id.(string))
	if err != nil {
		platformHTTP.Error(c, http.StatusNotFound, "user not found", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": application.UserResponse(u)})
}

// CreateUser godoc
// @Summary      Create portal user (operator admin)
// @Tags         Ad Portal - Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateUserRequest  true  "User"
// @Success      201   {object}  RegisterResponse
// @Failure      403   {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/users [post]
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	email, err := adportalvalidation.CreateUser(req.Name, req.Email, req.Password, req.Role, req.CompanyName)
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "validation failed", err.Error())
		return
	}
	u, err := h.auth.CreateOperatorUser(c.Request.Context(), req.Name, email, req.Password, req.Role, req.CompanyName)
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "create user failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": application.UserResponse(u)})
}
