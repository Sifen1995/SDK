import { useMemo } from 'react';
import { useQueryState, parseAsString } from 'nuqs';
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar,
} from 'recharts';
import { Search } from 'lucide-react';
import {
  Card, CardHeader, CardTitle, CardContent, KpiCard, Input, LoadingState, ErrorState,
  chartAxis, chartGrid, chartTooltip, chartColor,
} from '@skykin/ui';
import { fmtNum } from '../lib/format';
import { useDelivery } from '../lib/queries';

export default function AdminDeliveryAnalytics() {
  const { data, isPending, isError, error, refetch } = useDelivery();
  const [q, setQ] = useQueryState('q', parseAsString.withDefault(''));

  const topCampaigns = data?.top_campaigns ?? [];
  const filteredTop = useMemo(() => {
    const term = q.trim().toLowerCase();
    if (!term) return topCampaigns;
    return topCampaigns.filter(
      c =>
        c.name.toLowerCase().includes(term) ||
        c.company_name.toLowerCase().includes(term) ||
        c.target_intent.toLowerCase().includes(term),
    );
  }, [topCampaigns, q]);

  if (isPending) return <LoadingState label="Loading delivery analytics…" />;
  if (isError) return <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />;

  const last30 = data.last_30_days ?? [];
  const funnel = data.funnel_platform ?? [];

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <KpiCard label="Total deliveries" value={fmtNum(data.total_deliveries)} />
        <KpiCard label="30-day data points" value={fmtNum(last30.length)} />
        <KpiCard label="Funnel stages" value={fmtNum(funnel.length)} />
      </div>

      <Card>
        <CardHeader><CardTitle>30-day dispatch volume</CardTitle></CardHeader>
        <CardContent>
          <div className="h-72">
            {last30.length === 0 ? (
              <div className="flex h-full items-center justify-center text-sm text-muted-foreground">No delivery data in the last 30 days.</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={last30} margin={{ top: 10, right: 24, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="deliveryGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="var(--chart-1)" stopOpacity={0.35} />
                      <stop offset="95%" stopColor="var(--chart-1)" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid {...chartGrid} />
                  <XAxis dataKey="day" {...chartAxis} tickMargin={10} minTickGap={30} />
                  <YAxis {...chartAxis} tickFormatter={v => (v >= 1000 ? `${(v / 1000).toFixed(1)}k` : String(v))} />
                  <Tooltip {...chartTooltip} />
                  <Area type="monotone" dataKey="count" name="Deliveries" stroke="var(--chart-1)" strokeWidth={2.5} fill="url(#deliveryGradient)" />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle>Platform funnel</CardTitle></CardHeader>
          <CardContent>
            <div className="h-64">
              {funnel.length === 0 ? (
                <div className="flex h-full items-center justify-center text-sm text-muted-foreground">No funnel data yet.</div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={funnel} layout="vertical" margin={{ top: 0, right: 24, left: 24, bottom: 0 }}>
                    <CartesianGrid {...chartGrid} horizontal={false} vertical />
                    <XAxis type="number" {...chartAxis} />
                    <YAxis dataKey="status" type="category" {...chartAxis} width={96} />
                    <Tooltip {...chartTooltip} />
                    <Bar dataKey="count" name="Count" fill={chartColor(0)} radius={[0, 6, 6, 0]} barSize={30} />
                  </BarChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-center justify-between gap-3">
            <CardTitle>Top campaigns by delivery</CardTitle>
            <div className="relative w-44">
              <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input value={q} onChange={e => setQ(e.target.value || null)} placeholder="Filter…" className="h-8 pl-8 text-xs" />
            </div>
          </CardHeader>
          <CardContent className="space-y-2.5">
            {filteredTop.length === 0 ? (
              <p className="py-6 text-center text-sm text-muted-foreground">No campaigns match your filter.</p>
            ) : (
              filteredTop.slice(0, 8).map(c => (
                <div key={c.campaign_id} className="flex items-center justify-between rounded-lg border border-border bg-muted/40 p-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{c.name}</p>
                    <p className="truncate text-xs text-muted-foreground">{c.company_name} · {c.target_intent}</p>
                  </div>
                  <div className="shrink-0 text-right">
                    <p className="font-display font-bold tabular-nums">{fmtNum(c.delivery_count)}</p>
                    <p className="text-xs text-muted-foreground">deliveries</p>
                  </div>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
