package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	analyticsdomain "skykin-platform/internal/analytics/domain"
	intentApp "skykin-platform/internal/intents/application"
	"skykin-platform/internal/intents/domain"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

type intentAdIngestor interface {
	IngestAndFetchAd(ctx context.Context, profile *domain.IntentProfile, channelCode string, smsConsented bool) (*intentApp.IngestAndFetchAdResult, error)
}

type intentAggregateReporter interface {
	EnqueueReport(ctx context.Context, report *analyticsdomain.AggregateReport) error
}

// Handler exposes SDK intent HTTP endpoints.
type Handler struct {
	ingest     intentAdIngestor
	aggregates intentAggregateReporter
}

func NewHandler(svc *intentApp.IntentService, aggregates intentAggregateReporter) *Handler {
	return &Handler{ingest: svc, aggregates: aggregates}
}

// IngestIntentAd godoc
// @Summary      Ingest intent profile and fetch matching ad
// @Description  Flutter sends an on-device ML intent (from accessibility + app usage). When sms_consented is true and an SMS_PLUS campaign matches, returns 202 after mock/real SMS dispatch (no in-app ad body). Otherwise returns 200 with a non-SMS creative. Authorize with X-API-Key (pk_live_...) and X-SDK-Secret (sk_secret_...); Swagger UI auto-computes X-Signature.
// @Tags         SDK - Intents
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Param        body  body  IngestIntentAdRequest  true  "Intent profile + optional channel + sms_consented"
// @Success      200   {object}  IngestIntentAdResponse
// @Success      202   {object}  IngestIntentAdAcceptedResponse
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

	channelCode := normalizeChannelCode(strings.TrimSpace(req.ChannelCode))

	result, err := h.ingest.IngestAndFetchAd(c.Request.Context(), &domain.IntentProfile{
		PseudonymousID: strings.TrimSpace(req.PseudonymousID),
		IntentName:     strings.TrimSpace(req.IntentName),
		Confidence:     req.Confidence,
		ModelVersion:   strings.TrimSpace(req.ModelVersion),
	}, channelCode, req.SMSConsented)
	if err != nil {
		msg := err.Error()
		status := http.StatusInternalServerError
		switch {
		case strings.Contains(msg, "required") || strings.Contains(msg, "must be between"):
			status = http.StatusBadRequest
		case strings.Contains(msg, "no active campaign"), strings.Contains(msg, "no eligible campaign"):
			status = http.StatusNotFound
		}
		platformHTTP.Error(c, status, "intent ingest failed", msg)
		return
	}

	if result.SMSDispatched {
		c.JSON(http.StatusAccepted, IngestIntentAdAcceptedResponse{
			Status:       "accepted",
			CampaignID:   result.CampaignID,
			CampaignName: result.CampaignName,
			ChannelCode:  result.ChannelCode,
		})
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

// IngestIntentAggregate godoc
// @Summary      Ingest anonymous intent aggregates
// @Description  Accepts a device batch of anonymized intent counters for non-consented users. Enqueues to Redis (queue:analytics_aggregate) for async upsert into intent_aggregate_counts (signal_count += count, weighted_count += days_consistent). Does not select or return ads. Authorize with X-API-Key and X-SDK-Secret.
// @Tags         SDK - Intents
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Param        body  body  IngestIntentAggregateRequest  true  "Anonymous aggregate batch"
// @Success      202  "Accepted — batch queued"
// @Failure      400  {object}  platformHTTP.APIError
// @Failure      401  {object}  platformHTTP.APIError
// @Failure      503  {object}  platformHTTP.APIError
// @Router       /intents/ingest-aggregate [post]
func (h *Handler) IngestIntentAggregate(c *gin.Context) {
	if h == nil || h.aggregates == nil {
		platformHTTP.Error(c, http.StatusServiceUnavailable, "aggregate ingest unavailable", "")
		return
	}

	var req IngestIntentAggregateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid aggregate payload", err.Error())
		return
	}

	bucket, err := time.Parse("2006-01-02", strings.TrimSpace(req.DateBucket))
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid date_bucket", "expected YYYY-MM-DD")
		return
	}

	report := &analyticsdomain.AggregateReport{
		DateBucket: bucket.UTC(),
		Intents:    make([]analyticsdomain.AggregateIntentSignal, len(req.Intents)),
	}
	for i, item := range req.Intents {
		report.Intents[i] = analyticsdomain.AggregateIntentSignal{
			IntentName:     strings.TrimSpace(item.IntentName),
			Count:          item.Count,
			DaysConsistent: item.DaysConsistent,
		}
	}

	if err := h.aggregates.EnqueueReport(c.Request.Context(), report); err != nil {
		msg := err.Error()
		status := http.StatusServiceUnavailable
		if strings.Contains(msg, "required") || strings.Contains(msg, "must") || strings.Contains(msg, "empty") {
			status = http.StatusBadRequest
		}
		platformHTTP.Error(c, status, "aggregate ingest failed", msg)
		return
	}

	c.Status(http.StatusAccepted)
}

func normalizeChannelCode(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "SMS":
		return "SMS_PLUS"
	default:
		return code
	}
}
