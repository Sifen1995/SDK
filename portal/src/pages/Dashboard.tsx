import { Link } from 'react-router-dom';
import { Card, CardContent, Button, Badge, StatusPill, LoadingState, ErrorState, EmptyState, TooltipProvider, Tooltip, TooltipTrigger, TooltipContent } from '@skykin/ui';
import { Package, Plus, Copy } from 'lucide-react';
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
            <Card key={app.id} className="overflow-hidden transition-all hover:border-primary/30 hover:shadow-md">
              <CardContent className="p-0 flex flex-col h-full">
                <div className="p-5 flex items-start justify-between gap-3 bg-gradient-to-b from-card to-card/50 flex-1">
                  <div className="min-w-0 flex items-center gap-3">
                    <div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary border border-primary/20">
                      <Package className="size-5" />
                    </div>
                    <div className="min-w-0">
                      <h3 className="truncate font-semibold text-base leading-tight mb-2">{app.app_name}</h3>
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div 
                               className="flex items-center gap-1.5 rounded-md border border-border/60 bg-background/50 px-2 py-1 shadow-sm hover:border-primary/40 hover:bg-card hover:text-foreground transition-all cursor-pointer w-fit"
                               onClick={(e) => { e.preventDefault(); navigator.clipboard.writeText(app.bundle_id); }}
                            >
                              <p className="truncate font-mono text-[11px] font-medium text-muted-foreground">{app.bundle_id}</p>
                              <Copy className="size-3 text-muted-foreground/70" />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent className="bg-foreground text-background">
                            <p className="text-xs font-medium">Click to copy Bundle ID</p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                  </div>
                  <StatusPill status={app.status} />
                </div>
                <div className="border-t border-border/50 bg-muted/20 px-5 py-3 flex items-center justify-between">
                  <Badge variant="identity" className="bg-identity/10 text-identity hover:bg-identity/20 border-identity/20">{app.platform}</Badge>
                  <span className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">
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
