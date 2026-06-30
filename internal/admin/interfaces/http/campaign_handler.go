package http

import (
	"net/http"

	campaignApp "skykin-platform/internal/campaigns/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

// CampaignHandler exposes operator campaign moderation endpoints.
type CampaignHandler struct {
	moderation *campaignApp.ModerationService
}

func NewCampaignHandler(moderation *campaignApp.ModerationService) *CampaignHandler {
	return &CampaignHandler{moderation: moderation}
}

// ListPendingCampaigns godoc
// @Summary      List campaigns pending operator validation
// @Tags         Ad Portal - Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  CampaignListResponse
// @Failure      403  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/campaigns/pending [get]
func (h *CampaignHandler) ListPendingCampaigns(c *gin.Context) {
	list, err := h.moderation.ListPending(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaigns": list, "count": len(list)})
}

// ValidateCampaign godoc
// @Summary      Approve or reject a pending campaign
// @Description  Approve runs creative validation against the campaign channel and sets moderation_status=approved. Reject sets moderation_status=rejected.
// @Tags         Ad Portal - Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                 true  "Campaign ID"
// @Param        body  body  ValidateCampaignRequest  true  "approve or reject"
// @Success      200   {object}  skykin-platform_internal_campaigns_model.Campaign
// @Failure      400   {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/campaigns/{id}/validate [post]
func (h *CampaignHandler) ValidateCampaign(c *gin.Context) {
	var req ValidateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	operatorID, _ := c.Get("portal_user_id")
	camp, err := h.moderation.Validate(c.Request.Context(), c.Param("id"), operatorID.(string), req.Action, req.Notes)
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "validation failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, camp)
}

// ActivateCampaign godoc
// @Summary      Activate an approved campaign (go live)
// @Description  Sets is_active=true. Campaign must have moderation_status=approved and validation_status=passed.
// @Tags         Ad Portal - Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Campaign ID"
// @Success      200  {object}  skykin-platform_internal_campaigns_model.Campaign
// @Failure      400  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/campaigns/{id}/activate [post]
func (h *CampaignHandler) ActivateCampaign(c *gin.Context) {
	operatorID, _ := c.Get("portal_user_id")
	camp, err := h.moderation.Activate(c.Request.Context(), c.Param("id"), operatorID.(string))
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "activation failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, camp)
}
