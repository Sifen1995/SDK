import * as React from 'react';
import type { LucideIcon } from 'lucide-react';
import { ArrowUpRight, ArrowDownRight, Minus } from 'lucide-react';
import { cn } from '../lib/utils';
import { Card } from './ui/card';

export type Trend = { dir: 'up' | 'down' | 'flat'; label: string };

/**
 * KPI tile: label, large tabular value, optional secondary text and trend
 * delta. Icon sits in a neutral chip — semantic/identity color is reserved,
 * so KPI chrome stays quiet (the number is the hero).
 */
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
    <Card className={cn('p-5', className)}>
      <div className="flex items-center justify-between">
        <p className="text-xs font-medium text-muted-foreground">{label}</p>
        {Icon && (
          <span className="flex size-8 items-center justify-center rounded-md bg-muted text-muted-foreground">
            <Icon className="size-4" />
          </span>
        )}
      </div>
      <div className="mt-3 flex items-end justify-between gap-2">
        <span data-slot="kpi-value" className="font-display text-2xl font-bold leading-none tracking-tight tabular-nums">
          {value}
        </span>
        {trend && (
          <span className={cn('flex items-center gap-0.5 text-xs font-semibold tabular-nums', trendClass)}>
            <TrendIcon className="size-3.5" />
            {trend.label}
          </span>
        )}
      </div>
      {sub && <p className="mt-1 text-[11px] text-muted-foreground">{sub}</p>}
    </Card>
  );
}
