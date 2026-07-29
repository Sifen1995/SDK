import { useMemo } from 'react';
import { useQueryStates, parseAsString, parseAsInteger } from 'nuqs';
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts';
import { Download, Search } from 'lucide-react';
import {
  Card, CardHeader, CardTitle, CardContent, Button, Input, Badge,
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
  DataTable, type ColumnDef, StatusPill, LoadingState, ErrorState, exportToCsv,
  chartAxis, chartGrid, chartTooltip, chartColor,
} from '@skykin/ui';
import type { AdvertiserSummary } from '../types/analytics';
import { fmtEtb, fmtNum } from '../lib/format';
import { useAdvertisers } from '../lib/queries';

const PAGE_SIZE = 8;

const columns: ColumnDef<AdvertiserSummary>[] = [
  {
    accessorKey: 'company_name',
    header: 'Company',
    cell: ({ row }) => (
      <div>
        <p className="font-medium">{row.original.company_name}</p>
        <p className="mt-0.5 font-mono text-xs text-muted-foreground">{row.original.advertiser_id}</p>
      </div>
    ),
  },
  { accessorKey: 'plan_name', header: 'Plan', cell: ({ getValue }) => <Badge variant="identity">{(getValue() as string) || 'N/A'}</Badge> },
  { accessorKey: 'subscription_status', header: 'Status', cell: ({ getValue }) => <StatusPill status={(getValue() as string) || 'none'} /> },
  { accessorKey: 'campaign_count', header: 'Campaigns', meta: { className: 'text-right' }, cell: ({ getValue }) => <span className="tabular-nums">{fmtNum(getValue() as number)}</span> },
  { accessorKey: 'active_campaigns', header: 'Active', meta: { className: 'text-right' }, cell: ({ getValue }) => <span className="tabular-nums">{fmtNum(getValue() as number)}</span> },
  { accessorKey: 'total_deliveries', header: 'Deliveries', meta: { className: 'text-right' }, cell: ({ getValue }) => <span className="tabular-nums">{fmtNum(getValue() as number)}</span> },
  { accessorKey: 'segment_spend_etb', header: 'Segment spend', meta: { className: 'text-right' }, cell: ({ getValue }) => <span className="tabular-nums text-success">{fmtEtb(getValue() as number)}</span> },
];

export default function AdminAdvertisersAnalytics() {
  const { data, isPending, isError, error, refetch } = useAdvertisers();
  const [filters, setFilters] = useQueryStates({
    q: parseAsString.withDefault(''),
    status: parseAsString.withDefault('all'),
    page: parseAsInteger.withDefault(1),
  });
  const rows = data ?? [];

  const statusOptions = useMemo(() => {
    const s = [...new Set(rows.map(a => a.subscription_status || 'none'))];
    return [{ value: 'all', label: 'All subscriptions' }, ...s.map(v => ({ value: v, label: v }))];
  }, [rows]);

  const filtered = useMemo(() => {
    const q = filters.q.trim().toLowerCase();
    return rows.filter(a => {
      if (filters.status !== 'all' && (a.subscription_status || 'none') !== filters.status) return false;
      if (!q) return true;
      return a.company_name.toLowerCase().includes(q) || a.advertiser_id.toLowerCase().includes(q);
    });
  }, [rows, filters]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const page = Math.min(filters.page, totalPages);
  const pageRows = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  if (isPending) return <LoadingState label="Loading advertisers…" />;
  if (isError) return <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />;

  const statusDist = filtered.reduce((acc, a) => {
    const k = a.subscription_status || 'none';
    acc[k] = (acc[k] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);
  const pieData = Object.entries(statusDist).map(([name, value]) => ({ name, value }));
  const topSpenders = [...filtered].sort((a, b) => (b.segment_spend_etb ?? 0) - (a.segment_spend_etb ?? 0)).slice(0, 5);

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle>Subscription status</CardTitle></CardHeader>
          <CardContent>
            <div className="h-56">
              {pieData.length === 0 ? (
                <div className="flex h-full items-center justify-center text-sm text-muted-foreground">No data</div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" innerRadius={56} outerRadius={78} paddingAngle={4} stroke="none">
                      {pieData.map((_, i) => <Cell key={i} fill={chartColor(i)} />)}
                    </Pie>
                    <Tooltip {...chartTooltip} />
                  </PieChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle>Top spenders</CardTitle></CardHeader>
          <CardContent>
            <div className="h-56">
              {topSpenders.length === 0 ? (
                <div className="flex h-full items-center justify-center text-sm text-muted-foreground">No spenders</div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={topSpenders} layout="vertical" margin={{ top: 0, right: 24, left: 24, bottom: 0 }}>
                    <CartesianGrid {...chartGrid} horizontal={false} vertical />
                    <XAxis type="number" {...chartAxis} tickFormatter={v => `ETB ${v >= 1000 ? `${(v / 1000).toFixed(1)}k` : v}`} />
                    <YAxis dataKey="company_name" type="category" {...chartAxis} width={110} />
                    <Tooltip {...chartTooltip} formatter={(v) => [fmtEtb(Number(v ?? 0)), 'Spend']} />
                    <Bar dataKey="segment_spend_etb" name="Spend" fill={chartColor(0)} radius={[0, 6, 6, 0]} barSize={22} />
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
            <Input value={filters.q} onChange={e => setFilters({ q: e.target.value || '', page: 1 })} placeholder="Search company or ID…" className="pl-9" />
          </div>
          <Select value={filters.status} onValueChange={v => setFilters({ status: v, page: 1 })}>
            <SelectTrigger className="w-48"><SelectValue /></SelectTrigger>
            <SelectContent>
              {statusOptions.map(o => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            onClick={() => exportToCsv(filtered as unknown as Record<string, unknown>[], 'advertisers')}
          >
            <Download className="size-4" /> Export CSV
          </Button>
        </div>

        <DataTable
          columns={columns}
          data={pageRows}
          emptyState={<span className="text-sm text-muted-foreground">No advertisers match your filters.</span>}
        />

        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span>{filtered.length} advertiser{filtered.length === 1 ? '' : 's'}</span>
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
