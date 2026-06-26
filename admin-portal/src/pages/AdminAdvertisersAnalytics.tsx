import { useEffect, useMemo, useState } from 'react';
import { api } from '../lib/api';
import type { AdvertiserSummary } from '../types/analytics';
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
import { fmtEtb, fmtNum } from '../lib/format';
import FilterBar, { FilterSearch, FilterSelect } from '../components/FilterBar';
import OffsetPagination, { paginateSlice, PAGE_SIZE } from '../components/Pagination';

export default function AdminAdvertisersAnalytics() {
  const [data, setData] = useState<AdvertiserSummary[]>([]);
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [tablePage, setTablePage] = useState(1);

  useEffect(() => {
    api.analyticsAdvertisers()
      .then(res => {
        setData(res.advertisers ?? []);
        setCount(res.count ?? res.advertisers?.length ?? 0);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return data.filter(adv => {
      if (statusFilter !== 'all' && (adv.subscription_status || 'none') !== statusFilter) return false;
      if (!q) return true;
      return adv.company_name.toLowerCase().includes(q) || adv.advertiser_id.toLowerCase().includes(q);
    });
  }, [data, search, statusFilter]);

  useEffect(() => {
    setTablePage(1);
  }, [search, statusFilter]);

  const tableRows = useMemo(
    () => paginateSlice(filtered, tablePage, PAGE_SIZE),
    [filtered, tablePage],
  );

  const statusOptions = useMemo(() => {
    const statuses = [...new Set(data.map(a => a.subscription_status || 'none'))];
    return [{ value: 'all', label: 'All subscriptions' }, ...statuses.map(s => ({ value: s, label: s }))];
  }, [data]);

  if (loading) return <div className="text-muted">Loading advertisers...</div>;
  if (error) return <div className="alert-error">{error}</div>;

  const statusDist = filtered.reduce(
    (acc, curr) => {
      const status = curr.subscription_status || 'none';
      acc[status] = (acc[status] || 0) + 1;
      return acc;
    },
    {} as Record<string, number>,
  );
  const pieData = Object.entries(statusDist).map(([name, value]) => ({ name, value }));
  const topSpenders = [...filtered].sort((a, b) => (b.segment_spend_etb ?? 0) - (a.segment_spend_etb ?? 0)).slice(0, 5);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-bold text-primary mb-1">Advertisers</h1>
        <p className="text-muted">Operational summary across {count} registered advertisers.</p>
      </div>

      <FilterBar resultCount={filtered.length} totalCount={data.length}>
        <FilterSearch value={search} onChange={setSearch} placeholder="Search company or ID…" />
        <FilterSelect value={statusFilter} onChange={setStatusFilter} options={statusOptions} />
      </FilterBar>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="card-static p-6">
          <h3 className="font-semibold text-primary mb-6">Subscription Status</h3>
          <div className="h-64">
            {pieData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-muted">No data available</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" innerRadius={60} outerRadius={80} paddingAngle={5}>
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

        <div className="card-static p-6">
          <h3 className="font-semibold text-primary mb-6">Top Spenders</h3>
          <div className="h-64">
            {topSpenders.length === 0 ? (
              <div className="h-full flex items-center justify-center text-muted">No spenders</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={topSpenders} layout="vertical" margin={{ top: 0, right: 30, left: 40, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke={CHART_GRID} />
                  <XAxis
                    type="number"
                    stroke={axisStroke}
                    tick={axisTick}
                    tickFormatter={val => `ETB ${val >= 1000 ? `${(val / 1000).toFixed(1)}k` : val}`}
                  />
                  <YAxis dataKey="company_name" type="category" stroke={categoryAxisStroke} tick={{ ...axisTick, fontWeight: 500 }} width={100} />
                  <Tooltip
                    {...chartTooltipProps}
                    formatter={value => [fmtEtb(Number(value ?? 0)), 'Spend']}
                  />
                  <Bar dataKey="segment_spend_etb" fill={CHART_ACCENT} radius={[0, 4, 4, 0]} barSize={24} />
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
                <th className="p-4 font-medium">Company</th>
                <th className="p-4 font-medium">Plan</th>
                <th className="p-4 font-medium">Status</th>
                <th className="p-4 font-medium text-right">Campaigns</th>
                <th className="p-4 font-medium text-right">Active</th>
                <th className="p-4 font-medium text-right">Total Deliveries</th>
                <th className="p-4 font-medium text-right">Segment Spend</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--border)]">
              {tableRows.map(adv => (
                <tr key={adv.advertiser_id} className="hover:bg-[var(--bg-subtle)] transition">
                  <td className="p-4">
                    <p className="font-medium text-primary">{adv.company_name}</p>
                    <p className="text-xs font-mono text-muted mt-1">{adv.advertiser_id}</p>
                  </td>
                  <td className="p-4">
                    <span className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 border border-blue-200 dark:border-blue-800">
                      {adv.plan_name || 'N/A'}
                    </span>
                  </td>
                  <td className="p-4">
                    <span
                      className={`inline-flex items-center px-2 py-1 rounded-md text-xs font-medium border ${
                        adv.subscription_status === 'active'
                          ? 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/30 dark:text-green-400 dark:border-green-800'
                          : 'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700'
                      }`}
                    >
                      {adv.subscription_status || 'none'}
                    </span>
                  </td>
                  <td className="p-4 text-right text-primary font-medium">{adv.campaign_count}</td>
                  <td className="p-4 text-right text-primary font-medium">{adv.active_campaigns}</td>
                  <td className="p-4 text-right text-primary font-medium">{fmtNum(adv.total_deliveries)}</td>
                  <td className="p-4 text-right text-green-600 dark:text-green-400 font-medium">{fmtEtb(adv.segment_spend_etb)}</td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={7} className="p-8 text-center text-muted">
                    No advertisers match your filters.
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
