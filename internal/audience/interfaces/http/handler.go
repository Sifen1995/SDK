package http

import (
	"net/http"

	"skykin-platform/internal/audience/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

// Handler serves Audiencemart browse endpoints on the ad portal.
type Handler struct {
	list       *application.ListService
	candidates *application.ListSegmentCandidatesUseCase
}

func NewHandler(list *application.ListService, candidates *application.ListSegmentCandidatesUseCase) *Handler {
	return &Handler{list: list, candidates: candidates}
}

// ListSegments godoc
// @Summary      List purchasable audience segments
// @Description  Returns active Audiencemart catalog segments filtered by the advertiser subscription plan. Starter plan returns an empty list with audiencemart_enabled=false.
// @Tags         Ad Portal - Audience
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  application.ListSegmentsResult
// @Failure      403  {object}  platformHTTP.APIError
// @Router       /ad-portal/audience/segments [get]
func (h *Handler) ListSegments(c *gin.Context) {
	aid, _ := c.Get("advertiser_id")
	result, err := h.list.ListForAdvertiser(c.Request.Context(), aid.(string))
	if err != nil {
		platformHTTP.Error(c, http.StatusForbidden, "list segments failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetSegment godoc
// @Summary      Get audience segment by ID
// @Description  Returns a single active Audiencemart segment when the advertiser plan includes Audiencemart.
// @Tags         Ad Portal - Audience
// @Produce      json
// @Security     BearerAuth
// @Param        segment_id  path  string  true  "Segment ID"
// @Success      200  {object}  application.SegmentDTO
// @Failure      403  {object}  platformHTTP.APIError
// @Failure      404  {object}  platformHTTP.APIError
// @Router       /ad-portal/audience/segments/{segment_id} [get]
func (h *Handler) GetSegment(c *gin.Context) {
	aid, _ := c.Get("advertiser_id")
	seg, err := h.list.GetForAdvertiser(c.Request.Context(), aid.(string), c.Param("segment_id"))
	if err != nil {
		status := http.StatusForbidden
		if err.Error() == "segment not found" {
			status = http.StatusNotFound
		}
		platformHTTP.Error(c, status, "get segment failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, seg)
}
