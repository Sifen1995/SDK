package http

import (
	"net/http"
	"time"

	adminApp "skykin-platform/internal/admin/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
)

// CatalogHandler exposes operator plan, segment, and billing rate management.
type CatalogHandler struct {
	catalog *adminApp.PlanAndSegmentService
	billing *adminApp.BillingAdminService
}

func NewCatalogHandler(catalog *adminApp.PlanAndSegmentService, billing *adminApp.BillingAdminService) *CatalogHandler {
	return &CatalogHandler{catalog: catalog, billing: billing}
}

// CreatePlan godoc
// @Summary      Create subscription plan
// @Description  Creates a plan and seeds default billing rates (CPM/CPC/CPI/CPA/REV_SHARE).
// @Tags         Ad Portal - Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreatePlanRequest  true  "Plan"
// @Success      201   {object}  skykin-platform_internal_billing_model.SubscriptionPlan
// @Failure      400   {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/plans [post]
func (h *CatalogHandler) CreatePlan(c *gin.Context) {
	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	plan, err := h.catalog.CreatePlan(c.Request.Context(), adminApp.CreatePlanCmd{
		Name:                req.Name,
		MonthlyFeeETB:       req.MonthlyFeeETB,
		MaxActiveCampaigns:  req.MaxActiveCampaigns,
		MaxDailyBudgetETB:   req.MaxDailyBudgetETB,
		IncludedImpressions: req.IncludedImpressions,
		SMSPlusEnabled:      req.SMSPlusEnabled,
		AudiencemartEnabled: req.AudiencemartEnabled,
		CPCDiscountPct:      req.CPCDiscountPct,
	})
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "create plan failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, plan)
}

// CreateSegment godoc
// @Summary      Create audience segment
// @Description  Adds a new Audiencemart catalog segment.
// @Tags         Ad Portal - Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateSegmentRequest  true  "Segment"
// @Success      201   {object}  skykin-platform_internal_audience_model.AudienceSegment
// @Failure      400   {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/audience/segments [post]
func (h *CatalogHandler) CreateSegment(c *gin.Context) {
	var req CreateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	var availableFrom *time.Time
	if req.AvailableFrom != nil {
		t := req.AvailableFrom.UTC()
		availableFrom = &t
	}
	seg, err := h.catalog.CreateSegment(c.Request.Context(), adminApp.CreateSegmentCmd{
		Name:             req.Name,
		Description:      req.Description,
		TopIntentSignals: req.TopIntentSignals,
		ApproximateSize:  req.ApproximateSize,
		EstimatedCPM:     req.EstimatedCPM,
		AvailableFrom:    availableFrom,
		AvailableUntil:   req.AvailableUntil,
		IsActive:         req.IsActive,
	})
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "create segment failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, seg)
}

// ListSegments godoc
// @Summary      List audience catalog segments
// @Description  Returns all active Audiencemart segments for operator admin review.
// @Tags         Ad Portal - Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  SegmentListResponse
// @Failure      500  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/audience/segments [get]
func (h *CatalogHandler) ListSegments(c *gin.Context) {
	segments, err := h.catalog.ListSegments(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list segments failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, SegmentListResponse{Segments: segments, Count: len(segments)})
}

// ListBillingRates godoc
// @Summary      List billing rates for a plan
// @Tags         Ad Portal - Admin
// @Produce      json
// @Security     BearerAuth
// @Param        plan_id  path  string  true  "Plan ID"
// @Success      200  {object}  BillingRateListResponse
// @Failure      404  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/plans/{plan_id}/billing-rates [get]
func (h *CatalogHandler) ListBillingRates(c *gin.Context) {
	rates, err := h.billing.ListBillingRates(c.Request.Context(), c.Param("plan_id"))
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "plan not found" {
			status = http.StatusNotFound
		}
		platformHTTP.Error(c, status, "list billing rates failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"rates": rates, "count": len(rates)})
}

// UpdateBillingRate godoc
// @Summary      Update a billing rate
// @Tags         Ad Portal - Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string  true  "Billing rate ID"
// @Param        body  body  UpdateBillingRateRequest  true  "Rate update"
// @Success      200  {object}  skykin-platform_internal_billing_model.BillingRate
// @Failure      400  {object}  platformHTTP.APIError
// @Failure      404  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/billing-rates/{id} [patch]
func (h *CatalogHandler) UpdateBillingRate(c *gin.Context) {
	var req UpdateBillingRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	rate, err := h.billing.UpdateBillingRate(c.Request.Context(), adminApp.UpdateBillingRateCmd{
		RateID:   c.Param("id"),
		RateETB:  req.RateETB,
		IsActive: req.IsActive,
	})
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "billing rate not found" {
			status = http.StatusNotFound
		}
		platformHTTP.Error(c, status, "update billing rate failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, rate)
}
