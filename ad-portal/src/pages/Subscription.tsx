import { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Card, CardContent, Button, Badge, LoadingState, ErrorState, EmptyState, InlineError,
} from '@skykin/ui';
import { Check, ArrowRight, CreditCard } from 'lucide-react';
import { useSubscription } from '../context/SubscriptionContext';
import { usePlans, useSubscribe } from '../lib/queries';
import { formatDate, formatEtb } from '../lib/campaignUtils';

export default function SubscriptionPage() {
  const { subscribed, subscription, refresh } = useSubscription();
  const { data: plans, isPending, isError, error, refetch } = usePlans();
  const subscribe = useSubscribe();
  const [success, setSuccess] = useState('');

  function handleSubscribe(planId: string) {
    setSuccess('');
    subscribe.mutate(planId, {
      onSuccess: async () => { await refresh(); setSuccess('Subscription activated — you can now create campaigns.'); },
    });
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="font-display text-lg font-semibold">Subscription</h2>
        <p className="text-sm text-muted-foreground">Choose a plan to unlock campaign creation, delivery channels, and audience targeting.</p>
      </div>

      {success && <div className="rounded-md border border-success/30 bg-success-surface px-3 py-2 text-sm text-success">{success}</div>}
      {subscribe.isError && <InlineError message={(subscribe.error as Error).message} />}

      {subscribed && subscription && (
        <div className="relative overflow-hidden rounded-xl bg-primary p-6 text-primary-foreground">
          <div className="pointer-events-none absolute inset-0 opacity-90" style={{ background: 'radial-gradient(circle at 12% 88%, rgb(255 255 255 / 0.12), transparent 45%)' }} />
          <div className="relative flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="mb-1 text-xs font-semibold uppercase tracking-wider text-white/70">Current plan</p>
              <h3 className="font-display text-xl font-bold">{subscription.plan.name}</h3>
              <p className="mt-1 text-sm text-white/85 tabular-nums">
                {formatEtb(subscription.plan.monthly_fee_etb)}/mo · {subscription.impressions_used.toLocaleString()} / {subscription.plan.included_impressions.toLocaleString()} impressions used
              </p>
              <p className="mt-2 text-xs text-white/60">Period ends {formatDate(subscription.current_period_end)}</p>
            </div>
            <Button asChild variant="secondary" className="bg-white text-primary hover:bg-white/90">
              <Link to="/campaigns/new">Create campaign <ArrowRight className="size-4" /></Link>
            </Button>
          </div>
          <div className="relative mt-5 flex flex-wrap gap-2 text-xs">
            {subscription.plan.audiencemart_enabled && <span className="rounded-full border border-white/25 bg-white/12 px-3 py-1">AudienceMart segments</span>}
            {subscription.plan.sms_plus_enabled && <span className="rounded-full border border-white/25 bg-white/12 px-3 py-1">SMS+ channel</span>}
            <span className="rounded-full border border-white/25 bg-white/12 px-3 py-1">Up to {subscription.plan.max_active_campaigns} active campaigns</span>
            <span className="rounded-full border border-white/25 bg-white/12 px-3 py-1">{formatEtb(subscription.plan.max_daily_budget_etb)} daily cap</span>
          </div>
        </div>
      )}

      {!subscribed && (
        <div className="rounded-lg border border-warning/30 bg-warning-surface px-4 py-3 text-sm text-foreground">
          Subscribe to a plan before creating campaigns — creation stays locked until you have an active subscription.
        </div>
      )}

      {isPending ? (
        <LoadingState label="Loading plans…" />
      ) : isError ? (
        <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />
      ) : !plans || plans.length === 0 ? (
        <EmptyState icon={CreditCard} title="No plans available" description="Subscription plans will appear here once configured." />
      ) : (
        <div className="grid gap-5 md:grid-cols-2 lg:grid-cols-3">
          {plans.map(plan => {
            const isCurrent = subscription?.plan.id === plan.id;
            return (
              <Card key={plan.id} className={isCurrent ? 'ring-1 ring-identity' : undefined}>
                <CardContent className="p-5">
                  <div className="flex items-center justify-between">
                    <h3 className="font-display text-base font-bold">{plan.name}</h3>
                    {isCurrent && <Badge variant="identity">Your plan</Badge>}
                  </div>
                  <p className="mt-2 font-display text-3xl font-bold tabular-nums text-identity">
                    {formatEtb(plan.monthly_fee_etb)}<span className="text-sm font-normal text-muted-foreground">/mo</span>
                  </p>
                  <ul className="mt-5 space-y-2 text-sm text-muted-foreground">
                    <li className="flex items-center gap-2"><Check className="size-4 text-success" /> {plan.included_impressions.toLocaleString()} impressions</li>
                    <li className="flex items-center gap-2"><Check className="size-4 text-success" /> Up to {plan.max_active_campaigns} active campaigns</li>
                    <li className="flex items-center gap-2"><Check className="size-4 text-success" /> {formatEtb(plan.max_daily_budget_etb)} daily budget</li>
                    <li className="flex items-center gap-2">{plan.audiencemart_enabled ? <><Check className="size-4 text-success" /> AudienceMart segments</> : <span className="pl-6">Intent-only targeting</span>}</li>
                    {plan.sms_plus_enabled && <li className="flex items-center gap-2"><Check className="size-4 text-success" /> SMS+ premium channel</li>}
                    {plan.cpc_discount_pct > 0 && <li className="flex items-center gap-2"><Check className="size-4 text-success" /> {plan.cpc_discount_pct}% CPC discount</li>}
                  </ul>
                  {!subscribed && (
                    <Button className="mt-6 w-full" disabled={subscribe.isPending} onClick={() => handleSubscribe(plan.id)}>
                      {subscribe.isPending && subscribe.variables === plan.id ? 'Subscribing…' : `Choose ${plan.name}`}
                    </Button>
                  )}
                  {isCurrent && <p className="mt-6 text-center text-xs font-medium text-identity">Active subscription</p>}
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
