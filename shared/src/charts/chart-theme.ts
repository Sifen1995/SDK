/**
 * One shared Recharts theme for all three apps. Colors reference the token
 * CSS variables so charts follow light/dark and stay on the muted palette.
 */
export const CHART_COLORS = [
  'var(--chart-1)',
  'var(--chart-2)',
  'var(--chart-3)',
  'var(--chart-4)',
  'var(--chart-5)',
  'var(--chart-6)',
] as const;

export const chartColor = (i: number) => CHART_COLORS[i % CHART_COLORS.length];

export const chartAxis = {
  stroke: 'var(--border)',
  tick: { fill: 'var(--muted-foreground)', fontSize: 12 },
  tickLine: false,
  axisLine: false,
} as const;

export const chartGrid = {
  stroke: 'var(--border)',
  strokeDasharray: '3 3',
  vertical: false,
} as const;

export const chartTooltip = {
  cursor: { fill: 'var(--muted)', opacity: 0.5 },
  contentStyle: {
    background: 'var(--popover)',
    border: '1px solid var(--border)',
    borderRadius: '8px',
    boxShadow: '0 12px 32px rgba(8, 38, 62, 0.14), 0 3px 8px rgba(8, 38, 62, 0.06)',
    color: 'var(--popover-foreground)',
    fontSize: 12,
  },
  labelStyle: { color: 'var(--foreground)', fontWeight: 600, marginBottom: 4 },
  itemStyle: { color: 'var(--popover-foreground)' },
} as const;

export const chartLegend = {
  wrapperStyle: { fontSize: 12, color: 'var(--muted-foreground)', paddingTop: 12 },
} as const;
