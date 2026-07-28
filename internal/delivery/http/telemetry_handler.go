package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	platformHTTP "skykin-platform/internal/platform/http"
	platformredis "skykin-platform/internal/platform/redis"

	"github.com/gin-gonic/gin"
)

const (
	billingEventsStream    = "stream:billing_events"
	billingEventsStreamMax = 100000

	telemetryDedupKeyPrefix = "lock:telemetry:"
	telemetryClickTTL       = time.Hour
	telemetryImpressionTTL  = 5 * time.Minute
)

// AnonymousTrackRequest is a non-consented impression (or related) bill track payload.
type AnonymousTrackRequest struct {
	// CampaignID of the served creative
	CampaignID string `json:"campaign_id" binding:"required,uuid" example:"c1a2b3c4-d5e6-7890-abcd-ef1234567890"`
	// EventType for anonymous bill track (impression)
	EventType string `json:"event_type" binding:"required" example:"impression"`
}

// TelemetryTrackRequest is a consented ad interaction from the Flutter SDK.
type TelemetryTrackRequest struct {
	// CampaignID of the served creative
	CampaignID string `json:"campaign_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	// EventType: impression | click | install | signup | purchase
	EventType string `json:"event_type" binding:"required" example:"impression"`
	// PseudonymousID from consent; required for impression/click Redis dedup
	PseudonymousID string `json:"pseudonymous_id" binding:"omitempty,uuid" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"`
	// TransactionValue used for REV_SHARE / purchase events
	TransactionValue float64 `json:"transaction_value" binding:"omitempty,gte=0" example:"0"`
	// OccurredAt RFC3339 timestamp; defaults to server UTC now when omitted
	OccurredAt   string `json:"occurred_at" binding:"omitempty" example:"2026-07-18T12:00:00Z"`
	InstallToken string `json:"install_token,omitempty"`
}

type billingStreamPublisher interface {
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
	XAdd(ctx context.Context, stream string, maxLen int64, values map[string]interface{}) (string, error)
}

// TelemetryHandler accepts high-volume consented telemetry and write-behinds via Redis Streams.
type TelemetryHandler struct {
	rdb billingStreamPublisher
}

func NewTelemetryHandler(rdb *platformredis.RedisClient) *TelemetryHandler {
	return &TelemetryHandler{rdb: rdb}
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
	if h == nil || h.rdb == nil {
		platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry stream unavailable", "")
		return
	}

	var req TelemetryTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid telemetry payload", err.Error())
		return
	}

	eventType := strings.ToLower(strings.TrimSpace(req.EventType))
	switch eventType {
	case "impression", "click", "install", "signup", "purchase":
	default:
		platformHTTP.Error(c, http.StatusBadRequest, "invalid event_type", "must be impression, click, install, signup, or purchase")
		return
	}

	campaignID := strings.TrimSpace(req.CampaignID)
	pseudonymousID := strings.TrimSpace(req.PseudonymousID)

	if eventType == "install" && strings.TrimSpace(req.InstallToken) == "" {
		platformHTTP.Error(c, http.StatusBadRequest, "missing_token", "install_token is required for install events")
		return
	}

	// High-speed dedup gate for spam impressions/clicks before stream write-behind.
	if ttl, ok := telemetryDedupTTL(eventType); ok {
		if pseudonymousID == "" {
			platformHTTP.Error(c, http.StatusBadRequest, "pseudonymous_id required", "required for impression and click deduplication")
			return
		}
		lockKey := telemetryDedupKeyPrefix + pseudonymousID + ":" + campaignID + ":" + eventType
		acquired, err := h.rdb.SetNX(c.Request.Context(), lockKey, "1", ttl)
		if err != nil {
			platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry dedup failed", err.Error())
			return
		}
		if !acquired {
			// Duplicate within TTL window — accept silently, do not enqueue.
			c.Status(http.StatusAccepted)
			return
		}
	}

	occurredAt := strings.TrimSpace(req.OccurredAt)
	if occurredAt == "" {
		occurredAt = time.Now().UTC().Format(time.RFC3339Nano)
	} else if _, err := time.Parse(time.RFC3339Nano, occurredAt); err != nil {
		if _, err2 := time.Parse(time.RFC3339, occurredAt); err2 != nil {
			platformHTTP.Error(c, http.StatusBadRequest, "invalid occurred_at", "expected RFC3339")
			return
		}
	}

	values := map[string]interface{}{
		"campaign_id":       campaignID,
		"event_type":        eventType,
		"transaction_value": strconv.FormatFloat(req.TransactionValue, 'f', 4, 64),
		"occurred_at":       occurredAt,
	}
	if eventType == "install" {
		values["install_token"] = strings.TrimSpace(req.InstallToken)
	}
	if pseudonymousID != "" {
		values["pseudonymous_id"] = pseudonymousID
	}

	if _, err := h.rdb.XAdd(c.Request.Context(), billingEventsStream, billingEventsStreamMax, values); err != nil {
		platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry enqueue failed", err.Error())
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
	if h == nil || h.rdb == nil {
		platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry stream unavailable", "")
		return
	}

	var req AnonymousTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid anonymous track payload", err.Error())
		return
	}

	eventType := strings.ToLower(strings.TrimSpace(req.EventType))
	if eventType != "impression" {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid event_type", "anonymous track currently accepts impression only")
		return
	}

	campaignID := strings.TrimSpace(req.CampaignID)
	occurredAt := time.Now().UTC().Format(time.RFC3339Nano)

	values := map[string]interface{}{
		"campaign_id":       campaignID,
		"event_type":        eventType,
		"transaction_value": "0.0000",
		"occurred_at":       occurredAt,
		"source":            "anonymous",
	}

	if _, err := h.rdb.XAdd(c.Request.Context(), billingEventsStream, billingEventsStreamMax, values); err != nil {
		platformHTTP.Error(c, http.StatusServiceUnavailable, "telemetry enqueue failed", err.Error())
		return
	}

	c.Status(http.StatusAccepted)
}

// telemetryDedupTTL returns the lock TTL for event types that must be deduplicated.
func telemetryDedupTTL(eventType string) (time.Duration, bool) {
	switch eventType {
	case "click":
		return telemetryClickTTL, true
	case "impression":
		return telemetryImpressionTTL, true
	default:
		return 0, false
	}
}
