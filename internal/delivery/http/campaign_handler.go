package http

import (
	"context"
	"net/http"

	campaignApp "skykin-platform/internal/campaigns/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

type activeMasterLister interface {
	ListActiveMaster(ctx context.Context) ([]campaignApp.CampaignWithClickToken, error)
}

// CampaignHandler serves anonymous SDK campaign delivery endpoints.
type CampaignHandler struct {
	campaigns activeMasterLister
}

func NewCampaignHandler(campaigns activeMasterLister) *CampaignHandler {
	return &CampaignHandler{campaigns: campaigns}
}

// ListAnonymousCampaigns godoc
// @Summary      List active campaigns for anonymous delivery
// @Description  Returns the master list of all active, budget-unexhausted campaigns across intents. Stateless — no excluded IDs accepted. Flutter applies on-device frequency capping (e.g. 3 exposures). Authorize with X-API-Key and X-SDK-Secret.
// @Tags         SDK - Campaigns
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Success      200  {array}   AnonymousCampaignDTO
// @Failure      401  {object}  platformHTTP.APIError
// @Failure      500  {object}  platformHTTP.APIError
// @Router       /campaigns/anonymous [get]
func (h *CampaignHandler) ListAnonymousCampaigns(c *gin.Context) {
	if h == nil || h.campaigns == nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "anonymous campaigns unavailable", "")
		return
	}

	campaigns, err := h.campaigns.ListActiveMaster(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list anonymous campaigns failed", err.Error())
		return
	}

	out := make([]AnonymousCampaignDTO, 0, len(campaigns))
	for i := range campaigns {
		out = append(out, toAnonymousCampaignDTO(&campaigns[i]))
	}
	c.JSON(http.StatusOK, out)
}

func toAnonymousCampaignDTO(c *campaignApp.CampaignWithClickToken) AnonymousCampaignDTO {
	if c == nil {
		return AnonymousCampaignDTO{}
	}
	canvas := c.Campaign.CanvasJSON
	if canvas == nil {
		canvas = map[string]any{}
	}
	cap := c.FrequencyCapPerDay
	if cap <= 0 {
		cap = 3
	}
	return AnonymousCampaignDTO{
		ID:                 c.ID,
		Name:               c.Name,
		TargetIntent:       c.TargetIntent,
		ChannelCode:        c.ChannelCode,
		Title:              c.Title,
		BodyText:           c.BodyText,
		ImageURL:           c.ImageURL,
		DestinationURL:     c.DestinationURL,
		CanvasJSON:         canvas,
		FrequencyCapPerDay: cap,
		ClickToken:         c.ClickToken,
		PlanName:           c.PlanName,
		PlanMonthlyFeeETB:  c.PlanMonthlyFeeETB,
	}
}
