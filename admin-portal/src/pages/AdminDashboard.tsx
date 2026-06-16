import { useEffect, useMemo, useState } from 'react';
import { api } from '../lib/api';
import type { OverviewStats } from '../types/analytics';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts';
import {
  axisStroke,
  axisTick,
  categoryAxisStroke,
  CHART_ACCENT,
  CHART_GRID,
  CHART_PALETTE,
  chartLegendProps,
  chartTooltipProps,
} from '../lib/chartTheme';
import { fmtEtb, fmtNum } from '../lib/format';

export default function AdminDashboard() {
  const [stats, setStats] = useState<OverviewStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api.analyticsOverview()
      .then(setStats)
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const pieData = useMemo(() => {
    if (!stats) return [];
    return [
      { name: 'Active', value: stats.active_campaigns },
      { name: 'Pending', value: stats.pending_moderation },
      { name: 'Inactive', value: Math.max(0, stats.total_campaigns - stats.active_campaigns - stats.pending_moderation) },
    ].filter(d => d.value > 0);
  }, [stats]);

  if (loading) return <div className="text-muted">Loading overview...</div>;
  if (error) return <div className="alert-error">{error}</div>;
  if (!stats) return <div className="text-muted">No overview data available.</div>;

  const statCards = [
    { label: 'Active Campaigns', value: stats.active_campaigns, total: stats.total_campaigns, icon: '📢', color: 'text-brand-500 dark:text-brand-300', bg: 'bg-brand-50 dark:bg-brand-900/30' },
    { label: 'Pending Moderation', value: stats.pending_moderation, icon: '⚡', color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-900/30' },
    { label: 'Est. MRR (ETB)', value: fmtEtb(stats.estimated_mrr_etb), icon: '💰', color: 'text-green-600 dark:text-green-400', bg: 'bg-green-50 dark:bg-green-900/30' },
    { label: 'Total Advertisers', value: stats.total_advertisers, icon: '👥', color: 'text-blue-600 dark:text-blue-400', bg: 'bg-blue-50 dark:bg-blue-900/30' },
  ];

  return (
    <div>
      <h1 className="text-2xl font-bold text-primary mb-2">Platform Overview</h1>
      <p className="text-muted mb-8">High-level metrics and system status.</p>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {statCards.map((stat, i) => (
          <div key={i} className="card-static p-6">
            <div className="flex items-center justify-between mb-4">
              <div className={`p-3 rounded-xl ${stat.bg}`}>
                <span className={`text-xl ${stat.color}`}>{stat.icon}</span>
              </div>
            </div>
            <div>
              <p className="text-sm font-medium text-muted mb-1">{stat.label}</p>
              <div className="flex items-baseline gap-2">
                <h3 className="text-3xl font-bold text-primary">{stat.value}</h3>
                {stat.total !== undefined && (
                  <span className="text-xs font-medium text-muted">/ {stat.total} total</span>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card-static p-6">
          <div className="flex justify-between items-center mb-6">
            <h3 className="font-semibold text-primary">Delivery Volumes</h3>
            <span className="text-xs font-medium text-muted bg-[var(--bg-subtle)] px-2 py-1 rounded border border-[var(--border)]">
              Total: {fmtNum(stats.total_deliveries)}
            </span>
          </div>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={[
                  { name: 'Last 24h', count: stats.deliveries_last_24h },
                  { name: 'Last 7d', count: stats.deliveries_last_7d },
                ]}
                layout="vertical"
                margin={{ top: 0, right: 30, left: 40, bottom: 0 }}
              >
                <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke={CHART_GRID} />
                <XAxis
                  type="number"
                  stroke={axisStroke}
                  tick={axisTick}
                  tickFormatter={val => (val >= 1000 ? `${(val / 1000).toFixed(1)}k` : String(val))}
                />
                <YAxis dataKey="name" type="category" stroke={categoryAxisStroke} tick={{ ...axisTick, fontWeight: 500 }} />
                <Tooltip {...chartTooltipProps} />
                <Bar dataKey="count" fill={CHART_ACCENT} radius={[0, 4, 4, 0]} barSize={32} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="card-static p-6">
          <h3 className="font-semibold text-primary mb-6">Campaign Status Distribution</h3>
          <div className="h-64">
            {stats.total_campaigns === 0 ? (
              <div className="h-full flex items-center justify-center text-muted">No campaigns</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={pieData}
                    dataKey="value"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={5}
                  >
                    {pieData.map((_, index) => (
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
      </div>

      <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="card-static p-4 flex flex-col justify-center items-center text-center">
          <p className="text-sm text-muted mb-1">Unique Users Reached</p>
          <p className="text-2xl font-bold text-primary">{fmtNum(stats.unique_users_reached)}</p>
        </div>
        <div className="card-static p-4 flex flex-col justify-center items-center text-center">
          <p className="text-sm text-muted mb-1">Active Subscriptions</p>
          <p className="text-2xl font-bold text-primary">{stats.active_subscriptions}</p>
        </div>
        <div className="card-static p-4 flex flex-col justify-center items-center text-center">
          <p className="text-sm text-muted mb-1">Segment Sales Revenue</p>
          <p className="text-2xl font-bold text-green-600 dark:text-green-400">{fmtEtb(stats.segment_revenue_total_etb)}</p>
        </div>
      </div>
    </div>
  );
}
