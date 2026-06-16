import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../lib/api';
import type { CampaignDetail } from '../types/analytics';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { axisStroke, axisTick, categoryAxisStroke, CHART_ACCENT, CHART_GRID, chartTooltipProps } from '../lib/chartTheme';
import { fmtMoney, fmtNum } from '../lib/format';

export default function AdminCampaignDetailAnalytics() {
  const { id } = useParams<{ id: string }>();
  const [data, setData] = useState<CampaignDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!id) return;
    api.analyticsCampaignDetail(id)
      .then(res => setData({ ...res, funnel: res.funnel ?? [] }))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <div className="text-muted">Loading campaign details...</div>;
  if (error) return <div className="alert-error">{error}</div>;
  if (!data) return <div className="text-muted">Campaign not found.</div>;

  return (
    <div>
      <Link to="/campaigns" className="text-sm text-brand-500 hover:text-brand-400 font-medium block mb-4">
        &larr; Back to Campaigns
      </Link>

      <div className="flex justify-between items-start mb-8">
        <div>
          <h1 className="text-2xl font-bold text-primary">{data.name}</h1>
          <p className="text-muted mt-1">
            {data.company_name} &bull; {data.target_intent}
          </p>
        </div>
        <span
          className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${
            data.is_active
              ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
              : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400'
          }`}
        >
          {data.is_active ? 'LIVE' : 'PAUSED'}
        </span>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <div className="card-static p-4">
          <p className="text-xs text-muted mb-1">Total Deliveries</p>
          <p className="text-xl font-bold text-primary">{fmtNum(data.delivery_count)}</p>
        </div>
        <div className="card-static p-4">
          <p className="text-xs text-muted mb-1">Unique Users</p>
          <p className="text-xl font-bold text-primary">{fmtNum(data.unique_users)}</p>
        </div>
        <div className="card-static p-4">
          <p className="text-xs text-muted mb-1">Budget Spent</p>
          <p className="text-xl font-bold text-primary">{fmtMoney(data.budget_spent)}</p>
        </div>
        <div className="card-static p-4">
          <p className="text-xs text-muted mb-1">Daily Cap</p>
          <p className="text-xl font-bold text-primary">{fmtMoney(data.daily_budget_cap)}</p>
        </div>
      </div>

      <div className="card-static p-6">
        <h3 className="font-semibold text-primary mb-6">Delivery Funnel</h3>
        <div className="h-80">
          {data.funnel.length === 0 ? (
            <div className="h-full flex items-center justify-center text-muted">No delivery events found.</div>
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={data.funnel} layout="vertical" margin={{ top: 0, right: 30, left: 40, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke={CHART_GRID} />
                <XAxis type="number" stroke={axisStroke} tick={axisTick} />
                <YAxis dataKey="status" type="category" stroke={categoryAxisStroke} tick={{ ...axisTick, fontWeight: 500 }} width={90} />
                <Tooltip {...chartTooltipProps} />
                <Bar dataKey="count" fill={CHART_ACCENT} radius={[0, 4, 4, 0]} barSize={40} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>
    </div>
  );
}
