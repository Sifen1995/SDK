package http

import (
	"context"
	"log/slog"
	"net/http"

	adminApp "skykin-platform/internal/admin/application"
	analyticsApp "skykin-platform/internal/analytics/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler exposes analysis operations (insights only, no entity management).
type AnalyticsHandler struct {
	analyzeUC *analyticsApp.AnalyzeIntentConsistencyUseCase
}

func NewAnalyticsHandler(analyzeUC *analyticsApp.AnalyzeIntentConsistencyUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{analyzeUC: analyzeUC}
}

// TriggerIntentConsistency godoc
// @Summary      Run intent consistency analysis
// @Description  Starts an asynchronous scan of intent signals. Findings are published as events for audience module to persist as candidates.
// @Tags         Ad Portal - Admin Analytics
// @Produce      json
// @Security     BearerAuth
// @Success      202  {object}  map[string]string
// @Failure      401  {object}  platformHTTP.APIError
// @Failure      403  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/analytics/intent-consistency/run [post]
func (h *AnalyticsHandler) TriggerIntentConsistency(c *gin.Context) {
	go func() {
		if err := h.analyzeUC.Run(context.Background()); err != nil {
			slog.Default().Error("intent consistency analysis failed", "error", err)
		}
	}()
	c.JSON(http.StatusAccepted, gin.H{"message": "intent consistency analysis started"})
}

// SegmentCandidateHandler handles admin approval workflow for audience segment candidates.
type SegmentCandidateHandler struct {
	approve *adminApp.ApproveCandidateUseCase
	reject  *adminApp.RejectCandidateUseCase
}

func NewSegmentCandidateHandler(approve *adminApp.ApproveCandidateUseCase, reject *adminApp.RejectCandidateUseCase) *SegmentCandidateHandler {
	return &SegmentCandidateHandler{approve: approve, reject: reject}
}

// ApproveSegmentCandidate godoc
// @Summary      Approve segment candidate and publish segment
// @Description  Marks the candidate approved and asynchronously provisions the audience segment and memberships.
// @Tags         Ad Portal - Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string  true  "Candidate ID"
// @Param        body  body  ApproveSegmentCandidateRequest  true  "Segment details"
// @Success      202  {object}  map[string]string
// @Failure      400  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/audience/segment-candidates/{id}/approve [post]
func (h *SegmentCandidateHandler) ApproveSegmentCandidate(c *gin.Context) {
	var req ApproveSegmentCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	candidateID, err := parseUUIDParam(c, "id")
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid candidate id", nil)
		return
	}
	adminID, err := parsePortalUserID(c)
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid admin id", nil)
		return
	}
	if err := h.approve.Execute(c.Request.Context(), candidateID, adminID, req.Name, req.Description, req.EstimatedCPM); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "approve failed", err.Error())
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message":      "candidate approved; segment provisioning started",
		"candidate_id": candidateID.String(),
	})
}

// RejectSegmentCandidate godoc
// @Summary      Reject segment candidate
// @Tags         Ad Portal - Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string  true  "Candidate ID"
// @Param        body  body  RejectSegmentCandidateRequest  false  "Notes"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/audience/segment-candidates/{id}/reject [post]
func (h *SegmentCandidateHandler) RejectSegmentCandidate(c *gin.Context) {
	var req RejectSegmentCandidateRequest
	_ = c.ShouldBindJSON(&req)
	candidateID, err := parseUUIDParam(c, "id")
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid candidate id", nil)
		return
	}
	adminID, err := parsePortalUserID(c)
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid admin id", nil)
		return
	}
	if err := h.reject.Execute(c.Request.Context(), candidateID, adminID, req.Notes); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "reject failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "candidate rejected"})
}
