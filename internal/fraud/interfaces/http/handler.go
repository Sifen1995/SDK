package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	fraudapp "skykin-platform/internal/fraud/application"
	frauddomain "skykin-platform/internal/fraud/domain"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

type syncExecutor interface {
	Execute(ctx context.Context, since *time.Time) (*frauddomain.SyncSnapshot, error)
}

type reportExecutor interface {
	Execute(
		ctx context.Context,
		command fraudapp.IngestReportCommand,
	) (*fraudapp.IngestReportResult, error)
}

type Handler struct {
	sync   syncExecutor
	report reportExecutor
}

func NewHandler(sync syncExecutor, report ...reportExecutor) *Handler {
	handler := &Handler{sync: sync}
	if len(report) > 0 {
		handler.report = report[0]
	}
	return handler
}

// Report godoc
// @Summary      Submit an anonymous threat report
// @Description  Persists one normalized threat report and stages it for one-hour threshold aggregation. No user or pseudonymous identifier is collected.
// @Tags         SDK - Fraud
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Param        request  body  ThreatReportRequest  true  "Threat report"
// @Success      202  {object}  ThreatReportAcceptedResponse
// @Failure      400  {object}  platformHTTP.APIError
// @Failure      401  {object}  platformHTTP.APIError
// @Failure      500  {object}  platformHTTP.APIError
// @Failure      503  {object}  platformHTTP.APIError
// @Router       /reports [post]
func (h *Handler) Report(c *gin.Context) {
	var request ThreatReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid threat report", err.Error())
		return
	}
	if h.report == nil {
		platformHTTP.Error(c, http.StatusServiceUnavailable, "threat report ingestion unavailable", nil)
		return
	}

	result, err := h.report.Execute(c.Request.Context(), fraudapp.IngestReportCommand{
		ThreatType:      request.ThreatType,
		Severity:        request.Severity,
		SenderHash:      request.SenderHash,
		URLDomain:       request.URLDomain,
		DetectionSource: request.DetectionSource,
		SDKVersion:      request.SDKVersion,
	})
	if err != nil {
		switch {
		case errors.Is(err, fraudapp.ErrInvalidReport):
			platformHTTP.Error(c, http.StatusBadRequest, "invalid threat report", err.Error())
		case errors.Is(err, fraudapp.ErrQueueUnavailable):
			details := map[string]any{"persisted": result != nil}
			if result != nil {
				details["report_id"] = result.ReportID
			}
			platformHTTP.Error(c, http.StatusServiceUnavailable, "threat report staging failed", details)
		default:
			platformHTTP.Error(c, http.StatusInternalServerError, "threat report ingestion failed", nil)
		}
		return
	}

	c.JSON(http.StatusAccepted, ThreatReportAcceptedResponse{
		Status:     "accepted",
		ReportID:   result.ReportID,
		ReportedAt: formatTime(result.ReportedAt),
	})
}

// Sync godoc
// @Summary      Sync fraud intelligence
// @Description  Returns a full active snapshot when since is omitted. With an RFC3339 since cursor, returns changed rows including revoked, expired, and inactive tombstones. Store next_cursor and supply it on the next poll.
// @Tags         SDK - Fraud
// @Produce      json
// @Security     APIKeyAuth
// @Param        since  query  string  false  "Previous next_cursor (RFC3339/RFC3339Nano)"
// @Success      200  {object}  SyncResponse
// @Failure      400  {object}  platformHTTP.APIError
// @Failure      401  {object}  platformHTTP.APIError
// @Failure      500  {object}  platformHTTP.APIError
// @Router       /sync [get]
func (h *Handler) Sync(c *gin.Context) {
	var since *time.Time
	rawSince := strings.TrimSpace(c.Query("since"))
	if rawSince != "" {
		parsed, err := time.Parse(time.RFC3339Nano, rawSince)
		if err != nil {
			platformHTTP.Error(
				c,
				http.StatusBadRequest,
				"invalid since cursor",
				"since must be an RFC3339 timestamp returned as next_cursor",
			)
			return
		}
		since = &parsed
	}

	snapshot, err := h.sync.Execute(c.Request.Context(), since)
	if err != nil {
		if errors.Is(err, fraudapp.ErrFutureCursor) {
			platformHTTP.Error(c, http.StatusBadRequest, "invalid since cursor", err.Error())
			return
		}
		platformHTTP.Error(c, http.StatusInternalServerError, "fraud sync failed", nil)
		return
	}

	c.JSON(http.StatusOK, mapSyncResponse(snapshot))
}

func mapSyncResponse(snapshot *frauddomain.SyncSnapshot) SyncResponse {
	mode := "full"
	if snapshot.IsDelta {
		mode = "delta"
	}
	response := SyncResponse{
		Status:         "success",
		Mode:           mode,
		NextCursor:     formatTime(snapshot.NextCursor),
		BlockedDomains: make([]BlockedDomainDTO, 0, len(snapshot.BlockedDomains)),
		BlockedSenders: make([]BlockedSenderDTO, 0, len(snapshot.BlockedSenders)),
		ScamPatterns:   make([]ScamPatternDTO, 0, len(snapshot.ScamPatterns)),
	}

	for _, value := range snapshot.BlockedDomains {
		var expiresAt *string
		if value.ExpiresAt != nil {
			formatted := formatTime(*value.ExpiresAt)
			expiresAt = &formatted
		}
		response.BlockedDomains = append(response.BlockedDomains, BlockedDomainDTO{
			Domain:     value.Domain,
			ThreatType: value.ThreatType,
			Severity:   value.Severity,
			Source:     value.Source,
			Status:     value.Status,
			CreatedAt:  formatTime(value.CreatedAt),
			ExpiresAt:  expiresAt,
			UpdatedAt:  formatTime(value.UpdatedAt),
		})
	}
	for _, value := range snapshot.BlockedSenders {
		response.BlockedSenders = append(response.BlockedSenders, BlockedSenderDTO{
			SenderHash: value.SenderHash,
			ThreatType: value.ThreatType,
			Severity:   value.Severity,
			Source:     value.Source,
			Status:     value.Status,
			CreatedAt:  formatTime(value.CreatedAt),
			UpdatedAt:  formatTime(value.UpdatedAt),
		})
	}
	for _, value := range snapshot.ScamPatterns {
		response.ScamPatterns = append(response.ScamPatterns, ScamPatternDTO{
			ID:             value.ID,
			PatternType:    value.PatternType,
			PatternValue:   value.PatternValue,
			ThreatCategory: value.ThreatCategory,
			Confidence:     value.Confidence,
			Language:       value.Language,
			IsActive:       value.IsActive,
			UpdatedAt:      formatTime(value.UpdatedAt),
		})
	}
	return response
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
