package http

import (
	"errors"
	"net/http"

	analyticsApp "skykin-platform/internal/analytics/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler serves operator admin analytics endpoints.
type Handler struct {
	svc *analyticsApp.Service
}

func NewHandler(svc *analyticsApp.Service) *Handler {
	return &Handler{svc: svc}
}

// Overview godoc
// @Summary      Platform analytics overview
// @Description  KPI snapshot: advertisers, campaigns, deliveries, subscriptions, MRR estimate, segment revenue.
// @Tags         Ad Portal - Admin Analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  skykin-platform_internal_analytics_domain.OverviewStats
// @Router       /ad-portal/admin/analytics/overview [get]
func (h *Handler) Overview(c *gin.Context) {
	data, err := h.svc.Overview(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "overview failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, data)
}

// Campaigns godoc
// @Summary      Campaign performance table
// @Tags         Ad Portal - Admin Analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  CampaignListResponse
// @Router       /ad-portal/admin/analytics/campaigns [get]
func (h *Handler) Campaigns(c *gin.Context) {
	list, err := h.svc.Campaigns(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "campaigns analytics failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaigns": list, "count": len(list)})
}

// CampaignDetail godoc
// @Summary      Single campaign analytics drill-down
// @Tags         Ad Portal - Admin Analytics
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Campaign ID"
// @Success      200  {object}  skykin-platform_internal_analytics_domain.CampaignDetail
// @Router       /ad-portal/admin/analytics/campaigns/{id} [get]
func (h *Handler) CampaignDetail(c *gin.Context) {
	data, err := h.svc.CampaignDetail(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			platformHTTP.Error(c, http.StatusNotFound, "not found", nil)
			return
		}
		platformHTTP.Error(c, http.StatusInternalServerError, "campaign detail failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, data)
}

// Delivery godoc
// @Summary      Delivery volume and funnel analytics
// @Tags         Ad Portal - Admin Analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  skykin-platform_internal_analytics_domain.DeliveryAnalytics
// @Router       /ad-portal/admin/analytics/delivery [get]
func (h *Handler) Delivery(c *gin.Context) {
	data, err := h.svc.Delivery(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "delivery analytics failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, data)
}

// Revenue godoc
// @Summary      Revenue and subscription analytics
// @Description  MRR estimate, segment purchase revenue, billing events (admin only).
// @Tags         Ad Portal - Admin Analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  skykin-platform_internal_analytics_domain.RevenueOverview
// @Router       /ad-portal/admin/analytics/revenue [get]
func (h *Handler) Revenue(c *gin.Context) {
	data, err := h.svc.Revenue(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "revenue analytics failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, data)
}

// Advertisers godoc
// @Summary      Per-advertiser operational summary
// @Tags         Ad Portal - Admin Analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  AdvertiserListResponse
// @Router       /ad-portal/admin/analytics/advertisers [get]
func (h *Handler) Advertisers(c *gin.Context) {
	list, err := h.svc.Advertisers(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "advertiser analytics failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"advertisers": list, "count": len(list)})
}

// CampaignListResponse for swagger.
type CampaignListResponse struct {
	Campaigns []interface{} `json:"campaigns"`
	Count     int           `json:"count"`
}

// AdvertiserListResponse for swagger.
type AdvertiserListResponse struct {
	Advertisers []interface{} `json:"advertisers"`
	Count       int           `json:"count"`
}
