import * as React from 'react';
import { Loader2, AlertTriangle, Inbox } from 'lucide-react';
import { cn } from '../lib/utils';
import { Button } from './ui/button';
import { Skeleton } from './ui/skeleton';

export function Spinner({ className }: { className?: string }) {
  return <Loader2 className={cn('size-4 animate-spin text-muted-foreground', className)} aria-hidden />;
}

/** Full-block loading indicator. */
export function LoadingState({ label = 'Loading…', className }: { label?: string; className?: string }) {
  return (
    <div className={cn('flex items-center justify-center gap-2 py-16 text-sm text-muted-foreground', className)} role="status">
      <Spinner />
      {label}
    </div>
  );
}

/** Designed empty state — an invitation to act, not a blank screen. */
export function EmptyState({
  icon: Icon = Inbox,
  title,
  description,
  action,
  className,
}: {
  icon?: React.ComponentType<{ className?: string }>;
  title: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn('flex flex-col items-center justify-center rounded-lg border border-dashed border-border bg-card/50 px-6 py-16 text-center', className)}>
      <div className="mb-4 flex size-12 items-center justify-center rounded-full bg-identity/10 text-identity">
        <Icon className="size-6" />
      </div>
      <h3 className="font-display text-base font-semibold text-foreground">{title}</h3>
      {description && <p className="mt-1.5 max-w-sm text-sm text-muted-foreground">{description}</p>}
      {action && <div className="mt-5">{action}</div>}
    </div>
  );
}

/** Error state that says plainly what happened and how to fix it. */
export function ErrorState({
  title = 'Something went wrong',
  message,
  onRetry,
  className,
}: {
  title?: string;
  message?: string;
  onRetry?: () => void;
  className?: string;
}) {
  return (
    <div className={cn('flex flex-col items-center justify-center rounded-lg border border-destructive/30 bg-destructive-surface px-6 py-14 text-center', className)} role="alert">
      <div className="mb-4 flex size-12 items-center justify-center rounded-full bg-destructive/12 text-destructive">
        <AlertTriangle className="size-6" />
      </div>
      <h3 className="font-display text-base font-semibold text-foreground">{title}</h3>
      {message && <p className="mt-1.5 max-w-md text-sm text-muted-foreground">{message}</p>}
      {onRetry && (
        <Button variant="outline" size="sm" className="mt-5" onClick={onRetry}>
          Try again
        </Button>
      )}
    </div>
  );
}

/** Inline alert for form/action errors. */
export function InlineError({ message, className }: { message: string; className?: string }) {
  return (
    <div className={cn('flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive-surface px-3 py-2 text-sm text-destructive', className)} role="alert">
      <AlertTriangle className="mt-0.5 size-4 shrink-0" />
      <span>{message}</span>
    </div>
  );
}

/** Skeleton table body while a table query is pending. */
export function TableSkeleton({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <div className="space-y-2 py-2" aria-hidden>
      {Array.from({ length: rows }).map((_, r) => (
        <div key={r} className="flex gap-3">
          {Array.from({ length: cols }).map((_, c) => (
            <Skeleton key={c} className={cn('h-6 flex-1', c === 0 && 'max-w-[40%]')} />
          ))}
        </div>
      ))}
    </div>
  );
}
