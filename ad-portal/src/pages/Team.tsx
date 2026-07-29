import { useState, type FormEvent } from 'react';
import {
  Card, CardHeader, CardTitle, CardDescription, CardContent, Button, Input, Label, Badge, InlineError, EmptyState, cn,
} from '@skykin/ui';
import { CheckCircle2, Lock } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useCreateTeamUser } from '../lib/queries';
import { ROLE_META, type PortalRole } from '../types';

const ADMIN_ROLES: PortalRole[] = ['advertiser', 'read_only_analyst', 'operator_admin'];

export default function Team() {
  const { user, isAdmin } = useAuth();
  const create = useCreateTeamUser();
  const [form, setForm] = useState({ name: '', email: '', password: '', companyName: user?.company_name ?? '' });
  const [role, setRole] = useState<PortalRole>('advertiser');
  const [success, setSuccess] = useState('');

  if (!isAdmin) {
    return (
      <div className="mx-auto max-w-lg">
        <EmptyState icon={Lock} title="Access restricted" description="Only operator admins can manage team accounts." />
      </div>
    );
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSuccess('');
    create.mutate(
      { name: form.name, email: form.email, password: form.password, company_name: form.companyName || undefined, role },
      {
        onSuccess: res => {
          setSuccess(`Created account for ${res.user.email} (${ROLE_META[res.user.role].label})`);
          setForm(f => ({ ...f, name: '', email: '', password: '' }));
        },
      },
    );
  }

  return (
    <div className="max-w-2xl">
      <Card>
        <CardHeader>
          <CardTitle>Team management</CardTitle>
          <CardDescription>Provision advertiser, analyst, or operator admin accounts.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-5">
            {create.isError && <InlineError message={(create.error as Error).message} />}
            {success && (
              <div className="flex items-center gap-2 rounded-md border border-success/30 bg-success-surface px-3 py-2 text-sm text-success">
                <CheckCircle2 className="size-4" /> {success}
              </div>
            )}

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-1.5"><Label htmlFor="name">Name</Label><Input id="name" required value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} /></div>
              <div className="space-y-1.5"><Label htmlFor="email">Email</Label><Input id="email" type="email" required value={form.email} onChange={e => setForm(f => ({ ...f, email: e.target.value }))} /></div>
              <div className="space-y-1.5 sm:col-span-2"><Label htmlFor="password">Password</Label><Input id="password" type="password" required minLength={8} value={form.password} onChange={e => setForm(f => ({ ...f, password: e.target.value }))} /></div>
              <div className="space-y-1.5 sm:col-span-2"><Label htmlFor="company">Company name</Label><Input id="company" value={form.companyName} onChange={e => setForm(f => ({ ...f, companyName: e.target.value }))} /></div>
            </div>

            <div>
              <Label className="mb-2 block">Role</Label>
              <div className="space-y-2">
                {ADMIN_ROLES.map(r => (
                  <label key={r} className={cn('flex cursor-pointer items-start gap-3 rounded-lg border p-4 transition-colors', role === r ? 'border-identity bg-identity/5 ring-1 ring-identity' : 'border-border')}>
                    <input type="radio" name="role" value={r} checked={role === r} onChange={() => setRole(r)} className="mt-1 accent-[var(--identity)]" />
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium">{ROLE_META[r].label}</span>
                        <Badge variant="secondary">{r.replace(/_/g, ' ')}</Badge>
                      </div>
                      <p className="mt-1 text-xs text-muted-foreground">{ROLE_META[r].description}</p>
                    </div>
                  </label>
                ))}
              </div>
            </div>

            <Button type="submit" disabled={create.isPending}>{create.isPending ? 'Creating…' : 'Create user'}</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
