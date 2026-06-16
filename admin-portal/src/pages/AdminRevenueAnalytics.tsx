import { useEffect, useState } from 'react';
import { api } from '../lib/api';
import type { RevenueOverview } from '../types/analytics';
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts';
import {
  axisStroke,
  axisTick,
  categoryAxisStroke,
  CHART_GRID,
  CHART_PALETTE,
  chartLegendProps,
  chartTooltipProps,
} from '../lib/chartTheme';
import { fmtEtb, fmtNum } from '../lib/format';

export default function AdminRevenueAnalytics() {
  const [data, setData] = useState<RevenueOverview | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api.analyticsRevenue()
      .then(setData)
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="text-muted">Loading revenue analytics...</div>;
  if (error) return <div className="alert-error">{error}</div>;
  if (!data) return <div className="text-muted">No revenue data available.</div>;

  const plans = data.subscriptions_by_plan ?? [];
  const revenueBars = [
    { name: 'MRR', amount: data.estimated_mrr_etb },
    { name: 'Segments (30d)', amount: data.segment_revenue_30d_etb },
  ];

  return (
    <div>
      <h1 className="text-2xl font-bold text-primary mb-2">Revenue Analytics</h1>
      <p className="text-muted mb-8">Financial metrics and subscription breakdowns.</p>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {[
          { label: 'Estimated MRR', value: fmtEtb(data.estimated_mrr_etb) },
          { label: 'Segment Revenue (All Time)', value: fmtEtb(data.segment_revenue_total_etb) },
          { label: 'Segment Revenue (30d)', value: fmtEtb(data.segment_revenue_30d_etb) },
          { label: 'Total Billed Events', value: fmtEtb(data.billing_events_total_etb) },
        ].map((stat, i) => (
          <div key={i} className="card-static p-6 border-t-4 border-t-green-500">
            <p className="text-sm font-medium text-muted mb-2">{stat.label}</p>
            <h3 className="text-2xl font-bold text-primary">{stat.value}</h3>
          </div>
        ))}
      </div>

      <div className="grid lg:grid-cols-2 gap-6">
        <div className="card-static p-6">
          <h3 className="font-semibold text-primary mb-6">Subscriptions by Plan</h3>
          <div className="h-72">
            {plans.length === 0 ? (
              <div className="h-full flex items-center justify-center text-muted">No subscriptions found</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={plans} dataKey="count" nameKey="plan_name" cx="50%" cy="50%" innerRadius={60} outerRadius={100} paddingAngle={5}>
                    {plans.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={CHART_PALETTE[index % CHART_PALETTE.length]} />
                    ))}
                  </Pie>
                  <Tooltip {...chartTooltipProps} />
                  <Legend {...chartLegendProps} />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>

        <div className="card-static p-6">
          <h3 className="font-semibold text-primary mb-4">Financial Health</h3>
          <div className="space-y-4">
            <div className="p-4 bg-[var(--bg-subtle)] rounded-xl border border-[var(--border)] flex justify-between items-center">
              <div>
                <p className="font-medium text-primary">Unbilled Events</p>
                <p className="text-xs text-muted mt-1">Pending delivery logs not yet aggregated into billing</p>
              </div>
              <p className="font-bold text-amber-600 dark:text-amber-400 text-xl">{fmtNum(data.billing_events_unbilled)}</p>
            </div>

            <div className="p-4 bg-[var(--bg-subtle)] rounded-xl border border-[var(--border)]">
              <p className="font-medium text-primary mb-4">Revenue Streams</p>
              <div className="h-40">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={revenueBars} layout="vertical" margin={{ top: 0, right: 30, left: 40, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke={CHART_GRID} />
                    <XAxis
                      type="number"
                      stroke={axisStroke}
                      tick={axisTick}
                      tickFormatter={val => `ETB ${val >= 1000 ? `${(val / 1000).toFixed(1)}k` : val}`}
                    />
                    <YAxis dataKey="name" type="category" stroke={categoryAxisStroke} tick={{ ...axisTick, fontWeight: 500 }} />
                    <Tooltip
                      {...chartTooltipProps}
                      formatter={value => [fmtEtb(Number(value ?? 0)), 'Revenue']}
                    />
                    <Bar dataKey="amount" radius={[0, 4, 4, 0]} barSize={24}>
                      {revenueBars.map((_, index) => (
                        <Cell key={`cell-${index}`} fill={CHART_PALETTE[index % CHART_PALETTE.length]} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
