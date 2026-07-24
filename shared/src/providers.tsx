import * as React from 'react';
import { QueryClientProvider } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import { NuqsAdapter } from 'nuqs/adapters/react-router/v7';
import { TooltipProvider } from './components/ui/tooltip';
import { createQueryClient } from './query';

/**
 * Wraps an app with TanStack Query, the nuqs URL-state adapter (React Router
 * v7), and the tooltip provider. Must render INSIDE a Router.
 */
export function AppProviders({
  children,
  client,
}: {
  children: React.ReactNode;
  client?: QueryClient;
}) {
  const [qc] = React.useState(() => client ?? createQueryClient());
  return (
    <QueryClientProvider client={qc}>
      <NuqsAdapter>
        <TooltipProvider delayDuration={200}>{children}</TooltipProvider>
      </NuqsAdapter>
    </QueryClientProvider>
  );
}
