package http

import (
	"time"

	billingdomain "skykin-platform/internal/billing/domain"
)

// ValidateCampaignRequest is the operator approve/reject payload.
type ValidateCampaignRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Notes  string `json:"notes" binding:"max=2000"`
}

// CampaignListResponse wraps campaigns for admin list endpoints.
type CampaignListResponse struct {
	Campaigns []interface{} `json:"campaigns"`
	Count     int           `json:"count"`
}

// CreatePlanRequest is the HTTP body for POST /admin/plans.
type CreatePlanRequest struct {
	Name                string  `json:"name" binding:"required,min=2,max=100"`
	MonthlyFeeETB       float64 `json:"monthly_fee_etb" binding:"required,gte=0"`
	MaxActiveCampaigns  int     `json:"max_active_campaigns" binding:"required,min=1"`
	MaxDailyBudgetETB   float64 `json:"max_daily_budget_etb" binding:"required,gt=0"`
	IncludedImpressions int   `json:"included_impressions" binding:"gte=0"`
	SMSPlusEnabled      bool    `json:"sms_plus_enabled"`
	AudiencemartEnabled bool    `json:"audiencemart_enabled"`
	CPCDiscountPct      float64 `json:"cpc_discount_pct" binding:"gte=0,lte=100"`
}

// CreateSegmentRequest is the HTTP body for POST /admin/audience/segments.
type CreateSegmentRequest struct {
	Name             string     `json:"name" binding:"required,min=2,max=255"`
	Description      string     `json:"description" binding:"max=2000"`
	TopIntentSignals []string   `json:"top_intent_signals" binding:"required,min=1,dive,required"`
	ApproximateSize  int        `json:"approximate_size" binding:"gte=0"`
	EstimatedCPM     float64    `json:"estimated_cpm" binding:"required,gt=0"`
	AvailableFrom    *time.Time `json:"available_from"`
	AvailableUntil   *time.Time `json:"available_until"`
	IsActive         bool       `json:"is_active"`
}

// UpdateBillingRateRequest is the HTTP body for PATCH /admin/billing-rates/:id.
type UpdateBillingRateRequest struct {
	RateETB  float64 `json:"rate_etb" binding:"required,gte=0"`
	IsActive bool    `json:"is_active"`
}

// UpdatePlanRequest is the HTTP body for PATCH /admin/plans/:plan_id.
type UpdatePlanRequest struct {
	Name                string  `json:"name" binding:"required,min=2,max=100"`
	MonthlyFeeETB       float64 `json:"monthly_fee_etb" binding:"required,gte=0"`
	MaxActiveCampaigns  int     `json:"max_active_campaigns" binding:"required,min=1"`
	MaxDailyBudgetETB   float64 `json:"max_daily_budget_etb" binding:"required,gt=0"`
	IncludedImpressions int     `json:"included_impressions" binding:"gte=0"`
	SMSPlusEnabled      bool    `json:"sms_plus_enabled"`
	AudiencemartEnabled bool    `json:"audiencemart_enabled"`
	CPCDiscountPct      float64 `json:"cpc_discount_pct" binding:"gte=0,lte=100"`
	IsActive            bool    `json:"is_active"`
}

// BillingRateListResponse for swagger.
type BillingRateListResponse struct {
	Rates []interface{} `json:"rates"`
	Count int           `json:"count"`
}

// AdminPlanListResponse wraps all subscription plans for operator admin.
type AdminPlanListResponse struct {
	Plans []billingdomain.SubscriptionPlan `json:"plans"`
	Count int                             `json:"count"`
}
