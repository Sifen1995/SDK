import { useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Card, Badge, StatusPill, LoadingState, ErrorState, useQuery } from '@skykin/ui';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import { useSubscription } from '../context/SubscriptionContext';
import { formatDate, formatEtb } from '../lib/campaignUtils';
import { ROLE_META } from '../types';

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-3 gap-4 py-4">
      <dt className="text-sm font-medium text-muted-foreground">{label}</dt>
      <dd className="col-span-2 text-sm">{children}</dd>
    </div>
  );
}

export default function Profile() {
  const { user, refreshUser } = useAuth();
  const { subscribed, subscription } = useSubscription();
  const { data, isPending, isError, error, refetch } = useQuery({ queryKey: ['me'], queryFn: api.me });

  useEffect(() => { if (data?.user) refreshUser(data.user); }, [data, refreshUser]);

  if (isPending && !user) return <LoadingState label="Loading profile…" />;
  if (isError && !user) return <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />;
  if (!user) return null;

  const meta = ROLE_META[user.role] || {
    label: user.role || 'User',
    description: 'Portal user',
    canWrite: false,
    selfRegister: false,
  };

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h2 className="font-display text-lg font-semibold">Profile</h2>
        <p className="text-sm text-muted-foreground">Your advertiser portal account details.</p>
      </div>

      <Card className="overflow-hidden">
        <div className="brand-hero overflow-hidden px-6 py-8 text-white">
          <div className="pointer-events-none absolute inset-0 opacity-90" style={{ background: 'radial-gradient(circle at 90% 10%, rgb(255 255 255 / 0.12), transparent 45%)' }} />
          <div className="relative flex items-center gap-4">
            <div className="flex size-16 items-center justify-center rounded-2xl bg-white/20 text-2xl font-bold">{(user.name || 'U').charAt(0).toUpperCase()}</div>
            <div>
              <h3 className="font-display text-xl font-semibold">{user.name || 'Unknown User'}</h3>
              <p className="text-sm text-white/80">{user.email}</p>
              <Badge variant="secondary" className="mt-2 bg-white/20 text-white">{meta.label}</Badge>
            </div>
          </div>
        </div>

        <dl className="divide-y divide-border px-6">
          <Row label="Company">{user.company_name || '—'}</Row>
          <Row label="Subscription">
            {subscribed && subscription ? (
              <>
                <span className="font-medium">{subscription.plan.name}</span>
                <span className="text-muted-foreground"> · {formatEtb(subscription.plan.monthly_fee_etb)}/mo</span>
                <p className="mt-1 text-xs text-muted-foreground">Renews {formatDate(subscription.current_period_end)}</p>
              </>
            ) : (
              <>
                <span className="text-muted-foreground">Not subscribed</span>
                <Link to="/subscription" className="mt-1 block text-xs text-identity hover:underline">Choose a plan →</Link>
              </>
            )}
          </Row>
          <Row label="Role"><span className="font-medium">{meta.label}</span><p className="mt-1 text-muted-foreground">{meta.description}</p></Row>
          <Row label="Advertiser ID"><span className="break-all font-mono text-xs">{user.advertiser_id}</span></Row>
          <Row label="Account status"><StatusPill status={user.is_active ? 'active' : 'inactive'} /></Row>
        </dl>
      </Card>
    </div>
  );
}
