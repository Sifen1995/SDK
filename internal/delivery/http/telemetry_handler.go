package http

import (
	"net/http"
	"strings"

	deliveryApp "skykin-platform/internal/delivery/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

// AnonymousTrackRequest is a non-consented impression (or related) bill track payload.
type AnonymousTrackRequest struct {
	CampaignID string `json:"campaign_id" binding:"required,uuid" example:"c1a2b3c4-d5e6-7890-abcd-ef1234567890"`
	EventType  string `json:"event_type" binding:"required" example:"impression"`
}

// TelemetryTrackRequest is a consented ad interaction from the Flutter SDK.
type TelemetryTrackRequest struct {
	CampaignID       string  `json:"campaign_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	EventType        string  `json:"event_type" binding:"required" example:"impression"`
	PseudonymousID   string  `json:"pseudonymous_id" binding:"omitempty,uuid" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"`
	TransactionValue float64 `json:"transaction_value" binding:"omitempty,gte=0" example:"0"`
	OccurredAt       string  `json:"occurred_at" binding:"omitempty" example:"2026-07-18T12:00:00Z"`
	InstallToken     string  `json:"install_token,omitempty" binding:"omitempty" example:"signed-install-token"`
}

// TelemetryHandler accepts high-volume consented telemetry and write-behinds via Redis Streams.
type TelemetryHandler struct {
	ingest *deliveryApp.TelemetryIngestService
}

func NewTelemetryHandler(ingest *deliveryApp.TelemetryIngestService) *TelemetryHandler {
	return &TelemetryHandler{ingest: ingest}
}

// Track godoc
// @Summary      Track consented ad billing event
// @Description  Accepts a consented ad tracking log (impression/click/install/signup/purchase). Impression/click events are deduplicated via Redis SETNX lock:telemetry:{pseudonymous_id}:{campaign_id}:{event_type} (impression 5m, click 1h) before XADD to stream:billing_events. Duplicates still return 202 without enqueueing. Install events require install_token. Authorize with X-API-Key and X-SDK-Secret.
// @Tags         SDK - Bill Track
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Param        body  body  TelemetryTrackRequest  true  "Consented ad track payload"
// @Success      202  "Accepted — queued on stream:billing_events (or silently deduplicated)"
// @Failure      400  {object}  platformHTTP.APIError
// @Failure      401  {object}  platformHTTP.APIError
// @Failure      503  {object}  platformHTTP.APIError
// @Router       /telemetry/track [post]
func (h *TelemetryHandler) Track(c *gin.Context) {
	if h == nil || h.ingest == nil {
		platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry stream unavailable", "")
		return
	}

	var req TelemetryTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid telemetry payload", err.Error())
		return
	}

	err := h.ingest.TrackConsented(c.Request.Context(), deliveryApp.ConsentedTrackCommand{
		CampaignID:       req.CampaignID,
		EventType:        req.EventType,
		PseudonymousID:   req.PseudonymousID,
		TransactionValue: req.TransactionValue,
		OccurredAt:       req.OccurredAt,
		InstallToken:     req.InstallToken,
	})
	if err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "unavailable"):
			platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry stream unavailable", msg)
		case strings.Contains(msg, "dedup failed"), strings.Contains(msg, "enqueue failed"):
			platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry enqueue failed", msg)
		default:
			platformHTTP.Error(c, http.StatusBadRequest, "invalid telemetry payload", msg)
		}
		return
	}
	c.Status(http.StatusAccepted)
}

// TrackAnonymous godoc
// @Summary      Track anonymous (non-consented) ad impression
// @Description  Accepts a minimal bill-track payload for non-consented users (campaign_id + event_type). Enqueues to Redis Stream stream:billing_events with source=anonymous and returns 202. Two independent consumer groups write behind: billing_processor_group → billing_events; delivery_log_processor_group → campaign_delivery_logs. Authorize with X-API-Key and X-SDK-Secret.
// @Tags         SDK - Bill Track
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Param        body  body  AnonymousTrackRequest  true  "Anonymous ad track payload"
// @Success      202  "Accepted — queued on stream:billing_events"
// @Failure      400  {object}  platformHTTP.APIError
// @Failure      401  {object}  platformHTTP.APIError
// @Failure      503  {object}  platformHTTP.APIError
// @Router       /telemetry/track-anonymous [post]
func (h *TelemetryHandler) TrackAnonymous(c *gin.Context) {
	if h == nil || h.ingest == nil {
		platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry stream unavailable", "")
		return
	}

	var req AnonymousTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid anonymous track payload", err.Error())
		return
	}

	err := h.ingest.TrackAnonymous(c.Request.Context(), deliveryApp.AnonymousTrackCommand{
		CampaignID: req.CampaignID,
		EventType:  req.EventType,
	})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "unavailable") || strings.Contains(msg, "enqueue failed") {
			platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry enqueue failed", msg)
			return
		}
		platformHTTP.Error(c, http.StatusBadRequest, "invalid anonymous track payload", msg)
		return
	}
	c.Status(http.StatusAccepted)
}
