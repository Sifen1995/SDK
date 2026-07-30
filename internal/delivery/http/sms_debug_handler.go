package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	deliverydomain "skykin-platform/internal/delivery/domain"

	"github.com/gin-gonic/gin"
)

type smsAttemptLister interface {
	ListRecent(ctx context.Context, limit int) ([]deliverydomain.SMSSendAttempt, error)
}

type SMSDebugHandler struct {
	attempts smsAttemptLister
}

func NewSMSDebugHandler(attempts smsAttemptLister) *SMSDebugHandler {
	return &SMSDebugHandler{attempts: attempts}
}

// SMSSendAttemptDTO is the Swagger/demo payload for recent SMS+ sends (phone masked).
type SMSSendAttemptDTO struct {
	CampaignID     string `json:"campaign_id" example:"1a0d7721-ed1d-4bd7-ab63-8195b5e5d91d"`
	PseudonymousID string `json:"pseudonymous_id" example:"a9a1208b-7521-4ff0-8d88-52a48450784b"`
	PhoneMasked    string `json:"phone_masked" example:"+15*******07"`
	Provider       string `json:"provider" example:"mock"`
	Status         string `json:"status" example:"sent"`
	CreatedAt      string `json:"created_at" example:"2026-07-30T10:13:10Z"`
}

// ListRecent godoc
// @Summary      List recent SMS+ send attempts (demo)
// @Description  Returns the latest SMS+ send attempts with masked phone numbers for demo/debug inspection. Authorize with X-API-Key (and X-SDK-Secret for POSTs; GET only needs the API key).
// @Tags         SDK - SMS+
// @Produce      json
// @Security     APIKeyAuth
// @Success      200  {array}   SMSSendAttemptDTO
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Failure      503  {object}  map[string]string
// @Router       /telemetry/sms/debug/sends [get]
func (h *SMSDebugHandler) ListRecent(c *gin.Context) {
	if h == nil || h.attempts == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sms debug unavailable"})
		return
	}
	attempts, err := h.attempts.ListRecent(c.Request.Context(), 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list sms sends failed"})
		return
	}
	out := make([]SMSSendAttemptDTO, 0, len(attempts))
	for i := range attempts {
		out = append(out, SMSSendAttemptDTO{
			CampaignID:     attempts[i].CampaignID,
			PseudonymousID: attempts[i].PseudonymousID,
			PhoneMasked:    maskPhone(attempts[i].PhoneE164),
			Provider:       attempts[i].Provider,
			Status:         attempts[i].Status,
			CreatedAt:      attempts[i].CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, out)
}

func maskPhone(in string) string {
	in = strings.TrimSpace(in)
	if len(in) <= 4 {
		return in
	}
	return in[:3] + strings.Repeat("*", len(in)-5) + in[len(in)-2:]
}
