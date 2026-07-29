import { useParams, Link } from 'react-router-dom';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { ArrowLeft } from 'lucide-react';
import {
  Card, CardHeader, CardTitle, CardContent, KpiCard, StatusPill, LoadingState, ErrorState, Button,
  chartAxis, chartGrid, chartTooltip, chartColor,
} from '@skykin/ui';
import { fmtMoney, fmtNum } from '../lib/format';
import { useCampaignDetail } from '../lib/queries';

export default function AdminCampaignDetailAnalytics() {
  const { id = '' } = useParams<{ id: string }>();
  const { data, isPending, isError, error, refetch } = useCampaignDetail(id);

  if (isPending) return <LoadingState label="Loading campaign details…" />;
  if (isError) return <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />;
  if (!data) return <ErrorState title="Not found" message="Campaign not found." />;

  const funnel = data.funnel ?? [];

  return (
    <div className="space-y-6">
      <Button asChild variant="ghost" size="sm" className="-ml-2 w-fit">
        <Link to="/campaigns"><ArrowLeft className="size-4" /> Back to campaigns</Link>
      </Button>

      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="font-display text-xl font-bold">{data.name}</h2>
          <p className="mt-1 text-sm text-muted-foreground">{data.company_name} · {data.target_intent}</p>
        </div>
        <StatusPill status={data.is_active ? 'active' : 'inactive'} />
      </div>

      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <KpiCard label="Total deliveries" value={fmtNum(data.delivery_count)} />
        <KpiCard label="Unique users" value={fmtNum(data.unique_users)} />
        <KpiCard label="Budget spent" value={fmtMoney(data.budget_spent)} />
        <KpiCard label="Daily cap" value={fmtMoney(data.daily_budget_cap)} />
      </div>

      <Card>
        <CardHeader><CardTitle>Delivery funnel</CardTitle></CardHeader>
        <CardContent>
          <div className="h-80">
            {funnel.length === 0 ? (
              <div className="flex h-full items-center justify-center text-sm text-muted-foreground">No delivery events found.</div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={funnel} layout="vertical" margin={{ top: 0, right: 24, left: 24, bottom: 0 }}>
                  <CartesianGrid {...chartGrid} horizontal={false} vertical />
                  <XAxis type="number" {...chartAxis} />
                  <YAxis dataKey="status" type="category" {...chartAxis} width={96} />
                  <Tooltip {...chartTooltip} />
                  <Bar dataKey="count" name="Count" fill={chartColor(0)} radius={[0, 6, 6, 0]} barSize={40} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
