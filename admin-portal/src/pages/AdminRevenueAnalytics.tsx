import {
  PieChart, Pie, Cell, Tooltip, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid,
} from 'recharts';
import {
  Card, CardHeader, CardTitle, CardContent, KpiCard, LoadingState, ErrorState,
  chartAxis, chartGrid, chartTooltip, chartColor, CHART_COLORS,
} from '@skykin/ui';
import { fmtEtb, fmtNum } from '../lib/format';
import { useRevenue } from '../lib/queries';

export default function AdminRevenueAnalytics() {
  const { data, isPending, isError, error, refetch } = useRevenue();

  if (isPending) return <LoadingState label="Loading revenue analytics…" />;
  if (isError) return <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />;

  const plans = (data.subscriptions_by_plan ?? []).map(p => ({ name: p.plan_name, value: p.count }));
  const revenueBars = [
    { name: 'MRR', amount: data.estimated_mrr_etb },
    { name: 'Segments (30d)', amount: data.segment_revenue_30d_etb },
  ];

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <KpiCard label="Estimated MRR" value={fmtEtb(data.estimated_mrr_etb)} />
        <KpiCard label="Segment revenue (all-time)" value={fmtEtb(data.segment_revenue_total_etb)} />
        <KpiCard label="Segment revenue (30d)" value={fmtEtb(data.segment_revenue_30d_etb)} />
        <KpiCard label="Total billed events" value={fmtEtb(data.billing_events_total_etb)} />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle>Subscriptions by plan</CardTitle></CardHeader>
          <CardContent>
            {plans.length === 0 ? (
              <div className="flex h-64 items-center justify-center text-sm text-muted-foreground">No subscriptions found</div>
            ) : (
              <div className="flex flex-col items-center gap-6 sm:flex-row">
                <div className="h-56 w-56 shrink-0">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie data={plans} dataKey="value" nameKey="name" cx="50%" cy="50%" innerRadius={60} outerRadius={90} paddingAngle={3} stroke="none">
                        {plans.map((_, i) => <Cell key={i} fill={chartColor(i)} />)}
                      </Pie>
                      <Tooltip {...chartTooltip} />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <ul className="w-full flex-1 space-y-2">
                  {plans.map((p, i) => (
                    <li key={p.name} className="flex items-center justify-between text-sm">
                      <span className="flex items-center gap-2 text-muted-foreground">
                        <span className="size-2.5 rounded-full" style={{ background: CHART_COLORS[i % CHART_COLORS.length] }} />
                        {p.name}
                      </span>
                      <span className="font-semibold tabular-nums">{fmtNum(p.value)}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Financial health</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between rounded-lg border border-border bg-muted/50 p-4">
              <div>
                <p className="text-sm font-medium">Unbilled events</p>
                <p className="mt-0.5 text-xs text-muted-foreground">Delivery logs not yet aggregated into billing</p>
              </div>
              <p className="font-display text-xl font-bold tabular-nums text-warning">{fmtNum(data.billing_events_unbilled)}</p>
            </div>
            <div className="rounded-lg border border-border bg-muted/50 p-4">
              <p className="mb-3 text-sm font-medium">Revenue streams</p>
              <div className="h-40">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={revenueBars} layout="vertical" margin={{ top: 0, right: 24, left: 24, bottom: 0 }}>
                    <CartesianGrid {...chartGrid} horizontal={false} vertical />
                    <XAxis type="number" {...chartAxis} tickFormatter={v => `ETB ${v >= 1000 ? `${(v / 1000).toFixed(1)}k` : v}`} />
                    <YAxis dataKey="name" type="category" {...chartAxis} width={90} />
                    <Tooltip {...chartTooltip} formatter={(v) => [fmtEtb(Number(v ?? 0)), 'Revenue']} />
                    <Bar dataKey="amount" radius={[0, 6, 6, 0]} barSize={24}>
                      {revenueBars.map((_, i) => <Cell key={i} fill={chartColor(i)} />)}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
