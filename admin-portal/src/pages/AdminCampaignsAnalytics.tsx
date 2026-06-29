import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import type { CampaignPerformance } from '../types/analytics';
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts';
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
import { fmtMoney, fmtNum } from '../lib/format';
import FilterBar, { FilterSearch, FilterSelect } from '../components/FilterBar';
import OffsetPagination, { paginateSlice, PAGE_SIZE } from '../components/Pagination';

export default function AdminCampaignsAnalytics() {
  const [data, setData] = useState<CampaignPerformance[]>([]);
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [moderationFilter, setModerationFilter] = useState('all');
  const [tablePage, setTablePage] = useState(1);

  useEffect(() => {
    api.analyticsCampaigns()
      .then(res => {
        setData(res.campaigns ?? []);
        setCount(res.count ?? res.campaigns?.length ?? 0);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return data.filter(c => {
      if (statusFilter === 'live' && !c.is_active) return false;
      if (statusFilter === 'paused' && c.is_active) return false;
      if (moderationFilter !== 'all' && c.moderation_status !== moderationFilter) return false;
      if (!q) return true;
      return (
        c.name.toLowerCase().includes(q) ||
        c.company_name.toLowerCase().includes(q) ||
        c.target_intent.toLowerCase().includes(q)
      );
    });
  }, [data, search, statusFilter, moderationFilter]);

  useEffect(() => {
    setTablePage(1);
  }, [search, statusFilter, moderationFilter]);

  const tableRows = useMemo(
    () => paginateSlice(filtered, tablePage, PAGE_SIZE),
    [filtered, tablePage],
  );

  const moderationOptions = useMemo(() => {
    const statuses = [...new Set(data.map(c => c.moderation_status).filter(Boolean))];
    return [{ value: 'all', label: 'All moderation' }, ...statuses.map(s => ({ value: s, label: s }))];
  }, [data]);

  if (loading) return <div className="text-muted">Loading campaigns...</div>;
  if (error) return <div className="alert-error">{error}</div>;

  const statusDist = filtered.reduce(
    (acc, curr) => {
      const status = curr.is_active ? 'LIVE' : 'PAUSED';
      acc[status] = (acc[status] || 0) + 1;
      return acc;
    },
    {} as Record<string, number>,
  );
  const pieData = Object.entries(statusDist).map(([name, value]) => ({ name, value }));
  const topCampaigns = [...filtered].sort((a, b) => (b.delivery_count ?? 0) - (a.delivery_count ?? 0)).slice(0, 5);

  return (
    <div>
      <h1 className="text-xl font-bold text-primary mb-1">Campaigns Performance</h1>
      <p className="text-sm text-muted mb-5">Performance breakdown for all {count} campaigns across the platform.</p>

      <FilterBar resultCount={filtered.length} totalCount={data.length}>
        <FilterSearch value={search} onChange={setSearch} placeholder="Search name, company, intent…" className="min-w-[14rem]" />
        <FilterSelect
          value={statusFilter}
          onChange={setStatusFilter}
          options={[
            { value: 'all', label: 'All statuses' },
            { value: 'live', label: 'Live only' },
            { value: 'paused', label: 'Paused only' },
          ]}
        />
        <FilterSelect value={moderationFilter} onChange={setModerationFilter} options={moderationOptions} />
      </FilterBar>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="card-static p-6">
          <h3 className="font-semibold text-primary mb-6">Campaign Status</h3>
          <div className="h-64">
            {pieData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-muted">No data</div>
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
                    {pieData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.name === 'LIVE' ? CHART_PALETTE[2] : CHART_PALETTE[1]} />
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
          <h3 className="font-semibold text-primary mb-6">Top Campaigns by Delivery</h3>
          <div className="h-64">
            {topCampaigns.length === 0 ? (
              <div className="h-full flex items-center justify-center text-muted">No campaigns</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={topCampaigns} layout="vertical" margin={{ top: 0, right: 30, left: 40, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke={CHART_GRID} />
                  <XAxis
                    type="number"
                    stroke={axisStroke}
                    tick={axisTick}
                    tickFormatter={val => (val >= 1000 ? `${(val / 1000).toFixed(1)}k` : String(val))}
                  />
                  <YAxis
                    dataKey="name"
                    type="category"
                    stroke={categoryAxisStroke}
                    tick={{ ...axisTick, fontWeight: 500 }}
                    width={100}
                  />
                  <Tooltip
                    {...chartTooltipProps}
                    formatter={value => [fmtNum(Number(value ?? 0)), 'Deliveries']}
                  />
                  <Bar dataKey="delivery_count" fill={CHART_ACCENT} radius={[0, 4, 4, 0]} barSize={24} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>
      </div>

      <div className="card-static overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="border-b border-[var(--border)] bg-[var(--bg-subtle)] text-xs uppercase tracking-wider text-muted">
                <th className="p-4 font-medium">Campaign</th>
                <th className="p-4 font-medium">Status</th>
                <th className="p-4 font-medium text-right">Deliveries</th>
                <th className="p-4 font-medium text-right">Unique Users</th>
                <th className="p-4 font-medium text-right">Budget Spent</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--border)]">
              {tableRows.map(c => (
                <tr key={c.campaign_id} className="hover:bg-[var(--bg-subtle)] transition">
                  <td className="p-4">
                    <Link to={`/campaigns/${c.campaign_id}`} className="font-medium text-brand-500 hover:text-brand-400 block">
                      {c.name}
                    </Link>
                    <p className="text-xs text-muted mt-1">
                      {c.company_name} &bull; {c.target_intent}
                    </p>
                  </td>
                  <td className="p-4">
                    <div className="flex flex-col gap-1">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium border w-max ${
                          c.is_active
                            ? 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/30 dark:text-green-400 dark:border-green-800'
                            : 'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700'
                        }`}
                      >
                        {c.is_active ? 'LIVE' : 'PAUSED'}
                      </span>
                      <span className="text-xs text-muted">
                        Val: {c.validation_status} &bull; Mod: {c.moderation_status}
                      </span>
                    </div>
                  </td>
                  <td className="p-4 text-right text-primary font-medium">{fmtNum(c.delivery_count)}</td>
                  <td className="p-4 text-right text-primary font-medium">{fmtNum(c.unique_users)}</td>
                  <td className="p-4 text-right text-primary font-medium">
                    {fmtMoney(c.budget_spent)}{' '}
                    <span className="text-muted text-xs font-normal">/ {fmtMoney(c.daily_budget_cap)}/d</span>
                  </td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={5} className="p-8 text-center text-muted">
                    No campaigns match your filters.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <OffsetPagination page={tablePage} totalItems={filtered.length} onPageChange={setTablePage} />
      </div>
    </div>
  );
}
