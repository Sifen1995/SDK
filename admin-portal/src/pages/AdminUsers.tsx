import { useState, type FormEvent } from 'react';
import {
  Card, CardHeader, CardTitle, CardDescription, CardContent, Button, Input, Label,
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem, InlineError,
} from '@skykin/ui';
import { CheckCircle2 } from 'lucide-react';
import type { PortalRole } from '../types';
import { useCreateUser } from '../lib/queries';

export default function AdminUsers() {
  const create = useCreateUser();
  const [form, setForm] = useState({ name: '', email: '', password: '', companyName: '', role: 'operator_admin' as PortalRole });
  const [done, setDone] = useState<string | null>(null);

  function set<K extends keyof typeof form>(key: K, value: (typeof form)[K]) {
    setForm(f => ({ ...f, [key]: value }));
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setDone(null);
    create.mutate(
      { name: form.name, email: form.email, password: form.password, role: form.role, company_name: form.companyName || undefined },
      {
        onSuccess: () => {
          setDone(`User ${form.email} created successfully.`);
          setForm({ name: '', email: '', password: '', companyName: '', role: form.role });
        },
      },
    );
  }

  return (
    <div className="max-w-2xl">
      <Card>
        <CardHeader>
          <CardTitle>Create portal user</CardTitle>
          <CardDescription>Add operator admins or managed advertiser accounts.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-5">
            {done && (
              <div className="flex items-center gap-2 rounded-md border border-success/30 bg-success-surface px-3 py-2 text-sm text-success">
                <CheckCircle2 className="size-4" /> {done}
              </div>
            )}
            {create.isError && <InlineError message={(create.error as Error).message} />}

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-1.5">
                <Label htmlFor="name">Full name</Label>
                <Input id="name" required value={form.name} onChange={e => set('name', e.target.value)} placeholder="Jane Doe" />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="email">Email address</Label>
                <Input id="email" type="email" required value={form.email} onChange={e => set('email', e.target.value)} placeholder="jane@skykin.io" />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="role">Role</Label>
              <Select value={form.role} onValueChange={v => set('role', v as PortalRole)}>
                <SelectTrigger id="role"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="operator_admin">Operator Admin (full access)</SelectItem>
                  <SelectItem value="advertiser">Advertiser</SelectItem>
                  <SelectItem value="read_only_analyst">Read-only Analyst</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {form.role !== 'operator_admin' && (
              <div className="space-y-1.5">
                <Label htmlFor="company">Company name</Label>
                <Input id="company" required value={form.companyName} onChange={e => set('companyName', e.target.value)} placeholder="Acme Corp" />
              </div>
            )}

            <div className="space-y-1.5">
              <Label htmlFor="password">Temporary password</Label>
              <Input id="password" type="password" required minLength={8} value={form.password} onChange={e => set('password', e.target.value)} />
              <p className="text-xs text-muted-foreground">At least 8 characters. The user should change it after signing in.</p>
            </div>

            <div className="border-t border-border pt-4">
              <Button type="submit" disabled={create.isPending}>{create.isPending ? 'Creating…' : 'Create user'}</Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
