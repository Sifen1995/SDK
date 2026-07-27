package http

import (
	"encoding/json"
	"net/http"

	"skykin-platform/internal/delivery/application"
)

type CPCClickHandler struct {
	cpcService *application.CPCClickService
}

func NewCPCClickHandler(cpcService *application.CPCClickService) *CPCClickHandler {
	return &CPCClickHandler{cpcService: cpcService}
}

// TrackAnonymousClick godoc
// @Summary      Track anonymous CPC click
// @Description  Accepts an anonymous click callback for a campaign using a signed token. The endpoint validates the token, queues the event for async billing processing, and returns 202 Accepted immediately.
// @Tags         SDK - Bill Track
// @Accept       json
// @Produce      json
// @Param        campaign_id query string true "Campaign ID"
// @Param        token query string true "Signed click token"
// @Success      202  {object}  map[string]string
// @Failure      400  {string}  string
// @Failure      401  {string}  string
// @Router       /telemetry/anonymous-click [post]
func (h *CPCClickHandler) TrackAnonymousClick(w http.ResponseWriter, r *http.Request) {
	campaignID := r.URL.Query().Get("campaign_id")
	token := r.URL.Query().Get("token")

	if campaignID == "" || token == "" {
		http.Error(w, "Missing campaign_id or token", http.StatusBadRequest)
		return
	}

	// Verify token signature & queue event
	err := h.cpcService.ProcessClick(r.Context(), campaignID, token)
	if err != nil {
		http.Error(w, "Invalid or expired click token", http.StatusUnauthorized)
		return
	}

	// Return 202 Accepted instantly to the client
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"message": "Anonymous click registered",
	})
}
