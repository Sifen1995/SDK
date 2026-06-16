import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import { useSubscription } from '../context/SubscriptionContext';
import RoleBadge from '../components/RoleBadge';
import { formatDate, formatEtb } from '../lib/campaignUtils';
import { ROLE_META } from '../types';

export default function Profile() {
  const { user, refreshUser } = useAuth();
  const { subscribed, subscription } = useSubscription();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await api.me();
        if (!cancelled) refreshUser(res.user);
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Failed to load profile');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [refreshUser]);

  if (loading) {
    return <p className="text-muted">Loading profile…</p>;
  }

  if (!user) {
    return <p className="text-red-600 dark:text-red-400">{error || 'Not signed in'}</p>;
  }

  const meta = ROLE_META[user.role];

  return (
    <div className="max-w-2xl">
      <h1 className="text-2xl font-bold text-primary">Profile</h1>
      <p className="text-muted mt-1">Your advertiser portal account details</p>

      {error && (
        <div className="alert-warning mt-4">{error}</div>
      )}

      <div className="card mt-8 overflow-hidden">
        <div className="bg-gradient-to-r from-brand-600 to-brand-500 px-6 py-8">
          <div className="flex items-center gap-4">
            <div className="h-16 w-16 rounded-2xl bg-white/20 flex items-center justify-center text-2xl font-bold text-white">
              {user.name.charAt(0).toUpperCase()}
            </div>
            <div>
              <h2 className="text-xl font-semibold text-white">{user.name}</h2>
              <p className="text-brand-100 text-sm">{user.email}</p>
              <div className="mt-2">
                <RoleBadge role={user.role} />
              </div>
            </div>
          </div>
        </div>

        <dl className="divide-y divide-[var(--border)] px-6">
          <div className="py-4 grid grid-cols-3 gap-4">
            <dt className="text-sm font-medium text-muted">Company</dt>
            <dd className="text-sm text-primary col-span-2">{user.company_name || '—'}</dd>
          </div>
          <div className="py-4 grid grid-cols-3 gap-4">
            <dt className="text-sm font-medium text-muted">Subscription</dt>
            <dd className="text-sm text-primary col-span-2">
              {subscribed && subscription ? (
                <>
                  <span className="font-medium">{subscription.plan.name}</span>
                  <span className="text-muted"> · {formatEtb(subscription.plan.monthly_fee_etb)}/mo</span>
                  <p className="text-muted text-xs mt-1">
                    Renews {formatDate(subscription.current_period_end)}
                  </p>
                </>
              ) : (
                <>
                  <span className="text-muted">Not subscribed</span>
                  <Link to="/subscription" className="block text-brand-600 hover:underline text-xs mt-1">
                    Choose a plan →
                  </Link>
                </>
              )}
            </dd>
          </div>
          <div className="py-4 grid grid-cols-3 gap-4">
            <dt className="text-sm font-medium text-muted">Role</dt>
            <dd className="text-sm text-primary col-span-2">
              <span className="font-medium">{meta.label}</span>
              <p className="text-muted mt-1">{meta.description}</p>
            </dd>
          </div>
          <div className="py-4 grid grid-cols-3 gap-4">
            <dt className="text-sm font-medium text-muted">Advertiser ID</dt>
            <dd className="text-sm text-primary col-span-2 font-mono text-xs break-all">{user.advertiser_id}</dd>
          </div>
          <div className="py-4 grid grid-cols-3 gap-4">
            <dt className="text-sm font-medium text-muted">Account status</dt>
            <dd className="text-sm col-span-2">
              <span className={user.is_active ? 'text-emerald-600 dark:text-emerald-400 font-medium' : 'text-red-600 dark:text-red-400 font-medium'}>
                {user.is_active ? 'Active' : 'Inactive'}
              </span>
            </dd>
          </div>
          <div className="py-4 grid grid-cols-3 gap-4">
            <dt className="text-sm font-medium text-muted">Permissions</dt>
            <dd className="text-sm text-primary col-span-2">
              {meta.canWrite ? (
                <ul className="list-disc list-inside text-muted space-y-1">
                  <li>Create campaigns (requires subscription)</li>
                  <li>Preview creatives</li>
                  <li>Campaigns go live after operator moderation</li>
                </ul>
              ) : (
                <ul className="list-disc list-inside text-muted space-y-1">
                  <li>View campaigns and previews</li>
                  <li>Read-only access</li>
                </ul>
              )}
            </dd>
          </div>
        </dl>
      </div>
    </div>
  );
}
