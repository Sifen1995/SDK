import { useMemo } from 'react';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell,
} from 'recharts';
import { Megaphone, Clock, DollarSign, Users, Send, Activity } from 'lucide-react';
import {
  Card, CardHeader, CardTitle, CardContent, KpiCard, LoadingState, ErrorState,
  chartAxis, chartGrid, chartTooltip, chartColor, CHART_COLORS,
} from '@skykin/ui';
import { fmtEtb, fmtNum } from '../lib/format';
import { useOverview } from '../lib/queries';

export default function AdminDashboard() {
  const { data: stats, isPending, isError, error, refetch } = useOverview();

  const statusData = useMemo(() => {
    if (!stats) return [];
    return [
      { name: 'Active', value: stats.active_campaigns },
      { name: 'Pending', value: stats.pending_moderation },
      { name: 'Inactive', value: Math.max(0, stats.total_campaigns - stats.active_campaigns - stats.pending_moderation) },
    ].filter(d => d.value > 0);
  }, [stats]);

  if (isPending) return <LoadingState label="Loading platform overview…" />;
  if (isError) return <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />;

  const kpis = [
    { label: 'Total Advertisers', value: fmtNum(stats.total_advertisers), icon: Users },
    { label: 'Active Campaigns', value: fmtNum(stats.active_campaigns), icon: Megaphone, sub: `of ${fmtNum(stats.total_campaigns)} total` },
    { label: 'Deliveries · 24h', value: fmtNum(stats.deliveries_last_24h), icon: Send },
    { label: 'Deliveries · 7d', value: fmtNum(stats.deliveries_last_7d), icon: Activity },
    { label: 'Est. MRR (ETB)', value: fmtEtb(stats.estimated_mrr_etb), icon: DollarSign },
    { label: 'Pending Moderation', value: fmtNum(stats.pending_moderation), icon: Clock },
  ];

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {kpis.map(k => (
          <KpiCard key={k.label} label={k.label} value={k.value} icon={k.icon} sub={k.sub} />
        ))}
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex-row items-center justify-between">
            <CardTitle>Delivery volumes</CardTitle>
            <span className="rounded-md bg-muted px-2 py-0.5 text-[11px] font-medium tabular-nums text-muted-foreground">
              Total {fmtNum(stats.total_deliveries)}
            </span>
          </CardHeader>
          <CardContent>
            <div className="h-56">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart
                  data={[
                    { name: 'Last 24h', count: stats.deliveries_last_24h },
                    { name: 'Last 7d', count: stats.deliveries_last_7d },
                  ]}
                  layout="vertical"
                  margin={{ top: 0, right: 24, left: 24, bottom: 0 }}
                >
                  <CartesianGrid {...chartGrid} horizontal={false} vertical />
                  <XAxis type="number" {...chartAxis} tickFormatter={v => (v >= 1000 ? `${(v / 1000).toFixed(1)}k` : String(v))} />
                  <YAxis dataKey="name" type="category" {...chartAxis} width={64} />
                  <Tooltip {...chartTooltip} />
                  <Bar dataKey="count" name="Deliveries" fill={chartColor(0)} radius={[0, 6, 6, 0]} barSize={30} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Campaign status</CardTitle></CardHeader>
          <CardContent>
            {statusData.length === 0 ? (
              <div className="flex h-56 items-center justify-center text-sm text-muted-foreground">No campaigns</div>
            ) : (
              <div className="flex flex-col items-center gap-5 sm:flex-row">
                <div className="relative h-44 w-44 shrink-0">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie data={statusData} dataKey="value" nameKey="name" cx="50%" cy="50%" innerRadius={58} outerRadius={78} paddingAngle={3} stroke="none">
                        {statusData.map((_, i) => <Cell key={i} fill={chartColor(i)} />)}
                      </Pie>
                      <Tooltip {...chartTooltip} />
                    </PieChart>
                  </ResponsiveContainer>
                  <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
                    <span className="font-display text-xl font-bold tabular-nums">{fmtNum(stats.total_campaigns)}</span>
                    <span className="text-[10px] uppercase tracking-wider text-muted-foreground">campaigns</span>
                  </div>
                </div>
                <ul className="w-full flex-1 space-y-2">
                  {statusData.map((d, i) => (
                    <li key={d.name} className="flex items-center justify-between text-sm">
                      <span className="flex items-center gap-2 text-muted-foreground">
                        <span className="size-2.5 rounded-full" style={{ background: CHART_COLORS[i % CHART_COLORS.length] }} />
                        {d.name}
                      </span>
                      <span className="font-semibold tabular-nums">{fmtNum(d.value)}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <KpiCard label="Unique users reached" value={fmtNum(stats.unique_users_reached)} icon={Users} />
        <KpiCard label="Active subscriptions" value={fmtNum(stats.active_subscriptions)} icon={Activity} />
        <KpiCard label="Segment sales revenue" value={fmtEtb(stats.segment_revenue_total_etb)} icon={DollarSign} className="border-success/30 [&_.text-foreground]:text-success [&_span[data-slot=kpi-value]]:text-success" />
      </div>
    </div>
  );
}
