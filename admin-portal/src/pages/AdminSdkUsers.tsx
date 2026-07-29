import { useQueryState, parseAsInteger } from 'nuqs';
import {
  Card, Button, Badge, DataTable, type ColumnDef, LoadingState, ErrorState,
} from '@skykin/ui';
import type { SdkUser } from '../types/sdkUsers';
import { fmtNum } from '../lib/format';
import { useSdkUsers } from '../lib/queries';

const PER_PAGE = 20;

const columns: ColumnDef<SdkUser>[] = [
  {
    accessorKey: 'user_id',
    header: 'User',
    cell: ({ getValue }) => <span className="font-mono text-xs">#{String(getValue())}</span>,
  },
  {
    id: 'intent',
    header: 'Latest intent',
    cell: ({ row }) =>
      row.original.latest_intent ? (
        <div className="flex items-center gap-2">
          <Badge variant="identity">{row.original.latest_intent.intent_name.replace(/_/g, ' ')}</Badge>
          <span className="text-xs tabular-nums text-muted-foreground">
            {(row.original.latest_intent.confidence * 100).toFixed(0)}%
          </span>
        </div>
      ) : (
        <span className="text-xs text-muted-foreground">—</span>
      ),
  },
  {
    id: 'predicted',
    header: 'Predicted',
    cell: ({ row }) =>
      row.original.latest_intent?.predicted_at
        ? <span className="text-xs text-muted-foreground">{new Date(row.original.latest_intent.predicted_at).toLocaleDateString()}</span>
        : <span className="text-xs text-muted-foreground">—</span>,
  },
  {
    accessorKey: 'created_at',
    header: 'First seen',
    meta: { className: 'text-right' },
    cell: ({ getValue }) => <span className="text-xs text-muted-foreground">{new Date(getValue() as string).toLocaleDateString()}</span>,
  },
];

export default function AdminSdkUsers() {
  const [page, setPage] = useQueryState('page', parseAsInteger.withDefault(1));
  const { data, isPending, isError, error, refetch, isFetching } = useSdkUsers(page, PER_PAGE);

  if (isPending) return <LoadingState label="Loading SDK users…" />;
  if (isError) return <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />;

  const totalPages = data.total_pages || 1;

  return (
    <div className="space-y-4">
      <div className="flex items-end justify-between">
        <div>
          <h2 className="font-display text-lg font-semibold">SDK users</h2>
          <p className="text-sm text-muted-foreground">End-users reached by the SDK and their latest predicted intent.</p>
        </div>
        <Card className="px-4 py-2">
          <p className="text-[11px] text-muted-foreground">Total users</p>
          <p className="font-display text-lg font-bold tabular-nums">{fmtNum(data.total)}</p>
        </Card>
      </div>

      <DataTable columns={columns} data={data.users} loading={isFetching} emptyState={<span className="text-sm text-muted-foreground">No SDK users yet.</span>} />

      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span className="tabular-nums">Page {data.page} / {totalPages}</span>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>Next</Button>
        </div>
      </div>
    </div>
  );
}
