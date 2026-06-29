import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { useSubscription } from '../context/SubscriptionContext';
import { formatDate, formatEtb } from '../lib/campaignUtils';
import type { Plan } from '../types';

export default function SubscriptionPage() {
  const { subscribed, subscription, refresh } = useSubscription();
  const [plans, setPlans] = useState<Plan[]>([]);
  const [loading, setLoading] = useState(true);
  const [subscribingId, setSubscribingId] = useState<string | null>(null);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    api.listPlans()
      .then(setPlans)
      .catch(err => setError(err instanceof Error ? err.message : 'Failed to load plans'))
      .finally(() => setLoading(false));
  }, []);

  async function handleSubscribe(planId: string) {
    setSubscribingId(planId);
    setError('');
    setSuccess('');
    try {
      await api.subscribe(planId);
      await refresh();
      setSuccess('Subscription activated! You can now create campaigns.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Subscription failed');
    } finally {
      setSubscribingId(null);
    }
  }

  if (loading) {
    return <p className="text-muted">Loading plans…</p>;
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-primary">Subscription</h1>
        <p className="text-muted mt-1">
          Choose a plan to unlock campaign creation, delivery channels, and audience targeting.
        </p>
      </div>

      {error && <div className="alert-error mb-6">{error}</div>}
      {success && <div className="alert-success mb-6">{success}</div>}

      {subscribed && subscription && (
        <div className="subscription-active-banner mb-8">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div>
              <p className="text-xs font-semibold uppercase tracking-wider text-brand-300 mb-1">Current plan</p>
              <h2 className="text-xl font-bold text-white">{subscription.plan.name}</h2>
              <p className="text-sm text-white/80 mt-1">
                {formatEtb(subscription.plan.monthly_fee_etb)}/month ·{' '}
                {subscription.impressions_used.toLocaleString()} / {subscription.plan.included_impressions.toLocaleString()} impressions used
              </p>
              <p className="text-xs text-white/60 mt-2">
                Period ends {formatDate(subscription.current_period_end)}
              </p>
            </div>
            <Link to="/campaigns/new" className="btn-primary shrink-0 bg-white text-brand-700 hover:bg-brand-50">
              Create campaign →
            </Link>
          </div>
          <div className="mt-6 flex flex-wrap gap-2">
            {subscription.plan.audiencemart_enabled && (
              <span className="plan-feature-pill">Audiencemart segments</span>
            )}
            {subscription.plan.sms_plus_enabled && (
              <span className="plan-feature-pill">SMS+ channel</span>
            )}
            <span className="plan-feature-pill">
              Up to {subscription.plan.max_active_campaigns} active campaigns
            </span>
            <span className="plan-feature-pill">
              {formatEtb(subscription.plan.max_daily_budget_etb)} daily budget cap
            </span>
          </div>
        </div>
      )}

      {!subscribed && (
        <div className="mb-6 rounded-xl border border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/30 px-4 py-3 text-sm text-amber-800 dark:text-amber-300">
          Subscribe to a plan before creating campaigns. Campaign creation is blocked until you have an active subscription.
        </div>
      )}

      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
        {plans.map(plan => {
          const isCurrent = subscription?.plan.id === plan.id;
          return (
            <div
              key={plan.id}
              className={`plan-card ${isCurrent ? 'plan-card-current' : ''}`}
            >
              {isCurrent && (
                <span className="plan-card-badge">Your plan</span>
              )}
              <h3 className="text-lg font-bold text-primary">{plan.name}</h3>
              <p className="text-3xl font-bold text-brand-600 dark:text-brand-400 mt-2">
                {formatEtb(plan.monthly_fee_etb)}
                <span className="text-sm font-normal text-muted">/mo</span>
              </p>

              <ul className="mt-6 space-y-2 text-sm text-muted">
                <li>✓ {plan.included_impressions.toLocaleString()} included impressions</li>
                <li>✓ Up to {plan.max_active_campaigns} active campaigns</li>
                <li>✓ {formatEtb(plan.max_daily_budget_etb)} max daily budget</li>
                {plan.audiencemart_enabled ? (
                  <li>✓ Audiencemart audience segments</li>
                ) : (
                  <li className="text-faint">Intent-only targeting (no segments)</li>
                )}
                {plan.sms_plus_enabled && <li>✓ SMS+ premium channel</li>}
                {plan.cpc_discount_pct > 0 && (
                  <li>✓ {plan.cpc_discount_pct}% CPC discount</li>
                )}
              </ul>

              {!subscribed && (
                <button
                  type="button"
                  className="btn-primary w-full mt-6"
                  disabled={subscribingId === plan.id}
                  onClick={() => handleSubscribe(plan.id)}
                >
                  {subscribingId === plan.id ? 'Subscribing…' : `Choose ${plan.name}`}
                </button>
              )}
              {isCurrent && (
                <p className="text-center text-xs text-brand-600 dark:text-brand-400 font-medium mt-6">
                  Active subscription
                </p>
              )}
            </div>
          );
        })}
      </div>

      {plans.length === 0 && (
        <div className="card p-12 text-center text-muted">No plans available right now.</div>
      )}
    </div>
  );
}
