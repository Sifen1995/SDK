package http

import (
	"context"
	"net/http"
	"strings"

	intentApp "skykin-platform/internal/intents/application"
	"skykin-platform/internal/intents/domain"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

type intentAdIngestor interface {
	IngestAndFetchAd(ctx context.Context, profile *domain.IntentProfile, channelCode string) (*intentApp.IngestAndFetchAdResult, error)
}

// Handler exposes SDK intent HTTP endpoints.
type Handler struct {
	ingest intentAdIngestor
}

func NewHandler(svc *intentApp.IntentService) *Handler {
	return &Handler{ingest: svc}
}

// IngestIntentAd godoc
// @Summary      Ingest intent profile and fetch matching ad
// @Description  Flutter sends an on-device ML intent (from accessibility + app usage). Backend caches the active intent, persists the profile, and returns a campaign creative for the requested channel. Authorize with X-API-Key (pk_live_...) and X-SDK-Secret (sk_secret_...); Swagger UI auto-computes X-Signature.
// @Tags         SDK - Intents
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Param        body  body  IngestIntentAdRequest  true  "Intent profile + optional channel"
// @Success      200   {object}  IngestIntentAdResponse
// @Failure      400   {object}  platformHTTP.APIError
// @Failure      401   {object}  platformHTTP.APIError
// @Failure      404   {object}  platformHTTP.APIError
// @Failure      500   {object}  platformHTTP.APIError
// @Router       /intents/ingest-ad [post]
func (h *Handler) IngestIntentAd(c *gin.Context) {
	var req IngestIntentAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid intent payload", err.Error())
		return
	}

	result, err := h.ingest.IngestAndFetchAd(c.Request.Context(), &domain.IntentProfile{
		PseudonymousID: strings.TrimSpace(req.PseudonymousID),
		IntentName:     strings.TrimSpace(req.IntentName),
		Confidence:     req.Confidence,
		ModelVersion:   strings.TrimSpace(req.ModelVersion),
	}, strings.TrimSpace(req.ChannelCode))
	if err != nil {
		msg := err.Error()
		status := http.StatusInternalServerError
		switch {
		case strings.Contains(msg, "required") || strings.Contains(msg, "must be between"):
			status = http.StatusBadRequest
		case strings.Contains(msg, "no active campaign"):
			status = http.StatusNotFound
		}
		platformHTTP.Error(c, status, "intent ingest failed", msg)
		return
	}

	c.JSON(http.StatusOK, IngestIntentAdResponse{
		PseudonymousID: result.PseudonymousID,
		IntentName:     result.IntentName,
		Confidence:     result.Confidence,
		ModelVersion:   result.ModelVersion,
		CampaignID:     result.CampaignID,
		CampaignName:   result.CampaignName,
		ChannelCode:    result.ChannelCode,
		AdContent:      result.AdContent,
	})
}
