package http

import (
	"net/http"
	"strconv"

	adminApp "skykin-platform/internal/admin/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

// UsersHandler exposes operator endpoints for SDK end-users.
type UsersHandler struct {
	getUsers *adminApp.GetUsersWithIntentsUseCase
}

func NewUsersHandler(getUsers *adminApp.GetUsersWithIntentsUseCase) *UsersHandler {
	return &UsersHandler{getUsers: getUsers}
}

// ListUsers godoc
// @Summary      List SDK users with latest intent
// @Description  Returns paginated SDK users (bigint user_id, no phone/external_user_id) enriched with each user's most recent intent prediction when available.
// @Tags         Ad Portal - Admin
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int  false  "Page number (default 1)"
// @Param        per_page  query  int  false  "Items per page (default 20, max 100)"
// @Success      200  {object}  skykin-platform_internal_admin_application.GetUsersResult
// @Failure      403  {object}  platformHTTP.APIError
// @Failure      500  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/sdk-users [get]
func (h *UsersHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	result, err := h.getUsers.Execute(c.Request.Context(), adminApp.GetUsersRequest{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list users failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}
