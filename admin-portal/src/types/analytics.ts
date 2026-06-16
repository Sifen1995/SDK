export interface OverviewStats {
  total_advertisers: number;
  total_campaigns: number;
  active_campaigns: number;
  pending_moderation: number;
  total_deliveries: number;
  deliveries_last_24h: number;
  deliveries_last_7d: number;
  active_subscriptions: number;
  estimated_mrr_etb: number;
  segment_revenue_total_etb: number;
  unique_users_reached: number;
}

export interface PlanCount {
  plan_name: string;
  count: number;
}

export interface CampaignPerformance {
  campaign_id: string;
  name: string;
  advertiser_id: string;
  company_name: string;
  target_intent: string;
  is_active: boolean;
  moderation_status: string;
  validation_status: string;
  delivery_count: number;
  unique_users: number;
  budget_spent: number;
  daily_budget_cap: number;
}

export interface FunnelStep {
  status: string;
  count: number;
}

export interface CampaignDetail extends CampaignPerformance {
  funnel: FunnelStep[];
}

export interface DeliveryTrendPoint {
  day: string;
  count: number;
}

export interface DeliveryAnalytics {
  total_deliveries: number;
  last_30_days: DeliveryTrendPoint[];
  top_campaigns: CampaignPerformance[];
  funnel_platform: FunnelStep[];
}

export interface RevenueOverview {
  estimated_mrr_etb: number;
  segment_revenue_total_etb: number;
  segment_revenue_30d_etb: number;
  billing_events_total_etb: number;
  billing_events_unbilled: number;
  subscriptions_by_plan: PlanCount[];
}

export interface AdvertiserSummary {
  advertiser_id: string;
  company_name: string;
  plan_name: string;
  subscription_status: string;
  campaign_count: number;
  active_campaigns: number;
  total_deliveries: number;
  segment_spend_etb: number;
}
