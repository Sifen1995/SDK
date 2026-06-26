import type {
  Campaign,
  CampaignPreview,
  CreateCampaignRequest,
  CreateUserRequest,
  DeliveryChannel,
  Plan,
  PortalUser,
  RegisterRequest,
  SegmentsCatalog,
  SubscriptionStatus,
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
  const token = localStorage.getItem('adPortalToken');
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

  listPlans() {
    return request<{ plans: Plan[]; count: number }>('/plans').then(res => res.plans ?? []);
  },

  listChannels() {
    return request<{ channels: DeliveryChannel[]; count: number }>('/channels').then(res => res.channels ?? []);
  },

  getSubscription() {
    return request<SubscriptionStatus>('/subscription');
  },

  subscribe(planId: string) {
    return request<SubscriptionStatus>('/subscription', {
      method: 'POST',
      body: JSON.stringify({ plan_id: planId }),
    });
  },

  listSegments() {
    return request<SegmentsCatalog>('/audience/segments');
  },

  listCampaigns(offset: number = 0, limit: number = 10) {
    const query = new URLSearchParams();
    if (offset > 0) query.set('offset', offset.toString());
    if (limit > 0) query.set('limit', limit.toString());
    const qs = query.toString() ? `?${query.toString()}` : '';

    return request<{ campaigns: unknown[]; total?: number }>(`/campaigns${qs}`).then(res => ({
      campaigns: campaignsFromList(res),
      total: res.total,
    }));
  },

  async getCampaign(id: string): Promise<Campaign> {
    const raw = await request<Record<string, unknown>>(`/campaigns/${id}`);
    return normalizeCampaign(raw);
  },

  async createCampaign(data: CreateCampaignRequest): Promise<Campaign> {
    const payload = { ...data };
    if (!payload.segment_id) {
      delete payload.segment_id;
    }
    const raw = await request<Record<string, unknown>>('/campaigns', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return normalizeCampaign(raw);
  },

  previewCampaign(id: string) {
    return request<CampaignPreview>(`/campaigns/${id}/preview`);
  },

  async activateCampaign(id: string): Promise<Campaign> {
    const raw = await request<Record<string, unknown>>(`/campaigns/${id}/activate`, {
      method: 'POST',
    });
    return normalizeCampaign(raw);
  },
};
