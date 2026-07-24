import { Link } from 'react-router-dom';
import { Card, CardContent, Button, Badge, StatusPill, LoadingState, ErrorState, EmptyState } from '@skykin/ui';
import { Package, Plus } from 'lucide-react';
import { useApplications } from '../lib/queries';

export default function Dashboard() {
  const { data: apps, isPending, isError, error, refetch } = useApplications();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h2 className="font-display text-lg font-semibold">Applications</h2>
          <p className="text-sm text-muted-foreground">Manage your applications and SDK credentials.</p>
        </div>
        <Button asChild><Link to="/applications/new"><Plus className="size-4" /> New application</Link></Button>
      </div>

      {isPending ? (
        <LoadingState label="Loading applications…" />
      ) : isError ? (
        <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />
      ) : apps.length === 0 ? (
        <EmptyState
          icon={Package}
          title="No applications yet"
          description="Create your first application to get SDK credentials."
          action={<Button asChild><Link to="/applications/new"><Plus className="size-4" /> Create application</Link></Button>}
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {apps.map(app => (
            <Card key={app.id}>
              <CardContent className="p-5">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h3 className="truncate font-semibold">{app.app_name}</h3>
                    <p className="mt-0.5 truncate font-mono text-sm text-muted-foreground">{app.bundle_id}</p>
                  </div>
                  <StatusPill status={app.status} />
                </div>
                <div className="mt-4 flex items-center justify-between">
                  <Badge variant="identity">{app.platform}</Badge>
                  <span className="text-xs text-muted-foreground">
                    Created {new Date(app.created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })}
                  </span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
