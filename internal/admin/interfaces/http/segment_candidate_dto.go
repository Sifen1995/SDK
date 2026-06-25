package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ApproveSegmentCandidateRequest struct {
	Name         string  `json:"name" binding:"required" example:"Coffee Lovers"`
	Description  string  `json:"description" example:"Users with sustained coffee purchase intent"`
	EstimatedCPM float64 `json:"estimated_cpm" binding:"required,gt=0" example:"4.5"`
}

type RejectSegmentCandidateRequest struct {
	Notes string `json:"notes" example:"Insufficient user volume for this vertical"`
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, error) {
	return uuid.Parse(c.Param(name))
}

func parsePortalUserID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString("portal_user_id"))
}
