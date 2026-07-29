import { useMutation, useQuery, useQueryClient } from '@skykin/ui';
import { api } from './api';
import type { ApproveSegmentCandidateRequest, CreateSegmentRequest, CreateUserRequest, CreatePlanRequest, UpdateBillingRateRequest } from '../types';

/**
 * Central query-key factory — every key derives from here so invalidation is
 * granular and typo-proof (never a blanket cache clear).
 */
export const qk = {
  campaigns: {
    all: ['campaigns'] as const,
    pending: () => [...qk.campaigns.all, 'pending'] as const,
    ready: () => [...qk.campaigns.all, 'ready'] as const,
    preview: (id: string) => [...qk.campaigns.all, 'preview', id] as const,
  },
  analytics: {
    all: ['analytics'] as const,
    overview: () => [...qk.analytics.all, 'overview'] as const,
    revenue: () => [...qk.analytics.all, 'revenue'] as const,
    delivery: () => [...qk.analytics.all, 'delivery'] as const,
    advertisers: () => [...qk.analytics.all, 'advertisers'] as const,
    campaigns: () => [...qk.analytics.all, 'campaigns'] as const,
    campaign: (id: string) => [...qk.analytics.all, 'campaigns', id] as const,
  },
  segments: {
    all: ['segments'] as const,
    list: () => [...qk.segments.all, 'list'] as const,
    candidates: (status: string) => [...qk.segments.all, 'candidates', status] as const,
  },
  plans: {
    all: ['plans'] as const,
    billingRates: (planId: string) => [...qk.plans.all, planId, 'billing-rates'] as const,
  },
  sdkUsers: {
    all: ['sdk-users'] as const,
    page: (page: number, perPage: number) => [...qk.sdkUsers.all, page, perPage] as const,
  },
  rbac: {
    permissions: ['permissions'] as const,
    roles: ['roles'] as const,
  },
} as const;

/* ------------------------------------------------------------------ queries */
export const useOverview = () => useQuery({ queryKey: qk.analytics.overview(), queryFn: api.analyticsOverview });
export const useRevenue = () => useQuery({ queryKey: qk.analytics.revenue(), queryFn: api.analyticsRevenue });
export const useDelivery = () => useQuery({ queryKey: qk.analytics.delivery(), queryFn: api.analyticsDelivery });
export const useAdvertisers = () =>
  useQuery({ queryKey: qk.analytics.advertisers(), queryFn: () => api.analyticsAdvertisers().then(r => r.advertisers ?? []) });
export const useCampaignPerformance = () =>
  useQuery({ queryKey: qk.analytics.campaigns(), queryFn: () => api.analyticsCampaigns().then(r => r.campaigns ?? []) });
export const useCampaignDetail = (id: string) =>
  useQuery({ queryKey: qk.analytics.campaign(id), queryFn: () => api.analyticsCampaignDetail(id), enabled: !!id });
export const usePendingCampaigns = () =>
  useQuery({ queryKey: qk.campaigns.pending(), queryFn: api.listPendingCampaigns });
export const useReadyCampaigns = () =>
  useQuery({
    queryKey: qk.campaigns.ready(),
    queryFn: () =>
      api.listCampaigns(0, 500).then(res =>
        res.campaigns.filter(c => c.moderationStatus === 'approved' && c.validationStatus === 'passed' && !c.isActive),
      ),
  });
export const useCampaignPreview = (id: string) =>
  useQuery({ queryKey: qk.campaigns.preview(id), queryFn: () => api.previewCampaign(id), enabled: !!id, staleTime: 5 * 60_000 });
export const useSegments = () => useQuery({ queryKey: qk.segments.list(), queryFn: api.listSegments });
export const usePlans = () => useQuery({ queryKey: qk.plans.all, queryFn: api.listPlans });
export const useSegmentCandidates = (status: string) =>
  useQuery({ queryKey: qk.segments.candidates(status), queryFn: () => api.listSegmentCandidates(status) });
export const useBillingRates = (planId: string) =>
  useQuery({ queryKey: qk.plans.billingRates(planId), queryFn: () => api.listBillingRates(planId), enabled: !!planId });
export const useSdkUsers = (page: number, perPage: number) =>
  useQuery({ queryKey: qk.sdkUsers.page(page, perPage), queryFn: () => api.listSdkUsers(page, perPage), placeholderData: prev => prev });

/* ---------------------------------------------------------------- mutations */
export function useModerateCampaign() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, action, notes }: { id: string; action: 'approve' | 'reject'; notes?: string }) =>
      api.validateCampaign(id, action, notes),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: qk.campaigns.all });
      client.invalidateQueries({ queryKey: qk.analytics.overview() });
    },
  });
}

export function useActivateCampaign() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.adminActivateCampaign(id),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: qk.campaigns.all });
      client.invalidateQueries({ queryKey: qk.analytics.overview() });
    },
  });
}

export function useApproveAndGoLive() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await api.validateCampaign(id, 'approve');
      return api.adminActivateCampaign(id);
    },
    onSuccess: () => {
      client.invalidateQueries({ queryKey: qk.campaigns.all });
      client.invalidateQueries({ queryKey: qk.analytics.overview() });
    },
  });
}

export function useApproveCandidate() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: ApproveSegmentCandidateRequest }) =>
      api.approveSegmentCandidate(id, data),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: qk.segments.all });
    },
  });
}

export function useRejectCandidate() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, notes }: { id: string; notes: string }) => api.rejectSegmentCandidate(id, notes),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.segments.all }),
  });
}

export function useRunIntentConsistency() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: () => api.runIntentConsistency(),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.segments.all }),
  });
}

export function useCreateUser() {
  return useMutation({ mutationFn: (data: CreateUserRequest) => api.createUser(data) });
}

export function useCreatePlan() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePlanRequest) => api.createPlan(data),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.plans.all }),
  });
}

export function useCreateSegment() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSegmentRequest) => api.createSegment(data),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.segments.list() }),
  });
}

export function useUpdateBillingRate(planId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateBillingRateRequest }) => api.updateBillingRate(id, data),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.plans.billingRates(planId) }),
  });
}

export function useUpdatePlan() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreatePlanRequest> }) => api.updatePlan(id, data),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.plans.all }),
  });
}

export function useSuspendPlan() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.suspendPlan(id),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.plans.all }),
  });
}

export function useSuspendSegment() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.suspendSegment(id),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.segments.list() }),
  });
}

export const usePermissions = () => useQuery({ queryKey: qk.rbac.permissions, queryFn: api.listPermissions });
export const useRoles = () => useQuery({ queryKey: qk.rbac.roles, queryFn: api.listRoles });

export function useCreateRole() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; description: string }) => api.createRole(data),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.rbac.roles }),
  });
}

export function useAssignPermission() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) => api.assignPermission(roleId, permissionId),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.rbac.roles }),
  });
}

export function useRevokePermission() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) => api.revokePermission(roleId, permissionId),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.rbac.roles }),
  });
}
