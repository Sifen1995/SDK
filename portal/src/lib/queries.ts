import { useMutation, useQuery, useQueryClient } from '@skykin/ui';
import { api } from './api';

export const qk = {
  applications: ['applications'] as const,
};

export const useApplications = () =>
  useQuery({ queryKey: qk.applications, queryFn: () => api.getApplications().then(r => r.data.applications ?? []) });

export function useCreateApplication() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (v: { appName: string; platform: string; bundleId: string }) =>
      api.createApplication(v.appName, v.platform, v.bundleId).then(r => r.data),
    onSuccess: () => client.invalidateQueries({ queryKey: qk.applications }),
  });
}
