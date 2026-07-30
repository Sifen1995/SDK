package infrastructure

import (
	"context"
	"fmt"
	"time"

	analyticsdomain "skykin-platform/internal/analytics/domain"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var _ analyticsdomain.AnalyticsRepository = (*Repository)(nil)

// scan runs one aggregate query and labels any failure with the metric it belongs to,
// so an operator sees which table is unavailable instead of a zero.
func (r *Repository) scan(ctx context.Context, metric, query string, dest any, args ...any) error {
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(dest).Error; err != nil {
		return fmt.Errorf("analytics: %s: %w", metric, err)
	}
	return nil
}

func (r *Repository) Overview(ctx context.Context) (*analyticsdomain.OverviewStats, error) {
	var s analyticsdomain.OverviewStats
	queries := []struct {
		metric string
		query  string
		dest   any
	}{
		{"total_advertisers", `SELECT COUNT(*) FROM advertisers`, &s.TotalAdvertisers},
		{"total_campaigns", `SELECT COUNT(*) FROM campaigns`, &s.TotalCampaigns},
		{"active_campaigns", `SELECT COUNT(*) FROM campaigns WHERE is_active = true`, &s.ActiveCampaigns},
		{"pending_moderation", `SELECT COUNT(*) FROM campaigns WHERE moderation_status = 'pending' AND is_active = false`, &s.PendingModeration},
		{"total_deliveries", `SELECT COUNT(*) FROM delivery_jobs`, &s.TotalDeliveries},
		{"deliveries_last_24h", `SELECT COUNT(*) FROM delivery_jobs WHERE created_at >= NOW() - INTERVAL '24 hours'`, &s.DeliveriesLast24h},
		{"deliveries_last_7d", `SELECT COUNT(*) FROM delivery_jobs WHERE created_at >= NOW() - INTERVAL '7 days'`, &s.DeliveriesLast7d},
		{"active_subscriptions", `SELECT COUNT(*) FROM advertiser_subscriptions WHERE status = 'active'`, &s.ActiveSubscriptions},
		{"estimated_mrr", `
			SELECT COALESCE(SUM(sp.monthly_fee_etb), 0)
			FROM advertiser_subscriptions s
			JOIN subscription_plans sp ON sp.id = s.plan_id
			WHERE s.status = 'active'
		`, &s.EstimatedMRR_ETB},
		{"segment_revenue_total", `SELECT COALESCE(SUM(amount_paid), 0) FROM segment_purchases`, &s.SegmentRevenueTotalETB},
		{"unique_users_reached", `SELECT COUNT(DISTINCT pseudonymous_id) FROM delivery_jobs`, &s.UniqueUsersReached},
	}
	for _, q := range queries {
		if err := r.scan(ctx, q.metric, q.query, q.dest); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func (r *Repository) SubscriptionsByPlan(ctx context.Context) ([]analyticsdomain.PlanCount, error) {
	var rows []analyticsdomain.PlanCount
	err := r.scan(ctx, "subscriptions_by_plan", `
		SELECT sp.name AS plan_name, COUNT(*) AS count
		FROM advertiser_subscriptions s
		JOIN subscription_plans sp ON sp.id = s.plan_id
		WHERE s.status = 'active'
		GROUP BY sp.name
		ORDER BY count DESC
	`, &rows)
	return rows, err
}

func (r *Repository) CampaignPerformance(ctx context.Context) ([]analyticsdomain.CampaignPerformance, error) {
	var rows []analyticsdomain.CampaignPerformance
	err := r.scan(ctx, "campaign_performance", `
		SELECT
			c.id AS campaign_id,
			c.name,
			c.advertiser_id,
			COALESCE(a.company_name, '') AS company_name,
			c.target_intent,
			c.is_active,
			c.moderation_status,
			c.validation_status,
			COUNT(dj.id) AS delivery_count,
			COUNT(DISTINCT dj.pseudonymous_id) AS unique_users,
			c.budget_spent,
			c.daily_budget_cap
		FROM campaigns c
		LEFT JOIN advertisers a ON a.id = c.advertiser_id
		LEFT JOIN delivery_jobs dj ON dj.campaign_id = c.id
		GROUP BY c.id, a.company_name
		ORDER BY delivery_count DESC, c.created_at DESC
	`, &rows)
	return rows, err
}

func (r *Repository) CampaignDetail(ctx context.Context, campaignID string) (*analyticsdomain.CampaignDetail, error) {
	list, err := r.CampaignPerformance(ctx)
	if err != nil {
		return nil, err
	}
	var base *analyticsdomain.CampaignPerformance
	for i := range list {
		if list[i].CampaignID == campaignID {
			base = &list[i]
			break
		}
	}
	if base == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var funnel []analyticsdomain.FunnelStep
	if err := r.scan(ctx, "campaign_funnel", `
		SELECT delivery_status AS status, COUNT(*) AS count
		FROM campaign_delivery_logs
		WHERE campaign_id = ?
		GROUP BY delivery_status
		ORDER BY count DESC
	`, &funnel, campaignID); err != nil {
		return nil, err
	}
	return &analyticsdomain.CampaignDetail{CampaignPerformance: *base, Funnel: funnel}, nil
}

func (r *Repository) DeliveryAnalytics(ctx context.Context, since time.Time) (*analyticsdomain.DeliveryAnalytics, error) {
	out := &analyticsdomain.DeliveryAnalytics{}
	if err := r.scan(ctx, "total_deliveries", `SELECT COUNT(*) FROM delivery_jobs`, &out.TotalDeliveries); err != nil {
		return nil, err
	}
	if err := r.scan(ctx, "deliveries_by_day", `
		SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD') AS day, COUNT(*) AS count
		FROM delivery_jobs
		WHERE created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY day ASC
	`, &out.Last30Days, since); err != nil {
		return nil, err
	}
	top, err := r.CampaignPerformance(ctx)
	if err != nil {
		return nil, err
	}
	if len(top) > 10 {
		top = top[:10]
	}
	out.TopCampaigns = top
	if err := r.scan(ctx, "platform_funnel", `
		SELECT delivery_status AS status, COUNT(*) AS count
		FROM campaign_delivery_logs
		GROUP BY delivery_status
		ORDER BY count DESC
	`, &out.FunnelPlatform); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) RevenueOverview(ctx context.Context, since30d time.Time) (*analyticsdomain.RevenueOverview, error) {
	out := &analyticsdomain.RevenueOverview{}
	if err := r.scan(ctx, "estimated_mrr", `
		SELECT COALESCE(SUM(sp.monthly_fee_etb), 0)
		FROM advertiser_subscriptions s
		JOIN subscription_plans sp ON sp.id = s.plan_id
		WHERE s.status = 'active'
	`, &out.EstimatedMRR_ETB); err != nil {
		return nil, err
	}
	if err := r.scan(ctx, "segment_revenue_total",
		`SELECT COALESCE(SUM(amount_paid), 0) FROM segment_purchases`,
		&out.SegmentRevenueTotalETB); err != nil {
		return nil, err
	}
	if err := r.scan(ctx, "segment_revenue_30d",
		`SELECT COALESCE(SUM(amount_paid), 0) FROM segment_purchases WHERE created_at >= ?`,
		&out.SegmentRevenue30dETB, since30d); err != nil {
		return nil, err
	}
	if err := r.scan(ctx, "billing_events_total",
		`SELECT COALESCE(SUM(charge_etb), 0) FROM billing_events`,
		&out.BillingEventsTotalETB); err != nil {
		return nil, err
	}
	if err := r.scan(ctx, "billing_events_unbilled",
		`SELECT COUNT(*) FROM billing_events WHERE is_billed = false`,
		&out.BillingEventsUnbilled); err != nil {
		return nil, err
	}
	plans, err := r.SubscriptionsByPlan(ctx)
	if err != nil {
		return nil, err
	}
	out.SubscriptionsByPlan = plans
	return out, nil
}

func (r *Repository) AdvertiserSummaries(ctx context.Context) ([]analyticsdomain.AdvertiserSummary, error) {
	var rows []analyticsdomain.AdvertiserSummary
	err := r.scan(ctx, "advertiser_summaries", `
		SELECT
			a.id AS advertiser_id,
			a.company_name,
			COALESCE(sp.name, 'none') AS plan_name,
			COALESCE(s.status, 'none') AS subscription_status,
			(SELECT COUNT(*) FROM campaigns c WHERE c.advertiser_id = a.id) AS campaign_count,
			(SELECT COUNT(*) FROM campaigns c WHERE c.advertiser_id = a.id AND c.is_active = true) AS active_campaigns,
			(SELECT COUNT(*) FROM delivery_jobs dj
			 JOIN campaigns c ON c.id = dj.campaign_id WHERE c.advertiser_id = a.id) AS total_deliveries,
			(SELECT COALESCE(SUM(amount_paid), 0) FROM segment_purchases WHERE advertiser_id = a.id) AS segment_spend_etb
		FROM advertisers a
		LEFT JOIN advertiser_subscriptions s ON s.advertiser_id = a.id AND s.status = 'active'
		LEFT JOIN subscription_plans sp ON sp.id = s.plan_id
		ORDER BY total_deliveries DESC, a.company_name ASC
	`, &rows)
	return rows, err
}
