package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"skykin-platform/internal/delivery/application"
)

type CPCClickHandler struct {
	cpcService clickProcessor
}

type clickProcessor interface {
	ProcessClick(ctx context.Context, campaignID, token, billingModel string) error
}

func NewCPCClickHandler(cpcService *application.CPCClickService) *CPCClickHandler {
	return &CPCClickHandler{cpcService: cpcService}
}

// TrackAnonymousClick godoc
// @Summary      Track anonymous billing click
// @Description  Accepts an anonymous click callback with its billing model. The endpoint validates the token, queues the event for async billing processing, and returns 202 Accepted immediately.
// @Tags         SDK - Bill Track
// @Accept       json
// @Produce      json
// @Param        campaign_id query string true "Campaign ID"
// @Param        token query string true "Signed click token"
// @Param        billing_model query string true "Billing model (CPM, CPC, CPI, CPA, or REV_SHARE)"
// @Success      202  {object}  map[string]string
// @Failure      400  {string}  string
// @Failure      401  {string}  string
// @Router       /telemetry/anonymous-click [post]
func (h *CPCClickHandler) TrackAnonymousClick(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.cpcService == nil {
		http.Error(w, "Anonymous click tracking is unavailable", http.StatusServiceUnavailable)
		return
	}

	campaignID := strings.TrimSpace(r.URL.Query().Get("campaign_id"))
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	billingModel := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("billing_model")))

	if campaignID == "" || token == "" || !isBillingModel(billingModel) {
		http.Error(w, "Missing campaign_id, token, or valid billing_model", http.StatusBadRequest)
		return
	}

	// Verify token signature & queue event
	err := h.cpcService.ProcessClick(r.Context(), campaignID, token, billingModel)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, application.ErrClickQueueUnavailable) {
			status = http.StatusServiceUnavailable
		}
		http.Error(w, "Invalid or expired click token", status)
		return
	}

	// Return 202 Accepted instantly to the client
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"message": "Anonymous click registered",
	})
}

func isBillingModel(model string) bool {
	switch model {
	case "CPM", "CPC", "CPI", "CPA", "REV_SHARE":
		return true
	default:
		return false
	}
}
