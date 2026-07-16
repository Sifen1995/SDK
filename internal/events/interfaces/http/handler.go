package http

import (
	"net/http"
	"time"

	"skykin-platform/internal/events/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP endpoints for the events module.
type Handler struct {
	ingest *application.IngestEventsUseCase
}

func NewHandler(ingest *application.IngestEventsUseCase) *Handler {
	return &Handler{ingest: ingest}
}

// PostEvents handles SDK batch event ingestion.
// Not mounted while event ingestion is disabled; handler + use case kept for later reuse.
func (h *Handler) PostEvents(c *gin.Context) {
	appID, exists := c.Get("application_id")
	if !exists {
		platformHTTP.Error(c, http.StatusUnauthorized, "missing application context", nil)
		return
	}

	var req IngestEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid event batch payload", err.Error())
		return
	}

	cmd := application.IngestCommand{
		ApplicationID:  appID.(string),
		ExternalUserID: req.UserID,
		Events:         toApplicationInputs(req.Events),
	}

	result, err := h.ingest.Execute(c.Request.Context(), cmd)
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "event ingestion failed", err.Error())
		return
	}

	c.JSON(http.StatusAccepted, toResponseDTO(result))
}

func toApplicationInputs(items []EventInput) []application.EventInput {
	out := make([]application.EventInput, len(items))
	for i, item := range items {
		var createdAt time.Time
		if item.CreatedAt != nil {
			createdAt = item.CreatedAt.UTC()
		}
		out[i] = application.EventInput{
			EventID:    item.EventID,
			EventType:  item.EventType,
			Domain:     item.Domain,
			SessionID:  item.SessionID,
			ScreenName: item.ScreenName,
			Metadata:   item.Metadata,
			DeviceType: item.DeviceType,
			Platform:   item.Platform,
			AppVersion: item.AppVersion,
			CreatedAt:  createdAt,
		}
	}
	return out
}

func toResponseDTO(result *application.IngestResult) IngestEventsResponse {
	results := make([]EventIngestResultDTO, len(result.Results))
	for i, r := range result.Results {
		results[i] = EventIngestResultDTO{EventID: r.EventID, Status: r.Status}
	}
	return IngestEventsResponse{
		Accepted:         result.Accepted,
		PredictionQueued: result.PredictionQueued,
		Results:          results,
	}
}
