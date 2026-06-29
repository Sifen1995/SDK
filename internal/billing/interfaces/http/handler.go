package http

import (
	"errors"
	"net/http"
	"strings"

	billingApp "skykin-platform/internal/billing/application"
	platformHTTP "skykin-platform/internal/platform/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler serves billing and subscription endpoints on the ad portal.
type Handler struct {
	svc *billingApp.SubscriptionService
}

func NewHandler(svc *billingApp.SubscriptionService) *Handler {
	return &Handler{svc: svc}
}

// ListPlans godoc
// @Summary      List active subscription plans
// @Description  Returns active plans available for advertisers to choose before subscribing or creating campaigns.
// @Tags         Ad Portal - Billing
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  PlanListResponse
// @Router       /ad-portal/plans [get]
func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list plans failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans, "count": len(plans)})
}

// ListChannels godoc
// @Summary      List delivery channels
// @Description  Returns active delivery channels (IN_APP_BANNER, PUSH, SMS_PLUS, etc.) for campaign creation.
// @Tags         Ad Portal - Billing
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  ChannelListResponse
// @Router       /ad-portal/channels [get]
func (h *Handler) ListChannels(c *gin.Context) {
	channels, err := h.svc.ListChannels(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list channels failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"channels": channels, "count": len(channels)})
}

// GetSubscription godoc
// @Summary      Get current subscription
// @Description  Returns the advertiser active subscription, or subscribed=false when none.
// @Tags         Ad Portal - Billing
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  SubscriptionStatusResponse
// @Router       /ad-portal/subscription [get]
func (h *Handler) GetSubscription(c *gin.Context) {
	aid, _ := c.Get("advertiser_id")
	sub, err := h.svc.GetSubscription(c.Request.Context(), aid.(string))
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "subscription lookup failed", err.Error())
		return
	}
	if sub == nil {
		c.JSON(http.StatusOK, gin.H{"subscribed": false, "subscription": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscribed": true, "subscription": sub})
}

// Subscribe godoc
// @Summary      Subscribe to a plan
// @Description  Creates an active subscription for the advertiser. Required before campaign creation.
// @Tags         Ad Portal - Billing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      SubscribeRequest  true  "Plan selection"
// @Success      201   {object}  SubscriptionStatusResponse
// @Failure      400   {object}  platformHTTP.APIError
// @Failure      409   {object}  platformHTTP.APIError
// @Router       /ad-portal/subscription [post]
func (h *Handler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	aid, _ := c.Get("advertiser_id")
	sub, err := h.svc.Subscribe(c.Request.Context(), aid.(string), req.PlanID)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "already subscribed") {
			status = http.StatusConflict
		}
		platformHTTP.Error(c, status, "subscribe failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"subscribed": true, "subscription": sub})
}

// SubscribeRequest is the body for POST /subscription.
type SubscribeRequest struct {
	PlanID string `json:"plan_id" binding:"required,uuid"`
}

// PlanListResponse for swagger.
type PlanListResponse struct {
	Plans []billingApp.PlanDTO `json:"plans"`
	Count int                  `json:"count"`
}

// ChannelListResponse for swagger.
type ChannelListResponse struct {
	Channels []billingApp.ChannelDTO `json:"channels"`
	Count    int                     `json:"count"`
}

// SubscriptionStatusResponse for swagger.
type SubscriptionStatusResponse struct {
	Subscribed   bool                        `json:"subscribed"`
	Subscription *billingApp.SubscriptionDTO `json:"subscription"`
}
