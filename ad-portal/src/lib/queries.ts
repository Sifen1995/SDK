import { useMutation, useQuery, useQueryClient } from '@skykin/ui';
import { api } from './api';
import type { CreateCampaignRequest, CreateUserRequest, CreateZoneRequest } from '../types';

/** Query-key factory — granular invalidation, no blanket clears. */
export const qk = {
  campaigns: {
    all: ['campaigns'] as const,
    list: (offset: number, limit: number) => [...qk.campaigns.all, 'list', offset, limit] as const,
    detail: (id: string) => [...qk.campaigns.all, 'detail', id] as const,
    preview: (id: string) => [...qk.campaigns.all, 'preview', id] as const,
  },
  catalog: {
    plans: ['plans'] as const,
    channels: ['channels'] as const,
    segments: ['segments'] as const,
  },
  team: { all: ['team'] as const },
  zones: {
    all: ['zones'] as const,
    forCampaign: (id: string) => ['zones', 'campaign', id] as const,
  },
} as const;

export const useCampaigns = (offset: number, limit: number) =>
  useQuery({ queryKey: qk.campaigns.list(offset, limit), queryFn: () => api.listCampaigns(offset, limit), placeholderData: prev => prev });
export const useCampaign = (id: string) =>
  useQuery({ queryKey: qk.campaigns.detail(id), queryFn: () => api.getCampaign(id), enabled: !!id });
export const useCampaignPreview = (id: string) =>
  useQuery({ queryKey: qk.campaigns.preview(id), queryFn: () => api.previewCampaign(id), enabled: !!id, staleTime: 5 * 60_000 });

export const usePlans = () => useQuery({ queryKey: qk.catalog.plans, queryFn: api.listPlans });
export const useChannels = () => useQuery({ queryKey: qk.catalog.channels, queryFn: api.listChannels });
export const useSegments = () => useQuery({ queryKey: qk.catalog.segments, queryFn: api.listSegments });

export function useCreateCampaign() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateCampaignRequest) => api.createCampaign(data),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.campaigns.all }),
  });
}

export const useZones = () => useQuery({ queryKey: qk.zones.all, queryFn: api.listZones });
export const useCampaignZones = (id: string) =>
  useQuery({
    queryKey: qk.zones.forCampaign(id),
    queryFn: () => api.listCampaignZones(id),
    enabled: !!id,
  });

export function useCreateZone() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateZoneRequest) => api.createZone(data),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.zones.all }),
  });
}

export function useLinkCampaignZones() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ campaignId, zoneIds }: { campaignId: string; zoneIds: string[] }) =>
      api.linkCampaignZones(campaignId, zoneIds),
    onSuccess: (_, { campaignId }) =>
      client.invalidateQueries({ queryKey: qk.zones.forCampaign(campaignId) }),
  });
}

export function useSubscribe() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (planId: string) => api.subscribe(planId),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.catalog.plans }),
  });
}

export function useCreateTeamUser() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateUserRequest) => api.createUser(data),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.team.all }),
  });
}
