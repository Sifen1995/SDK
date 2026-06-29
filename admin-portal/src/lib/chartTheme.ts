/** Shared Recharts styling — uses CSS variables for light/dark readability. */

export const CHART_PALETTE = ['#1589FF', '#10b981', '#f59e0b', '#ef4444', '#06b6d4', '#8b5cf6'];

export const CHART_ACCENT = 'var(--chart-accent)';
export const CHART_GRID = 'var(--chart-grid)';

export const chartTooltipProps = {
  cursor: { fill: 'var(--bg-subtle)', opacity: 0.6 },
  contentStyle: {
    backgroundColor: 'var(--surface-elevated)',
    border: '1px solid var(--border-strong)',
    borderRadius: '10px',
    boxShadow: '0 8px 24px rgb(0 0 0 / 0.18)',
    color: 'var(--text-primary)',
  },
  labelStyle: { color: 'var(--text-primary)', fontWeight: 600, marginBottom: 4 },
  itemStyle: { color: 'var(--text-primary)' },
};

export const chartLegendProps = {
  verticalAlign: 'bottom' as const,
  height: 36,
  wrapperStyle: { paddingTop: '16px', color: 'var(--text-muted)', fontSize: 12 },
};

export const axisTick = { fill: 'var(--text-muted)', fontSize: 12 };
export const axisStroke = 'var(--text-muted)';
export const categoryAxisStroke = 'var(--text-primary)';
