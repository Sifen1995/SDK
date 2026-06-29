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
  billingModel: string;
  dailyBudgetCap: number;
  totalBudgetCap: number;
  budgetSpent: number;
  frequencyCapPerDay: number;
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
  billing_model: string;
  daily_budget_cap: number;
  total_budget_cap: number;
  frequency_cap_per_day: number;
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

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
  company_name: string;
  role?: 'advertiser' | 'read_only_analyst';
}

export interface CreateUserRequest {
  name: string;
  email: string;
  password: string;
  company_name?: string;
  role: PortalRole;
}

export const BILLING_MODELS = [
  { value: 'CPC', label: 'CPC', description: 'Cost per click' },
  { value: 'CPM', label: 'CPM', description: 'Cost per 1,000 impressions' },
  { value: 'CPI', label: 'CPI', description: 'Cost per install' },
  { value: 'CPA', label: 'CPA', description: 'Cost per action' },
  { value: 'REV_SHARE', label: 'Rev share', description: 'Revenue share model' },
] as const;

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
