import type {
  Campaign,
  CampaignPreview,
  CreateCampaignRequest,
  CreateUserRequest,
  PortalUser,
  RegisterRequest,
} from '../types';
import { campaignsFromList, normalizeCampaign } from './campaignUtils';

const BASE = '/api/v1/ad-portal';

interface APIErrorBody {
  status?: string;
  code?: number;
  message?: string;
  details?: unknown;
  error?: string;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('adminPortalToken');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> | undefined),
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${BASE}${path}`, { ...options, headers });
  const body = (await res.json().catch(() => ({}))) as T & APIErrorBody;

  if (!res.ok) {
    const detailText =
      typeof body.details === 'string'
        ? body.details
        : body.details
          ? JSON.stringify(body.details)
          : undefined;
    const message =
      body.message ||
      body.error ||
      detailText ||
      `Request failed (${res.status})`;
    throw new Error(detailText && body.message ? `${body.message}: ${detailText}` : message);
  }

  return body as T;
}

export const api = {
  register(data: RegisterRequest) {
    return request<{ user: PortalUser }>('/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  login(email: string, password: string) {
    return request<{ token: string; user: PortalUser }>('/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  },

  me() {
    return request<{ user: PortalUser }>('/me');
  },

  createUser(data: CreateUserRequest) {
    return request<{ user: PortalUser }>('/admin/users', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  listCampaigns(offset: number = 0, limit: number = 10) {
    const query = new URLSearchParams();
    if (offset > 0) query.set('offset', offset.toString());
    if (limit > 0) query.set('limit', limit.toString());
    const qs = query.toString() ? `?${query.toString()}` : '';

    return request<{ campaigns: unknown[], total?: number }>(`/campaigns${qs}`).then(res => ({
      campaigns: campaignsFromList(res),
      total: res.total,
    }));
  },

  async getCampaign(id: string): Promise<Campaign> {
    const raw = await request<Record<string, unknown>>(`/campaigns/${id}`);
    return normalizeCampaign(raw);
  },

  async createCampaign(data: CreateCampaignRequest): Promise<Campaign> {
    const raw = await request<Record<string, unknown>>('/campaigns', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return normalizeCampaign(raw);
  },

  async activateCampaign(id: string): Promise<Campaign> {
    const raw = await request<Record<string, unknown>>(`/campaigns/${id}/activate`, {
      method: 'POST',
    });
    return normalizeCampaign(raw);
  },

  previewCampaign(id: string) {
    return request<CampaignPreview>(`/campaigns/${id}/preview`);
  },

  listPendingCampaigns() {
    return request<{ campaigns: unknown[] }>('/admin/campaigns/pending').then(res => campaignsFromList(res));
  },

  async adminActivateCampaign(id: string): Promise<Campaign> {
    const raw = await request<Record<string, unknown>>(`/admin/campaigns/${id}/activate`, {
      method: 'POST',
    });
    return normalizeCampaign(raw);
  },

  async validateCampaign(id: string, action: 'approve' | 'reject', notes?: string): Promise<Campaign> {
    const raw = await request<Record<string, unknown>>(`/admin/campaigns/${id}/validate`, {
      method: 'POST',
      body: JSON.stringify({ action, notes }),
    });
    return normalizeCampaign(raw);
  },

  listSegments() {
    return request<{ segments: import('../types').AudienceSegment[] }>('/audience/segments').then(res => res.segments);
  },

  // Analytics Endpoints
  analyticsOverview() {
    return request<import('../types/analytics').OverviewStats>('/admin/analytics/overview');
  },
  analyticsRevenue() {
    return request<import('../types/analytics').RevenueOverview>('/admin/analytics/revenue');
  },
  analyticsDelivery() {
    return request<import('../types/analytics').DeliveryAnalytics>('/admin/analytics/delivery');
  },
  analyticsAdvertisers() {
    return request<{ advertisers: import('../types/analytics').AdvertiserSummary[], count: number }>('/admin/analytics/advertisers');
  },
  analyticsCampaigns() {
    return request<{ campaigns: import('../types/analytics').CampaignPerformance[], count: number }>('/admin/analytics/campaigns');
  },
  analyticsCampaignDetail(id: string) {
    return request<import('../types/analytics').CampaignDetail>(`/admin/analytics/campaigns/${id}`);
  },

  // Segment Candidates
  listSegmentCandidates(status: string = 'pending') {
    return request<import('../types').SegmentCandidate[]>(`/audience/segment-candidates?status=${status}`);
  },
  approveSegmentCandidate(id: string, data: import('../types').ApproveSegmentCandidateRequest) {
    return request<import('../types').AudienceSegment>(`/admin/audience/segment-candidates/${id}/approve`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },
  rejectSegmentCandidate(id: string, notes: string) {
    return request<{ message: string }>(`/admin/audience/segment-candidates/${id}/reject`, {
      method: 'POST',
      body: JSON.stringify({ notes }),
    });
  },

  // Plans & Billing
  createPlan(data: import('../types').CreatePlanRequest) {
    return request<Record<string, unknown>>('/admin/plans', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },
  listBillingRates(planId: string) {
    return request<{ rates: import('../types').BillingRate[]; count: number }>(`/admin/plans/${planId}/billing-rates`);
  },
  updateBillingRate(id: string, data: import('../types').UpdateBillingRateRequest) {
    return request<import('../types').BillingRate>(`/admin/billing-rates/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  // Admin Segments
  createSegment(data: import('../types').CreateSegmentRequest) {
    return request<import('../types').AudienceSegment>('/admin/audience/segments', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // Intent Consistency Analysis
  runIntentConsistency() {
    return request<{ message: string }>('/admin/analytics/intent-consistency/run', {
      method: 'POST',
    });
  },
};
