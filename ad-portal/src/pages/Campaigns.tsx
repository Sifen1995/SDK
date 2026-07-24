import { Link } from 'react-router-dom';
import { useQueryState, parseAsInteger } from 'nuqs';
import {
  Card, CardContent, Button, KpiCard, StatusPill,
  LoadingState, ErrorState, EmptyState,
} from '@skykin/ui';
import { Megaphone, Send, CreditCard, Layers, Sparkles, ArrowRight, Plus } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useSubscription } from '../context/SubscriptionContext';
import { useCampaigns } from '../lib/queries';
import { formatDate, formatEtb, formatLabel } from '../lib/campaignUtils';

const PAGE_SIZE = 6;

function Progress({ value, max }: { value: number; max: number }) {
  const pct = max > 0 ? Math.min(100, Math.round((value / max) * 100)) : 0;
  return (
    <div className="h-1.5 w-full overflow-hidden rounded-full bg-muted">
      <div className="h-full rounded-full bg-primary transition-[width]" style={{ width: `${pct}%` }} />
    </div>
  );
}

export default function Campaigns() {
  const { canWrite } = useAuth();
  const { subscribed, subscription, loading: subLoading } = useSubscription();
  const [page, setPage] = useQueryState('page', parseAsInteger.withDefault(1));
  const offset = (page - 1) * PAGE_SIZE;
  const { data, isPending, isError, error, refetch } = useCampaigns(offset, PAGE_SIZE);

  const campaigns = data?.campaigns ?? [];
  const total = data?.total ?? campaigns.length;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const kpis = subscribed && subscription
    ? [
        { label: 'Total campaigns', value: total.toLocaleString(), icon: Megaphone },
        { label: 'Impressions used', value: subscription.impressions_used.toLocaleString(), sub: `of ${subscription.plan.included_impressions.toLocaleString()}`, icon: Send },
        { label: 'Current plan', value: subscription.plan.name, icon: CreditCard },
        { label: 'Campaign slots', value: String(subscription.plan.max_active_campaigns), sub: 'plan limit', icon: Layers },
      ]
    : [
        { label: 'Total campaigns', value: total.toLocaleString(), icon: Megaphone },
        { label: 'Subscription', value: 'None', sub: 'subscribe to launch', icon: CreditCard },
      ];

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {kpis.map(k => <KpiCard key={k.label} label={k.label} value={k.value} icon={k.icon} sub={k.sub} />)}
      </div>

      {/* Subscription upsell / callout */}
      {!subLoading && !subscribed && (
        <UpsellCallout
          eyebrow="Get started"
          title="Launch intent-targeted campaigns"
          body="Choose a plan to unlock campaign creation, audience segments and delivery analytics."
          cta="View plans" to="/subscription"
        />
      )}
      {subscribed && subscription && (
        <UpsellCallout
          eyebrow={`${subscription.plan.name} plan`}
          title="Need more reach?"
          body="Track monthly impression usage and upgrade any time for a larger audience and more campaign slots."
          cta="Manage plan" to="/subscription"
          usage={{ used: subscription.impressions_used, total: subscription.plan.included_impressions }}
        />
      )}

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="font-display text-lg font-semibold">Your campaigns</h2>
          <p className="text-sm text-muted-foreground">Manage intent-targeted ad creatives.</p>
        </div>
        {canWrite && (
          <Button asChild>
            <Link to={subscribed ? '/campaigns/new' : '/subscription'}>
              <Plus className="size-4" /> {subscribed ? 'New campaign' : 'Subscribe to create'}
            </Link>
          </Button>
        )}
      </div>

      {!canWrite && (
        <div className="rounded-lg border border-border bg-muted/50 px-4 py-3 text-sm text-muted-foreground">
          You have read-only access. Contact an advertiser on your team to create campaigns.
        </div>
      )}

      {isPending ? (
        <LoadingState label="Loading campaigns…" />
      ) : isError ? (
        <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />
      ) : campaigns.length === 0 ? (
        <EmptyState
          icon={Megaphone}
          title="No campaigns yet"
          description="Create your first campaign to target users by predicted intent."
          action={canWrite && (
            <Button asChild><Link to={subscribed ? '/campaigns/new' : '/subscription'}>{subscribed ? 'Create campaign' : 'Choose a plan'}</Link></Button>
          )}
        />
      ) : (
        <div className="grid gap-3">
          {campaigns.map(c => {
            const cap = c.dailyBudgetCap || c.totalBudgetCap || 0;
            return (
              <Card key={c.id}>
                <CardContent className="flex flex-col gap-4 p-4 sm:flex-row sm:items-start sm:justify-between">
                  <Link to={`/campaigns/${c.id}`} className="min-w-0 flex-1 group">
                    <div className="mb-1 flex flex-wrap items-center gap-2">
                      <h3 className="font-semibold group-hover:underline">{c.name}</h3>
                      <StatusPill status={c.isActive ? 'active' : 'inactive'} />
                      <StatusPill status={c.moderationStatus} />
                      <StatusPill status={c.validationStatus} />
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {formatLabel(c.channelCode || 'channel')} · {formatLabel(c.targetIntent)}{c.billingModel && ` · ${c.billingModel}`}
                    </p>
                    {c.moderationNotes && c.moderationStatus === 'rejected' && <p className="mt-1 text-xs text-destructive">{c.moderationNotes}</p>}
                    <p className="mt-2 text-xs text-muted-foreground">Created {formatDate(c.createdAt)}</p>
                  </Link>
                  <div className="w-full shrink-0 sm:w-52">
                    <div className="mb-1 flex items-center justify-between text-xs">
                      <span className="text-muted-foreground">Budget</span>
                      <span className="tabular-nums">{formatEtb(c.budgetSpent)} <span className="text-muted-foreground">/ {formatEtb(cap)}</span></span>
                    </div>
                    <Progress value={c.budgetSpent} max={cap} />
                    <Button asChild variant="outline" size="sm" className="mt-3 w-full">
                      <Link to={`/campaigns/${c.id}`}>View details</Link>
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {total > PAGE_SIZE && (
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span className="tabular-nums">Page {page} / {totalPages}</span>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
            <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>Next</Button>
          </div>
        </div>
      )}
    </div>
  );
}

function UpsellCallout({ eyebrow, title, body, cta, to, usage }: {
  eyebrow: string; title: string; body: string; cta: string; to: string;
  usage?: { used: number; total: number };
}) {
  const pct = usage && usage.total > 0 ? Math.min(100, Math.round((usage.used / usage.total) * 100)) : null;
  return (
    <div className="relative overflow-hidden rounded-xl bg-primary p-6 text-primary-foreground">
      <div className="pointer-events-none absolute inset-0 opacity-90" style={{ background: 'radial-gradient(circle at 12% 88%, rgb(255 255 255 / 0.12), transparent 45%)' }} />
      <div className="relative flex flex-col gap-4">
        <div>
          <span className="inline-flex items-center gap-1.5 rounded-full border border-white/25 bg-white/12 px-3 py-1 text-xs font-medium">
            <Sparkles className="size-3.5" /> {eyebrow}
          </span>
          <h3 className="mt-2 font-display text-lg font-bold">{title}</h3>
          <p className="mt-1 max-w-md text-sm text-white/80">{body}</p>
        </div>
        {pct !== null && usage && (
          <div>
            <div className="mb-1.5 flex items-center justify-between text-xs font-medium tabular-nums">
              <span>{usage.used.toLocaleString()} / {usage.total.toLocaleString()} impressions</span>
              <span>{pct}%</span>
            </div>
            <div className="h-2 overflow-hidden rounded-full bg-white/20"><div className="h-full rounded-full bg-white" style={{ width: `${pct}%` }} /></div>
          </div>
        )}
        <Button asChild variant="secondary" className="w-fit bg-white text-primary hover:bg-white/90">
          <Link to={to}>{cta} <ArrowRight className="size-4" /></Link>
        </Button>
      </div>
    </div>
  );
}
