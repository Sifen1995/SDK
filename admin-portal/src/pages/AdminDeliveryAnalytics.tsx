import { useEffect, useMemo, useState } from 'react';
import { api } from '../lib/api';
import type { DeliveryAnalytics } from '../types/analytics';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts';
import {
  axisStroke,
  axisTick,
  categoryAxisStroke,
  CHART_ACCENT,
  CHART_GRID,
  chartTooltipProps,
} from '../lib/chartTheme';
import { fmtNum } from '../lib/format';
import FilterBar, { FilterSearch } from '../components/FilterBar';

function normalizeDelivery(raw: DeliveryAnalytics): DeliveryAnalytics {
  return {
    total_deliveries: raw.total_deliveries ?? 0,
    last_30_days: raw.last_30_days ?? [],
    top_campaigns: raw.top_campaigns ?? [],
    funnel_platform: raw.funnel_platform ?? [],
  };
}

export default function AdminDeliveryAnalytics() {
  const [data, setData] = useState<DeliveryAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');

  useEffect(() => {
    api.analyticsDelivery()
      .then(res => setData(normalizeDelivery(res)))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const filteredTop = useMemo(() => {
    if (!data) return [];
    const q = search.trim().toLowerCase();
    const list = data.top_campaigns;
    if (!q) return list;
    return list.filter(
      c =>
        c.name.toLowerCase().includes(q) ||
        c.company_name.toLowerCase().includes(q) ||
        c.target_intent.toLowerCase().includes(q),
    );
  }, [data, search]);

  if (loading) return <div className="text-muted">Loading delivery analytics...</div>;
  if (error) return <div className="alert-error">{error}</div>;
  if (!data) return <div className="text-muted">No delivery data available.</div>;

  return (
    <div>
      <h1 className="text-2xl font-bold text-primary mb-2">Delivery Analytics</h1>
      <p className="text-muted mb-6">Delivery volume trends and funnel conversion.</p>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <div className="card-static p-5 border-t-4 border-t-brand-500">
          <p className="text-sm text-muted mb-1">Total Deliveries</p>
          <p className="text-2xl font-bold text-primary">{fmtNum(data.total_deliveries)}</p>
        </div>
        <div className="card-static p-5 border-t-4 border-t-blue-500">
          <p className="text-sm text-muted mb-1">30-Day Data Points</p>
          <p className="text-2xl font-bold text-primary">{data.last_30_days.length}</p>
        </div>
        <div className="card-static p-5 border-t-4 border-t-green-500">
          <p className="text-sm text-muted mb-1">Funnel Stages</p>
          <p className="text-2xl font-bold text-primary">{data.funnel_platform.length}</p>
        </div>
      </div>

      <div className="card-static p-6 mb-8">
        <h3 className="font-semibold text-primary mb-6">30-Day Dispatch Volume</h3>
        <div className="h-72">
          {data.last_30_days.length === 0 ? (
            <div className="h-full flex items-center justify-center text-muted">No delivery data in the last 30 days.</div>
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={data.last_30_days} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="deliveryGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={CHART_ACCENT} stopOpacity={0.35} />
                    <stop offset="95%" stopColor={CHART_ACCENT} stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke={CHART_GRID} />
                <XAxis dataKey="day" stroke={axisStroke} tick={axisTick} tickMargin={10} minTickGap={30} />
                <YAxis
                  stroke={axisStroke}
                  tick={axisTick}
                  tickFormatter={val => (val >= 1000 ? `${(val / 1000).toFixed(1)}k` : String(val))}
                />
                <Tooltip {...chartTooltipProps} />
                <Area
                  type="monotone"
                  dataKey="count"
                  stroke={CHART_ACCENT}
                  strokeWidth={2.5}
                  fillOpacity={1}
                  fill="url(#deliveryGradient)"
                />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>

      <div className="grid lg:grid-cols-2 gap-6">
        <div className="card-static p-6">
          <h3 className="font-semibold text-primary mb-6">Platform Funnel</h3>
          <div className="h-64">
            {data.funnel_platform.length === 0 ? (
              <div className="h-full flex items-center justify-center text-muted">No funnel data yet.</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={data.funnel_platform} layout="vertical" margin={{ top: 0, right: 30, left: 40, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke={CHART_GRID} />
                  <XAxis type="number" stroke={axisStroke} tick={axisTick} />
                  <YAxis dataKey="status" type="category" stroke={categoryAxisStroke} tick={{ ...axisTick, fontWeight: 500 }} width={90} />
                  <Tooltip {...chartTooltipProps} />
                  <Bar dataKey="count" fill={CHART_ACCENT} radius={[0, 4, 4, 0]} barSize={32} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>

        <div className="card-static p-6">
          <h3 className="font-semibold text-primary mb-4">Top Campaigns by Delivery</h3>

          <FilterBar resultCount={filteredTop.length} totalCount={data.top_campaigns.length}>
            <FilterSearch value={search} onChange={setSearch} placeholder="Filter campaigns…" />
          </FilterBar>

          <div className="space-y-3">
            {filteredTop.length === 0 ? (
              <p className="text-muted text-sm">No campaigns match your filter.</p>
            ) : (
              filteredTop.slice(0, 8).map(c => (
                <div
                  key={c.campaign_id}
                  className="flex justify-between items-center p-3 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)]"
                >
                  <div>
                    <p className="font-medium text-sm text-primary">{c.name}</p>
                    <p className="text-xs text-muted">
                      {c.company_name} &bull; {c.target_intent}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-brand-500 dark:text-brand-300">{fmtNum(c.delivery_count)}</p>
                    <p className="text-xs text-muted">deliveries</p>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
