package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	permApp "skykin-platform/internal/permissions/application"
	"skykin-platform/internal/permissions/domain"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type permissionLister interface {
	List(ctx context.Context) ([]*domain.Permission, error)
}

type roleLister interface {
	List(ctx context.Context) ([]*domain.Role, error)
}

type roleCreator interface {
	Execute(ctx context.Context, name, description string, createdBy uuid.UUID) (*domain.Role, error)
}

type permissionAssigner interface {
	Execute(ctx context.Context, roleID, permissionID, grantedBy uuid.UUID) error
}

type permissionRevoker interface {
	Execute(ctx context.Context, roleID, permissionID uuid.UUID) error
}

type Handler struct {
	listPermissions permissionLister
	listRoles       roleLister
	createRole      roleCreator
	assignPerm      permissionAssigner
	revokePerm      permissionRevoker
}

func NewHandler(
	listPermissions *permApp.ListPermissionsUseCase,
	listRoles *permApp.ListRolesUseCase,
	createRole *permApp.CreateRoleUseCase,
	assignPerm *permApp.AssignPermissionUseCase,
	revokePerm *permApp.RevokePermissionUseCase,
) *Handler {
	return &Handler{
		listPermissions: listPermissions,
		listRoles:       listRoles,
		createRole:      createRole,
		assignPerm:      assignPerm,
		revokePerm:      revokePerm,
	}
}

// ListPermissions godoc
// @Summary      List all permissions
// @Tags         Ad Portal - Admin Permissions
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   PermissionResponse
// @Failure      500  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/permissions [get]
func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.listPermissions.List(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "failed to list permissions", nil)
		return
	}
	c.JSON(http.StatusOK, mapPermissions(perms))
}

// ListRoles godoc
// @Summary      List all roles with permissions
// @Tags         Ad Portal - Admin Permissions
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   RoleResponse
// @Failure      500  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/roles [get]
func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.listRoles.List(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "failed to list roles", nil)
		return
	}
	c.JSON(http.StatusOK, mapRoles(roles))
}

// CreateRole godoc
// @Summary      Create a custom role
// @Tags         Ad Portal - Admin Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateRoleRequest  true  "Role"
// @Success      201   {object}  RoleResponse
// @Failure      422   {object}  platformHTTP.APIError
// @Failure      500   {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/roles [post]
func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusUnprocessableEntity, "invalid payload", err.Error())
		return
	}
	adminID, err := portalUserUUID(c)
	if err != nil {
		platformHTTP.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	role, err := h.createRole.Execute(c.Request.Context(), req.Name, req.Description, adminID)
	if err != nil {
		writeUseCaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, mapRole(role))
}

// AssignPermission godoc
// @Summary      Assign permission to role
// @Tags         Ad Portal - Admin Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        role_id  path  string  true  "Role ID"
// @Param        body     body  AssignPermissionRequest  true  "Permission"
// @Success      200  {object}  MessageResponse
// @Router       /ad-portal/admin/roles/{role_id}/permissions [post]
func (h *Handler) AssignPermission(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		platformHTTP.Error(c, http.StatusUnprocessableEntity, "invalid role id", nil)
		return
	}
	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusUnprocessableEntity, "invalid payload", err.Error())
		return
	}
	permID, err := uuid.Parse(req.PermissionID)
	if err != nil {
		platformHTTP.Error(c, http.StatusUnprocessableEntity, "invalid permission id", nil)
		return
	}
	adminID, err := portalUserUUID(c)
	if err != nil {
		platformHTTP.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	if err := h.assignPerm.Execute(c.Request.Context(), roleID, permID, adminID); err != nil {
		writeUseCaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, MessageResponse{Message: "permission assigned successfully"})
}

// RevokePermission godoc
// @Summary      Revoke permission from role
// @Tags         Ad Portal - Admin Permissions
// @Produce      json
// @Security     BearerAuth
// @Param        role_id        path  string  true  "Role ID"
// @Param        permission_id  path  string  true  "Permission ID"
// @Success      200  {object}  MessageResponse
// @Router       /ad-portal/admin/roles/{role_id}/permissions/{permission_id} [delete]
func (h *Handler) RevokePermission(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		platformHTTP.Error(c, http.StatusUnprocessableEntity, "invalid role id", nil)
		return
	}
	permID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		platformHTTP.Error(c, http.StatusUnprocessableEntity, "invalid permission id", nil)
		return
	}
	if err := h.revokePerm.Execute(c.Request.Context(), roleID, permID); err != nil {
		writeUseCaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, MessageResponse{Message: "permission revoked successfully"})
}

func (h *Handler) RegisterRoutes(admin *gin.RouterGroup) {
	admin.GET("/permissions", h.ListPermissions)
	admin.GET("/roles", h.ListRoles)
	admin.POST("/roles", h.CreateRole)
	admin.POST("/roles/:role_id/permissions", h.AssignPermission)
	admin.DELETE("/roles/:role_id/permissions/:permission_id", h.RevokePermission)
}

func portalUserUUID(c *gin.Context) (uuid.UUID, error) {
	id, ok := c.Get("portal_user_id")
	if !ok {
		return uuid.Nil, errors.New("missing user")
	}
	return uuid.Parse(id.(string))
}

func writeUseCaseError(c *gin.Context, err error) {
	msg := err.Error()
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(msg, "not found"):
		platformHTTP.Error(c, http.StatusNotFound, msg, nil)
	case strings.Contains(msg, "cannot be modified") || strings.Contains(msg, "reserved"):
		platformHTTP.Error(c, http.StatusForbidden, msg, nil)
	case strings.Contains(msg, "already exists") || strings.Contains(msg, "required"):
		platformHTTP.Error(c, http.StatusUnprocessableEntity, msg, nil)
	default:
		platformHTTP.Error(c, http.StatusInternalServerError, "operation failed", nil)
	}
}

func mapPermissions(perms []*domain.Permission) []PermissionResponse {
	out := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		out[i] = PermissionResponse{
			ID: p.ID.String(), Name: p.Name, Resource: p.Resource,
			Action: p.Action, Description: p.Description,
		}
	}
	return out
}

func mapRoles(roles []*domain.Role) []RoleResponse {
	out := make([]RoleResponse, len(roles))
	for i, r := range roles {
		out[i] = mapRole(r)
	}
	return out
}

func mapRole(r *domain.Role) RoleResponse {
	perms := make([]PermissionResponse, len(r.Permissions))
	for i, p := range r.Permissions {
		perms[i] = PermissionResponse{
			ID: p.ID.String(), Name: p.Name, Resource: p.Resource,
			Action: p.Action, Description: p.Description,
		}
	}
	return RoleResponse{
		ID: r.ID.String(), Name: r.Name, Description: r.Description,
		IsSystem: r.IsSystem, Permissions: perms,
	}
}
