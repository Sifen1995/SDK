package http

import (
	"errors"
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
// @Description  Scans intent signals, merges new users into existing matching segments, refreshes pending candidates, or creates new candidates only when needed.
// @Tags         Ad Portal - Admin Analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  analyticsApp.RunReport
// @Failure      401  {object}  platformHTTP.APIError
// @Failure      403  {object}  platformHTTP.APIError
// @Failure      500  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/analytics/intent-consistency/run [post]
func (h *AnalyticsHandler) TriggerIntentConsistency(c *gin.Context) {
	report, err := h.analyzeUC.Run(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "intent consistency analysis failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, report)
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
// @Description  Publishes the audience segment and its memberships in a single transaction and returns the published segment id. Fails with 409 when the candidate has already been reviewed.
// @Tags         Ad Portal - Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string  true  "Candidate ID"
// @Param        body  body  ApproveSegmentCandidateRequest  true  "Segment details"
// @Success      201  {object}  ApproveSegmentCandidateResponse
// @Failure      400  {object}  platformHTTP.APIError
// @Failure      409  {object}  platformHTTP.APIError
// @Failure      500  {object}  platformHTTP.APIError
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
	segment, err := h.approve.Execute(c.Request.Context(), candidateID, adminID, req.Name, req.Description, req.EstimatedCPM)
	if err != nil {
		if errors.Is(err, adminApp.ErrCandidateNotPending) {
			platformHTTP.Error(c, http.StatusConflict, "candidate already reviewed", err.Error())
			return
		}
		platformHTTP.Error(c, http.StatusBadRequest, "approve failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, ApproveSegmentCandidateResponse{
		Message:     "segment published from candidate",
		CandidateID: candidateID.String(),
		SegmentID:   segment.SegmentID,
		MemberCount: segment.MemberCount,
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
// @Failure      409  {object}  platformHTTP.APIError
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
		if errors.Is(err, adminApp.ErrCandidateNotPending) {
			platformHTTP.Error(c, http.StatusConflict, "candidate already reviewed", err.Error())
			return
		}
		platformHTTP.Error(c, http.StatusBadRequest, "reject failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "candidate rejected"})
}
