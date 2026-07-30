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

type ApproveSegmentCandidateResponse struct {
	Message     string `json:"message" example:"segment published from candidate"`
	CandidateID string `json:"candidate_id" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"`
	SegmentID   string `json:"segment_id" example:"3f2504e0-4f89-11d3-9a0c-0305e82c3301"`
	MemberCount int    `json:"member_count" example:"1240"`
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
