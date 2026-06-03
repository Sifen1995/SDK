package http

import (
	"net/http"

	"skykin-platform/internal/campaigns/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *application.CampaignService
}

func NewHandler(svc *application.CampaignService) *Handler {
	return &Handler{svc: svc}
}

// CreateCampaign godoc
// @Summary      Create campaign with embedded creative
// @Description  One campaign row includes targeting, budget caps, and creative fields per schema. Starts inactive until activated.
// @Tags         Ad Portal - Campaigns
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateCampaignRequest  true  "Campaign + creative"
// @Success      201   {object}  skykin-platform_internal_campaigns_model.Campaign
// @Failure      400   {object}  platformHTTP.APIError
// @Failure      403   {object}  platformHTTP.APIError
// @Router       /ad-portal/campaigns [post]
func (h *Handler) CreateCampaign(c *gin.Context) {
	var req CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	aid, _ := c.Get("advertiser_id")
	role, _ := c.Get("portal_role")
	camp, err := h.svc.Create(c.Request.Context(), aid.(string), role.(string), application.CreateCampaignInput{
		Name: req.Name, TargetIntent: req.TargetIntent, ApplicationID: req.ApplicationID,
		CreativeFormat: req.CreativeFormat, Title: req.Title, BodyText: req.BodyText,
		ImageURL: req.ImageURL, DestinationURL: req.DestinationURL, CanvasJSON: req.CanvasJSON,
		DailyBudgetCap: req.DailyBudgetCap, TotalBudgetCap: req.TotalBudgetCap,
	})
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "create failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, camp)
}

// ListCampaigns godoc
// @Summary      List campaigns
// @Tags         Ad Portal - Campaigns
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  CampaignListResponse
// @Router       /ad-portal/campaigns [get]
func (h *Handler) ListCampaigns(c *gin.Context) {
	aid, _ := c.Get("advertiser_id")
	role, _ := c.Get("portal_role")
	list, err := h.svc.List(c.Request.Context(), aid.(string), role.(string))
	if err != nil {
		platformHTTP.Error(c, http.StatusForbidden, "list failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaigns": list})
}

// GetCampaign godoc
// @Summary      Get campaign
// @Tags         Ad Portal - Campaigns
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Campaign ID"
// @Success      200  {object}  skykin-platform_internal_campaigns_model.Campaign
// @Router       /ad-portal/campaigns/{id} [get]
func (h *Handler) GetCampaign(c *gin.Context) {
	aid, _ := c.Get("advertiser_id")
	role, _ := c.Get("portal_role")
	camp, err := h.svc.Get(c.Request.Context(), aid.(string), role.(string), c.Param("id"))
	if err != nil {
		platformHTTP.Error(c, http.StatusNotFound, "not found", err.Error())
		return
	}
	c.JSON(http.StatusOK, camp)
}

// ActivateCampaign godoc
// @Summary      Activate campaign
// @Description  Sets is_active=true when validation_status is passed.
// @Tags         Ad Portal - Campaigns
// @Security     BearerAuth
// @Param        id   path  string  true  "Campaign ID"
// @Success      200  {object}  skykin-platform_internal_campaigns_model.Campaign
// @Router       /ad-portal/campaigns/{id}/activate [post]
func (h *Handler) ActivateCampaign(c *gin.Context) {
	aid, _ := c.Get("advertiser_id")
	role, _ := c.Get("portal_role")
	camp, err := h.svc.Activate(c.Request.Context(), aid.(string), role.(string), c.Param("id"))
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "activation failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, camp)
}

// PreviewCampaign godoc
// @Summary      Preview campaign creative
// @Tags         Ad Portal - Campaigns
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Campaign ID"
// @Success      200  {object}  CampaignPreviewResponse
// @Router       /ad-portal/campaigns/{id}/preview [get]
func (h *Handler) PreviewCampaign(c *gin.Context) {
	aid, _ := c.Get("advertiser_id")
	role, _ := c.Get("portal_role")
	prev, err := h.svc.Preview(c.Request.Context(), aid.(string), role.(string), c.Param("id"))
	if err != nil {
		platformHTTP.Error(c, http.StatusNotFound, "preview failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, prev)
}

// CampaignListResponse for swagger.
type CampaignListResponse struct {
	Campaigns []interface{} `json:"campaigns"`
}

// CampaignPreviewResponse for swagger.
type CampaignPreviewResponse struct {
	Format       string                 `json:"format"`
	CampaignName string                 `json:"campaign_name"`
	Simulator    bool                   `json:"simulator"`
	ChannelLabel string                 `json:"channel_label"`
	Preview      map[string]interface{} `json:"preview"`
}
