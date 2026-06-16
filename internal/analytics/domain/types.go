package domain

import (
	"context"
	"time"
)

// OverviewStats is the platform-wide KPI snapshot for the admin dashboard.
type OverviewStats struct {
	TotalAdvertisers       int64   `json:"total_advertisers"`
	TotalCampaigns         int64   `json:"total_campaigns"`
	ActiveCampaigns        int64   `json:"active_campaigns"`
	PendingModeration      int64   `json:"pending_moderation"`
	TotalDeliveries        int64   `json:"total_deliveries"`
	DeliveriesLast24h      int64   `json:"deliveries_last_24h"`
	DeliveriesLast7d       int64   `json:"deliveries_last_7d"`
	ActiveSubscriptions    int64   `json:"active_subscriptions"`
	EstimatedMRR_ETB       float64 `json:"estimated_mrr_etb"`
	SegmentRevenueTotalETB float64 `json:"segment_revenue_total_etb"`
	UniqueUsersReached     int64   `json:"unique_users_reached"`
}

// PlanCount is subscription count grouped by plan name.
type PlanCount struct {
	PlanName string `json:"plan_name"`
	Count    int64  `json:"count"`
}

// CampaignPerformance is per-campaign delivery and status metrics.
type CampaignPerformance struct {
	CampaignID       string  `json:"campaign_id"`
	Name             string  `json:"name"`
	AdvertiserID     string  `json:"advertiser_id"`
	CompanyName      string  `json:"company_name"`
	TargetIntent     string  `json:"target_intent"`
	IsActive         bool    `json:"is_active"`
	ModerationStatus string  `json:"moderation_status"`
	ValidationStatus string  `json:"validation_status"`
	DeliveryCount    int64   `json:"delivery_count"`
	UniqueUsers      int64   `json:"unique_users"`
	BudgetSpent      float64 `json:"budget_spent"`
	DailyBudgetCap   float64 `json:"daily_budget_cap"`
}

// CampaignDetail extends performance with funnel breakdown.
type CampaignDetail struct {
	CampaignPerformance
	Funnel []FunnelStep `json:"funnel"`
}

// FunnelStep counts delivery lifecycle events from campaign_delivery_logs.
type FunnelStep struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// DeliveryTrendPoint is deliveries per calendar day.
type DeliveryTrendPoint struct {
	Day   string `json:"day"`
	Count int64  `json:"count"`
}

// DeliveryAnalytics aggregates dispatch volume and trends.
type DeliveryAnalytics struct {
	TotalDeliveries   int64                `json:"total_deliveries"`
	Last30Days        []DeliveryTrendPoint `json:"last_30_days"`
	TopCampaigns      []CampaignPerformance `json:"top_campaigns"`
	FunnelPlatform    []FunnelStep         `json:"funnel_platform"`
}

// RevenueOverview is admin-only financial snapshot.
type RevenueOverview struct {
	EstimatedMRR_ETB       float64     `json:"estimated_mrr_etb"`
	SegmentRevenueTotalETB float64     `json:"segment_revenue_total_etb"`
	SegmentRevenue30dETB   float64     `json:"segment_revenue_30d_etb"`
	BillingEventsTotalETB  float64     `json:"billing_events_total_etb"`
	BillingEventsUnbilled  int64       `json:"billing_events_unbilled"`
	SubscriptionsByPlan    []PlanCount `json:"subscriptions_by_plan"`
}

// AdvertiserSummary is per-advertiser operational stats.
type AdvertiserSummary struct {
	AdvertiserID    string  `json:"advertiser_id"`
	CompanyName     string  `json:"company_name"`
	PlanName        string  `json:"plan_name"`
	SubscriptionStatus string `json:"subscription_status"`
	CampaignCount   int64   `json:"campaign_count"`
	ActiveCampaigns int64   `json:"active_campaigns"`
	TotalDeliveries int64   `json:"total_deliveries"`
	SegmentSpendETB float64 `json:"segment_spend_etb"`
}

// AnalyticsRepository reads aggregated metrics (read-only).
type AnalyticsRepository interface {
	Overview(ctx context.Context) (*OverviewStats, error)
	SubscriptionsByPlan(ctx context.Context) ([]PlanCount, error)
	CampaignPerformance(ctx context.Context) ([]CampaignPerformance, error)
	CampaignDetail(ctx context.Context, campaignID string) (*CampaignDetail, error)
	DeliveryAnalytics(ctx context.Context, since time.Time) (*DeliveryAnalytics, error)
	RevenueOverview(ctx context.Context, since30d time.Time) (*RevenueOverview, error)
	AdvertiserSummaries(ctx context.Context) ([]AdvertiserSummary, error)
}
