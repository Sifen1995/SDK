import * as React from 'react';
import type { LucideIcon } from 'lucide-react';
import { ArrowUpRight, ArrowDownRight, Minus } from 'lucide-react';
import { cn } from '../lib/utils';
import { Card } from './ui/card';

export type Trend = { dir: 'up' | 'down' | 'flat'; label: string };

export function KpiCard({
  label,
  value,
  icon: Icon,
  sub,
  trend,
  className,
}: {
  label: string;
  value: React.ReactNode;
  icon?: LucideIcon;
  sub?: string;
  trend?: Trend;
  className?: string;
}) {
  const TrendIcon = trend?.dir === 'up' ? ArrowUpRight : trend?.dir === 'down' ? ArrowDownRight : Minus;
  const trendClass =
    trend?.dir === 'up' ? 'text-success' : trend?.dir === 'down' ? 'text-destructive' : 'text-muted-foreground';

  return (
    <Card
      className={cn(
        'lift group relative overflow-hidden p-5',
        // tight edge shadow + soft brand-tinted diffuse glow + inset top highlight, all in one value
        'shadow-[0_2px_4px_rgba(8,38,62,0.08),0_16px_36px_-12px_rgba(8,38,62,0.22),inset_0_1px_0_0_rgba(255,255,255,0.7)]',
        'border border-primary/40',
        className,
      )}
    >
      {/* signature accent: soft blurred brand blob, clipped corner */}
      <span
        aria-hidden
        className="pointer-events-none absolute -right-8 -top-10 size-32 rounded-full bg-primary/10 blur-2xl"
      />

      {/* brand accent hairline — hidden at rest, appears on hover */}
      <span
        aria-hidden
        className="pointer-events-none absolute inset-x-0 top-0 h-0.5 bg-gradient-to-r from-primary to-transparent opacity-0 transition-opacity duration-150 group-hover:opacity-90"
      />

      <div className="relative flex items-center justify-between">
        <p className="text-xs font-medium text-muted-foreground">{label}</p>
        {Icon && (
          <span className="flex size-8 items-center justify-center rounded-md bg-primary/10 text-primary ring-1 ring-primary/15 shadow-sm">
            <Icon className="size-4" />
          </span>
        )}
      </div>

      <div className="relative mt-4">
        <div className="flex items-baseline justify-between gap-2">
          <span data-slot="kpi-value" className="font-display text-2xl font-bold leading-none tracking-tight tabular-nums text-foreground">
            {value}
          </span>
          {trend && (
            <span className={cn('flex items-center gap-0.5 text-xs font-semibold tabular-nums px-1.5 py-0.5 rounded-sm bg-muted/50', trendClass)}>
              <TrendIcon className="size-3.5" />
              {trend.label}
            </span>
          )}
        </div>
      </div>
      {sub && (
         <div className="relative mt-4 border-t border-border/40 pt-3">
            <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">{sub}</p>
         </div>
      )}
    </Card>
  );
}