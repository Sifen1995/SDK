import { QueryClient } from '@tanstack/react-query';

/** Shared QueryClient defaults for all portals. */
export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  });
}

export * from '@tanstack/react-query';
