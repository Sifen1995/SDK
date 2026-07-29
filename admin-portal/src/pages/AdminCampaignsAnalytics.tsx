import { useMemo } from 'react';
import { Link } from 'react-router-dom';
import { useQueryStates, parseAsString, parseAsInteger } from 'nuqs';
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts';
import { Download, Search } from 'lucide-react';
import {
  Card, CardHeader, CardTitle, CardContent, Button, Input,
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
  DataTable, type ColumnDef, StatusPill, LoadingState, ErrorState, exportToCsv,
  chartAxis, chartGrid, chartTooltip, chartColor,
} from '@skykin/ui';
import type { CampaignPerformance } from '../types/analytics';
import { fmtMoney, fmtNum } from '../lib/format';
import { useCampaignPerformance } from '../lib/queries';

const PAGE_SIZE = 8;

const columns: ColumnDef<CampaignPerformance>[] = [
  {
    accessorKey: 'name',
    header: 'Campaign',
    cell: ({ row }) => (
      <div className="min-w-0">
        <Link to={`/campaigns/${row.original.campaign_id}`} className="font-medium text-identity hover:underline">
          {row.original.name}
        </Link>
        <p className="mt-0.5 text-xs text-muted-foreground">{row.original.company_name} · {row.original.target_intent}</p>
      </div>
    ),
  },
  {
    id: 'status',
    header: 'Status',
    cell: ({ row }) => (
      <div className="flex flex-col gap-1">
        <StatusPill status={row.original.is_active ? 'active' : 'inactive'} />
        <span className="text-[11px] text-muted-foreground">
          Val: {row.original.validation_status} · Mod: {row.original.moderation_status}
        </span>
      </div>
    ),
  },
  {
    accessorKey: 'delivery_count',
    header: 'Deliveries',
    meta: { className: 'text-right' },
    cell: ({ getValue }) => <span className="tabular-nums">{fmtNum(getValue() as number)}</span>,
  },
  {
    accessorKey: 'unique_users',
    header: 'Unique users',
    meta: { className: 'text-right' },
    cell: ({ getValue }) => <span className="tabular-nums">{fmtNum(getValue() as number)}</span>,
  },
  {
    accessorKey: 'budget_spent',
    header: 'Budget spent',
    meta: { className: 'text-right' },
    cell: ({ row }) => (
      <span className="tabular-nums">
        {fmtMoney(row.original.budget_spent)}
        <span className="text-xs font-normal text-muted-foreground"> / {fmtMoney(row.original.daily_budget_cap)}/d</span>
      </span>
    ),
  },
];

export default function AdminCampaignsAnalytics() {
  const { data, isPending, isError, error, refetch } = useCampaignPerformance();
  const [filters, setFilters] = useQueryStates({
    q: parseAsString.withDefault(''),
    status: parseAsString.withDefault('all'),
    mod: parseAsString.withDefault('all'),
    page: parseAsInteger.withDefault(1),
  });

  const rows = data ?? [];

  const moderationOptions = useMemo(() => {
    const statuses = [...new Set(rows.map(c => c.moderation_status).filter(Boolean))];
    return [{ value: 'all', label: 'All moderation' }, ...statuses.map(s => ({ value: s, label: s }))];
  }, [rows]);

  const filtered = useMemo(() => {
    const q = filters.q.trim().toLowerCase();
    return rows.filter(c => {
      if (filters.status === 'live' && !c.is_active) return false;
      if (filters.status === 'paused' && c.is_active) return false;
      if (filters.mod !== 'all' && c.moderation_status !== filters.mod) return false;
      if (!q) return true;
      return c.name.toLowerCase().includes(q) || c.company_name.toLowerCase().includes(q) || c.target_intent.toLowerCase().includes(q);
    });
  }, [rows, filters]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const page = Math.min(filters.page, totalPages);
  const pageRows = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  if (isPending) return <LoadingState label="Loading campaign performance…" />;
  if (isError) return <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />;

  const statusDist = filtered.reduce((acc, c) => {
    const k = c.is_active ? 'Live' : 'Paused';
    acc[k] = (acc[k] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);
  const pieData = Object.entries(statusDist).map(([name, value]) => ({ name, value }));
  const topCampaigns = [...filtered].sort((a, b) => (b.delivery_count ?? 0) - (a.delivery_count ?? 0)).slice(0, 5);

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle>Campaign status</CardTitle></CardHeader>
          <CardContent>
            <div className="h-56">
              {pieData.length === 0 ? (
                <div className="flex h-full items-center justify-center text-sm text-muted-foreground">No data</div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" innerRadius={56} outerRadius={78} paddingAngle={4} stroke="none">
                      {pieData.map((e, i) => <Cell key={i} fill={e.name === 'Live' ? chartColor(4) : chartColor(1)} />)}
                    </Pie>
                    <Tooltip {...chartTooltip} />
                  </PieChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle>Top campaigns by delivery</CardTitle></CardHeader>
          <CardContent>
            <div className="h-56">
              {topCampaigns.length === 0 ? (
                <div className="flex h-full items-center justify-center text-sm text-muted-foreground">No campaigns</div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={topCampaigns} layout="vertical" margin={{ top: 0, right: 24, left: 24, bottom: 0 }}>
                    <CartesianGrid {...chartGrid} horizontal={false} vertical />
                    <XAxis type="number" {...chartAxis} tickFormatter={v => (v >= 1000 ? `${(v / 1000).toFixed(1)}k` : String(v))} />
                    <YAxis dataKey="name" type="category" {...chartAxis} width={110} />
                    <Tooltip {...chartTooltip} formatter={(v) => [fmtNum(Number(v ?? 0)), 'Deliveries']} />
                    <Bar dataKey="delivery_count" name="Deliveries" fill={chartColor(0)} radius={[0, 6, 6, 0]} barSize={22} />
                  </BarChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="space-y-3">
        <div className="flex flex-wrap items-center gap-2">
          <div className="relative min-w-56 flex-1">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input value={filters.q} onChange={e => setFilters({ q: e.target.value || '', page: 1 })} placeholder="Search name, company, intent…" className="pl-9" />
          </div>
          <Select value={filters.status} onValueChange={v => setFilters({ status: v, page: 1 })}>
            <SelectTrigger className="w-40"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              <SelectItem value="live">Live only</SelectItem>
              <SelectItem value="paused">Paused only</SelectItem>
            </SelectContent>
          </Select>
          <Select value={filters.mod} onValueChange={v => setFilters({ mod: v, page: 1 })}>
            <SelectTrigger className="w-44"><SelectValue /></SelectTrigger>
            <SelectContent>
              {moderationOptions.map(o => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            onClick={() =>
              exportToCsv(
                filtered.map(c => ({
                  name: c.name, company: c.company_name, intent: c.target_intent,
                  status: c.is_active ? 'live' : 'paused', moderation: c.moderation_status,
                  validation: c.validation_status, deliveries: c.delivery_count,
                  unique_users: c.unique_users, budget_spent: c.budget_spent,
                })),
                'campaign-performance',
              )
            }
          >
            <Download className="size-4" /> Export CSV
          </Button>
        </div>

        <DataTable
          columns={columns}
          data={pageRows}
          emptyState={<span className="text-sm text-muted-foreground">No campaigns match your filters.</span>}
        />

        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span>{filtered.length} campaign{filtered.length === 1 ? '' : 's'}</span>
          <div className="flex items-center gap-2">
            <span className="tabular-nums">Page {page} / {totalPages}</span>
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setFilters({ page: page - 1 })}>Previous</Button>
            <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setFilters({ page: page + 1 })}>Next</Button>
          </div>
        </div>
      </div>
    </div>
  );
}
