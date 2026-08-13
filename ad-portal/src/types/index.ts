export type PortalRole = 'operator_admin' | 'advertiser' | 'read_only_analyst';

export interface PortalUser {
  id: string;
  email: string;
  name: string;
  role: PortalRole;
  role_id: string;
  advertiser_id: string;
  company_name?: string;
  is_active: boolean;
}

export interface Plan {
  id: string;
  name: string;
  monthly_fee_etb: number;
  max_active_campaigns: number;
  max_daily_budget_etb: number;
  included_impressions: number;
  sms_plus_enabled: boolean;
  audiencemart_enabled: boolean;
  cpc_discount_pct: number;
}

export interface Subscription {
  id: string;
  plan: Plan;
  status: string;
  current_period_start: string;
  current_period_end: string;
  impressions_used: number;
}

export interface SubscriptionStatus {
  subscribed: boolean;
  subscription: Subscription | null;
}

export interface DeliveryChannel {
  id: string;
  code: string;
  name: string;
  description: string;
  is_premium: boolean;
}

export interface AudienceSegment {
  id: string;
  name: string;
  description: string;
  top_intent_signals: string[];
  approximate_size: number;
  estimated_cpm: number;
  estimated_price_etb: number;
  purchasable: boolean;
}

export interface SegmentsCatalog {
  plan_name: string;
  audiencemart_enabled: boolean;
  segments: AudienceSegment[];
}

export interface Campaign {
  id: string;
  advertiserId: string;
  name: string;
  targetIntent: string;
  channelId: string;
  channelCode: string;
  segmentId: string | null;
  title: string;
  bodyText: string;
  imageUrl: string;
  destinationUrl: string;
  canvasJson: Record<string, unknown>;
  dailyBudgetCap: number;
  totalBudgetCap: number;
  budgetSpent: number;
  frequencyCapPerDay: number;
  scheduledStartAt: string | null;
  scheduledEndAt: string | null;
  isActive: boolean;
  validationStatus: string;
  validationNotes: string;
  moderationStatus: string;
  moderationNotes: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateCampaignRequest {
  name: string;
  target_intent: string;
  channel_id: string;
  segment_id?: string | null;
  title?: string;
  body_text: string;
  image_url?: string;
  destination_url: string;
  canvas_json?: Record<string, unknown>;
  daily_budget_cap: number;
  total_budget_cap: number;
  frequency_cap_per_day: number;
  /** RFC3339. The backend enforces end > start. */
  scheduled_start_at?: string;
  scheduled_end_at?: string;
}

export interface CampaignPreview {
  format: string;
  campaign_name: string;
  simulator: boolean;
  channel_label: string;
  preview: {
    title?: string;
    body_text?: string;
    image_url?: string;
    destination_url?: string;
    creative_format?: string;
    canvas_json?: Record<string, unknown>;
  };
}

/**
 * A store geofence zone. `POST /ad-portal/geofences` always returns
 * `is_active: false` — an operator activates it, or approving a campaign it is
 * linked to does so automatically.
 */
export interface GeofenceZone {
  id: string;
  advertiser_id?: string;
  latitude: number;
  longitude: number;
  radius_metres: number;
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface CreateZoneRequest {
  latitude: number;
  longitude: number;
  radius_metres: number;
}

/**
 * POST /ad-portal/register. Exactly these three fields — the handler hardcodes
 * the advertiser role and sets `advertisers.company_name` to the user's name,
 * so `role` and `company_name` were both accepted-then-discarded by gin's
 * non-strict binding. Operators create analyst/admin users via CreateUserRequest.
 */
export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

/** The 201 body — a flat object, not wrapped in `{ user }`. */
export interface RegisterResponse {
  id: string;
  email: string;
  name: string;
  role: PortalRole;
  created_at: string;
}

export interface CreateUserRequest {
  name: string;
  email: string;
  password: string;
  company_name?: string;
  role: PortalRole;
}

// BILLING_MODELS used to drive a campaign-level select. The backend dropped
// `campaigns.billing_model` (migration 20260728120000) — billing is standardized
// to CPC and the create DTO does not accept the field. Per-event billing models
// still exist on `billing_events`, but they are derived server-side from the
// event type and configured by operators under Plans & Billing, not per campaign.

export const TARGET_INTENTS = [
  { value: 'crypto_interest', label: 'Crypto interest' },
  { value: 'fashion_interest', label: 'Fashion interest' },
  { value: 'food_interest', label: 'Food interest' },
  { value: 'education_interest', label: 'Education interest' },
  { value: 'gaming_interest', label: 'Gaming interest' },
  { value: 'fintech_interest', label: 'Fintech interest' },
  { value: 'general_interest', label: 'General interest' },
] as const;

export const ROLE_META: Record<PortalRole, { label: string; description: string; canWrite: boolean; selfRegister: boolean }> = {
  advertiser: {
    label: 'Advertiser',
    description: 'Create and manage campaigns for your company.',
    canWrite: true,
    selfRegister: true,
  },
  read_only_analyst: {
    label: 'Read-only Analyst',
    description: 'View campaigns and previews. Cannot create campaigns.',
    canWrite: false,
    selfRegister: true,
  },
  operator_admin: {
    label: 'Operator Admin',
    description: 'Full platform access including team user management.',
    canWrite: true,
    selfRegister: false,
  },
};

export function canWriteCampaigns(role: PortalRole): boolean {
  return ROLE_META[role].canWrite;
}

export function isOperatorAdmin(role: PortalRole): boolean {
  return role === 'operator_admin';
}

export function channelNeedsImage(code: string): boolean {
  return code === 'IN_APP_BANNER' || code === 'NATIVE_FEED';
}

export function channelNeedsRichCopy(code: string): boolean {
  return code === 'PUSH' || code === 'SMS_PLUS';
}
