package http

import (
	"net/http"
	"time"

	audienceDomain "skykin-platform/internal/audience/domain"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

type CandidateResponse struct {
	ID            string  `json:"id"`
	IntentName    string  `json:"intent_name"`
	UserCount     int     `json:"user_count"`
	AvgConfidence float64 `json:"avg_confidence"`
	AvgDaysActive float64 `json:"avg_days_active"`
	MinDaysActive int     `json:"min_days_active"`
	LookbackDays  int     `json:"lookback_days"`
	Status        string  `json:"status"`
	ScannedAt     string  `json:"scanned_at"`
}

func toCandidateResponse(c *audienceDomain.SegmentCandidate) CandidateResponse {
	return CandidateResponse{
		ID: c.ID.String(), IntentName: c.IntentName, UserCount: c.UserCount,
		AvgConfidence: c.AvgConfidence, AvgDaysActive: c.AvgDaysActive,
		MinDaysActive: c.MinDaysActive, LookbackDays: c.LookbackDays,
		Status: string(c.Status), ScannedAt: c.ScannedAt.Format(time.RFC3339),
	}
}

// ListSegmentCandidates godoc
// @Summary      List segment candidates awaiting review
// @Description  Returns audience segment candidates created from intent consistency findings.
// @Tags         Ad Portal - Admin
// @Produce      json
// @Security     BearerAuth
// @Param        status  query  string  false  "pending|approved|rejected" default(pending)
// @Success      200  {array}  CandidateResponse
// @Failure      401  {object}  platformHTTP.APIError
// @Failure      403  {object}  platformHTTP.APIError
// @Failure      500  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/audience/segment-candidates [get]
func (h *Handler) ListSegmentCandidates(c *gin.Context) {
	status := c.DefaultQuery("status", string(audienceDomain.CandidateStatusPending))
	list, err := h.candidates.ListByStatus(c.Request.Context(), audienceDomain.CandidateStatus(status))
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list candidates failed", err.Error())
		return
	}
	out := make([]CandidateResponse, 0, len(list))
	for _, item := range list {
		out = append(out, toCandidateResponse(item))
	}
	c.JSON(http.StatusOK, out)
}
