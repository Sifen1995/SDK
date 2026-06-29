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

export interface AudienceSegment {
  id: string;
  name: string;
  description: string;
  price_etb: number;
  top_intent_signals?: string[];
  approximate_size?: number;
  estimated_cpm?: number;
  is_active?: boolean;
}

export interface SegmentCandidate {
  id: string;
  intent_name: string;
  user_count: number;
  avg_confidence: number;
  avg_days_active: number;
  min_days_active: number;
  lookback_days: number;
  status: 'pending' | 'approved' | 'rejected';
  scanned_at: string;
}

export interface BillingRate {
  id: string;
  plan_id: string;
  billing_model: string;
  rate_etb: number;
  is_active: boolean;
}

export interface Campaign {
  id: string;
  advertiserId: string;
  name: string;
  targetIntent: string;
  creativeFormat: string;
  title: string;
  bodyText: string;
  imageUrl: string;
  destinationUrl: string;
  canvasJson: Record<string, unknown>;
  dailyBudgetCap: number;
  totalBudgetCap: number;
  isActive: boolean;
  validationStatus: string;
  validationNotes: string;
  moderationStatus: string;
  moderationNotes: string;
  channelId: string;
  channelCode: string;
  billingModel: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateCampaignRequest {
  name: string;
  target_intent: string;
  creative_format: string;
  title?: string;
  body_text?: string;
  image_url?: string;
  destination_url: string;
  canvas_json?: Record<string, unknown>;
  daily_budget_cap?: number;
  total_budget_cap?: number;
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

export const TARGET_INTENTS = [
  { value: 'crypto_interest', label: 'Crypto interest' },
  { value: 'fashion_interest', label: 'Fashion interest' },
  { value: 'food_interest', label: 'Food interest' },
  { value: 'education_interest', label: 'Education interest' },
  { value: 'gaming_interest', label: 'Gaming interest' },
  { value: 'fintech_interest', label: 'Fintech interest' },
  { value: 'general_interest', label: 'General interest' },
] as const;

export const CREATIVE_FORMATS = [
  { value: 'BANNER', label: 'In-App Banner', description: 'Image banner with optional title overlay' },
  { value: 'PUSH_PLUS', label: 'Push+ Notification', description: 'Rich push with title (1–50) and body (1–120)' },
  { value: 'SMS_PLUS', label: 'SMS+ Canvas', description: 'Interactive canvas with title (1–40) and body (1–160)' },
] as const;

export const ROLE_META: Record<PortalRole, { label: string; description: string; canWrite: boolean; selfRegister: boolean }> = {
  advertiser: {
    label: 'Advertiser',
    description: 'Create, activate, and manage campaigns for your company.',
    canWrite: true,
    selfRegister: true,
  },
  read_only_analyst: {
    label: 'Read-only Analyst',
    description: 'View campaigns and previews. Cannot create or activate campaigns.',
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

export interface CreatePlanRequest {
  name: string;
  monthly_fee_etb: number;
  max_active_campaigns: number;
  max_daily_budget_etb: number;
  included_impressions: number;
  sms_plus_enabled: boolean;
  audiencemart_enabled: boolean;
  cpc_discount_pct: number;
}

export interface CreateSegmentRequest {
  name: string;
  description: string;
  top_intent_signals: string[];
  approximate_size: number;
  estimated_cpm: number;
  is_active: boolean;
}

export interface UpdateBillingRateRequest {
  rate_etb: number;
  is_active: boolean;
}

export interface ApproveSegmentCandidateRequest {
  name: string;
  description: string;
  estimated_cpm: number;
}

export interface RejectSegmentCandidateRequest {
  notes: string;
}
