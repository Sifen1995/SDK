package infrastructure

import (
	"context"
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

func (r *Repository) Overview(ctx context.Context) (*analyticsdomain.OverviewStats, error) {
	var s analyticsdomain.OverviewStats
	db := r.db.WithContext(ctx)
	db.Raw(`SELECT COUNT(*) FROM advertisers`).Scan(&s.TotalAdvertisers)
	db.Raw(`SELECT COUNT(*) FROM campaigns`).Scan(&s.TotalCampaigns)
	db.Raw(`SELECT COUNT(*) FROM campaigns WHERE is_active = true`).Scan(&s.ActiveCampaigns)
	db.Raw(`SELECT COUNT(*) FROM campaigns WHERE moderation_status = 'pending' AND is_active = false`).Scan(&s.PendingModeration)
	db.Raw(`SELECT COUNT(*) FROM delivery_jobs`).Scan(&s.TotalDeliveries)
	db.Raw(`SELECT COUNT(*) FROM delivery_jobs WHERE created_at >= NOW() - INTERVAL '24 hours'`).Scan(&s.DeliveriesLast24h)
	db.Raw(`SELECT COUNT(*) FROM delivery_jobs WHERE created_at >= NOW() - INTERVAL '7 days'`).Scan(&s.DeliveriesLast7d)
	db.Raw(`SELECT COUNT(*) FROM advertiser_subscriptions WHERE status = 'active'`).Scan(&s.ActiveSubscriptions)
	db.Raw(`
		SELECT COALESCE(SUM(sp.monthly_fee_etb), 0)
		FROM advertiser_subscriptions s
		JOIN subscription_plans sp ON sp.id = s.plan_id
		WHERE s.status = 'active'
	`).Scan(&s.EstimatedMRR_ETB)
	db.Raw(`SELECT COALESCE(SUM(amount_paid), 0) FROM segment_purchases`).Scan(&s.SegmentRevenueTotalETB)
	db.Raw(`SELECT COUNT(DISTINCT user_id) FROM delivery_jobs`).Scan(&s.UniqueUsersReached)
	return &s, nil
}

func (r *Repository) SubscriptionsByPlan(ctx context.Context) ([]analyticsdomain.PlanCount, error) {
	var rows []analyticsdomain.PlanCount
	err := r.db.WithContext(ctx).Raw(`
		SELECT sp.name AS plan_name, COUNT(*) AS count
		FROM advertiser_subscriptions s
		JOIN subscription_plans sp ON sp.id = s.plan_id
		WHERE s.status = 'active'
		GROUP BY sp.name
		ORDER BY count DESC
	`).Scan(&rows).Error
	return rows, err
}

func (r *Repository) CampaignPerformance(ctx context.Context) ([]analyticsdomain.CampaignPerformance, error) {
	var rows []analyticsdomain.CampaignPerformance
	err := r.db.WithContext(ctx).Raw(`
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
			COUNT(DISTINCT dj.user_id) AS unique_users,
			c.budget_spent,
			c.daily_budget_cap
		FROM campaigns c
		LEFT JOIN advertisers a ON a.id = c.advertiser_id
		LEFT JOIN delivery_jobs dj ON dj.campaign_id = c.id
		GROUP BY c.id, a.company_name
		ORDER BY delivery_count DESC, c.created_at DESC
	`).Scan(&rows).Error
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
	_ = r.db.WithContext(ctx).Raw(`
		SELECT delivery_status AS status, COUNT(*) AS count
		FROM campaign_delivery_logs
		WHERE campaign_id = ?
		GROUP BY delivery_status
		ORDER BY count DESC
	`, campaignID).Scan(&funnel).Error
	return &analyticsdomain.CampaignDetail{CampaignPerformance: *base, Funnel: funnel}, nil
}

func (r *Repository) DeliveryAnalytics(ctx context.Context, since time.Time) (*analyticsdomain.DeliveryAnalytics, error) {
	out := &analyticsdomain.DeliveryAnalytics{}
	r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM delivery_jobs`).Scan(&out.TotalDeliveries)
	_ = r.db.WithContext(ctx).Raw(`
		SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD') AS day, COUNT(*) AS count
		FROM delivery_jobs
		WHERE created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY day ASC
	`, since).Scan(&out.Last30Days)
	top, _ := r.CampaignPerformance(ctx)
	if len(top) > 10 {
		top = top[:10]
	}
	out.TopCampaigns = top
	_ = r.db.WithContext(ctx).Raw(`
		SELECT delivery_status AS status, COUNT(*) AS count
		FROM campaign_delivery_logs
		GROUP BY delivery_status
		ORDER BY count DESC
	`).Scan(&out.FunnelPlatform)
	return out, nil
}

func (r *Repository) RevenueOverview(ctx context.Context, since30d time.Time) (*analyticsdomain.RevenueOverview, error) {
	out := &analyticsdomain.RevenueOverview{}
	r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(sp.monthly_fee_etb), 0)
		FROM advertiser_subscriptions s
		JOIN subscription_plans sp ON sp.id = s.plan_id
		WHERE s.status = 'active'
	`).Scan(&out.EstimatedMRR_ETB)
	r.db.WithContext(ctx).Raw(`SELECT COALESCE(SUM(amount_paid), 0) FROM segment_purchases`).Scan(&out.SegmentRevenueTotalETB)
	r.db.WithContext(ctx).Raw(`SELECT COALESCE(SUM(amount_paid), 0) FROM segment_purchases WHERE created_at >= ?`, since30d).Scan(&out.SegmentRevenue30dETB)
	r.db.WithContext(ctx).Raw(`SELECT COALESCE(SUM(charge_etb), 0) FROM billing_events`).Scan(&out.BillingEventsTotalETB)
	r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM billing_events WHERE is_billed = false`).Scan(&out.BillingEventsUnbilled)
	plans, _ := r.SubscriptionsByPlan(ctx)
	out.SubscriptionsByPlan = plans
	return out, nil
}

func (r *Repository) AdvertiserSummaries(ctx context.Context) ([]analyticsdomain.AdvertiserSummary, error) {
	var rows []analyticsdomain.AdvertiserSummary
	err := r.db.WithContext(ctx).Raw(`
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
	`).Scan(&rows).Error
	return rows, err
}
