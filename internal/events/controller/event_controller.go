package controller

import (
	"net/http"
	"skykin-platform/internal/events/model"
	"skykin-platform/internal/events/service"
	"time"

	"github.com/gin-gonic/gin"
)

type EventController struct {
	service service.EventServiceInterface
}

func NewEventController(s service.EventServiceInterface) *EventController {
	return &EventController{service: s}
}

func (ctrl *EventController) PostEvent(c *gin.Context) {
	var input struct {
		ExternalUserID string                 `json:"external_user_id" binding:"required"`
		EventType      string                 `json:"event_type" binding:"required"`
		Metadata       map[string]interface{} `json:"metadata"`
		Timestamp      string                 `json:"timestamp" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"` // ISO8601 format
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse timestamp and map to model
	var eventTime time.Time
	if input.Timestamp == "" {
		// Default to current time if not provided
		eventTime = time.Now()
	} else {
		// Try to parse provided timestamp
		t, err := time.Parse(time.RFC3339, input.Timestamp)
		if err != nil {
			// Fallback to now if format is wrong
			eventTime = time.Now()
		} else {
			eventTime = t
		}
	}
	event := &model.Event{
		EventType: input.EventType,
		Metadata:  input.Metadata,
		Timestamp: eventTime,
	}

	response, statusCode, err := ctrl.service.ProcessEvent(c.Request.Context(), input.ExternalUserID, event)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, response)
}
