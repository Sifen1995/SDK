package controller

import (
	"net/http"
	"skykin-platform/internal/common/response"
	"skykin-platform/internal/events/dto"
	"skykin-platform/internal/events/service"

	"github.com/gin-gonic/gin"
)

type EventController struct {
	service service.EventServiceInterface
}

func NewEventController(s service.EventServiceInterface) *EventController {
	return &EventController{service: s}
}

// PostEvent godoc
// @Summary      Ingest a user event
// @Description  Receives an SDK event, persists it, runs ML intent prediction, and triggers rewards if the confidence threshold is met. Requires HMAC signature via X-Signature header.
// @Tags         SDK - Events
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth
// @Param        X-Signature  header    string              false  "HMAC-SHA256 signature of the request body"
// @Param        body         body      dto.EventRequestDTO  true   "Event payload"
// @Success      201          {object}  dto.EventResponseDTO
// @Success      202          {object}  dto.EventResponseDTO  "Event queued — cold start or ML unavailable"
// @Failure      400          {object}  response.APIError
// @Failure      401          {object}  response.APIError
// @Failure      500          {object}  response.APIError
// @Router       /events [post]
func (ctrl *EventController) PostEvent(c *gin.Context) {
	appID, exists := c.Get("application_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Missing verified application context scope", nil)
		return
	}

	var input dto.EventRequestDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Event request payload validation failed", err.Error())
		return
	}

	res, statusCode, err := ctrl.service.ProcessEvent(c.Request.Context(), appID.(string), &input)
	if err != nil {
		response.Error(c, statusCode, "Transaction failed to complete cleanly", err.Error())
		return
	}

	c.JSON(statusCode, res)
}

// PostBatchEvents godoc
// @Summary      Ingest batched user events
// @Description  Receives a batch of events from the SDK (collected over 30s or on app background), persists them all, runs one ML intent prediction on the session, and returns the predicted intent and reward if triggered.
// @Tags         SDK - Events
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth
// @Param        X-Signature  header    string                     false  "HMAC-SHA256 signature of the request body"
// @Param        body         body      dto.BatchEventRequestDTO   true   "Batched events payload"
// @Success      201          {object}  dto.BatchEventResponseDTO
// @Success      202          {object}  dto.BatchEventResponseDTO  "Events stored — cold start or ML unavailable"
// @Failure      400          {object}  response.APIError
// @Failure      401          {object}  response.APIError
// @Failure      500          {object}  response.APIError
// @Router       /events/batch [post]
func (ctrl *EventController) PostBatchEvents(c *gin.Context) {
	appID, exists := c.Get("application_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Missing verified application context scope", nil)
		return
	}

	var input dto.BatchEventRequestDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Batch event payload validation failed", err.Error())
		return
	}

	res, statusCode, err := ctrl.service.ProcessBatchEvents(c.Request.Context(), appID.(string), &input)
	if err != nil {
		response.Error(c, statusCode, "Batch processing failed", err.Error())
		return
	}

	c.JSON(statusCode, res)
}
