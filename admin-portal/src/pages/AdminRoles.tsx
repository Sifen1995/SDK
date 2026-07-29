import { useState } from 'react';
import {
  Card, CardContent, Button, Badge, Input, Label, Separator,
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
  LoadingState, ErrorState, EmptyState, InlineError,
} from '@skykin/ui';
import { Plus, X, ShieldCheck, KeyRound } from 'lucide-react';
import type { Role } from '../types/rbac';
import { usePermissions, useRoles, useCreateRole, useAssignPermission, useRevokePermission } from '../lib/queries';

export default function AdminRoles() {
  const roles = useRoles();
  const permissions = usePermissions();
  const createRole = useCreateRole();
  const assign = useAssignPermission();
  const revoke = useRevokePermission();

  const [newRole, setNewRole] = useState(false);
  const [form, setForm] = useState({ name: '', description: '' });
  const [assignFor, setAssignFor] = useState<Role | null>(null);
  const [selectedPerm, setSelectedPerm] = useState('');

  const allPerms = permissions.data ?? [];

  function submitRole() {
    createRole.mutate(form, {
      onSuccess: () => { setNewRole(false); setForm({ name: '', description: '' }); },
    });
  }

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-display text-lg font-semibold">Roles &amp; permissions</h2>
          <p className="text-sm text-muted-foreground">Define operator roles and the permissions granted to each.</p>
        </div>
        <Button size="sm" onClick={() => { createRole.reset(); setForm({ name: '', description: '' }); setNewRole(true); }}>
          <Plus className="size-4" /> New role
        </Button>
      </div>

      {roles.isPending ? (
        <LoadingState label="Loading roles…" />
      ) : roles.isError ? (
        <ErrorState message={(roles.error as Error)?.message} onRetry={() => roles.refetch()} />
      ) : !roles.data || roles.data.length === 0 ? (
        <EmptyState icon={ShieldCheck} title="No roles defined" description="Create a role to start granting permissions." />
      ) : (
        <div className="grid gap-4 lg:grid-cols-2">
          {roles.data.map(role => (
            <Card key={role.id}>
              <CardContent className="p-5">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-display text-base font-semibold">{role.name}</h3>
                      {role.is_system && <Badge variant="secondary">System</Badge>}
                    </div>
                    {role.description && <p className="mt-0.5 text-sm text-muted-foreground">{role.description}</p>}
                  </div>
                  {!role.is_system && (
                    <Button size="sm" variant="outline" onClick={() => { assign.reset(); setSelectedPerm(''); setAssignFor(role); }}>
                      <KeyRound className="size-3.5" /> Grant
                    </Button>
                  )}
                </div>

                <Separator className="my-3" />

                {role.permissions.length === 0 ? (
                  <p className="text-xs text-muted-foreground">No permissions granted.</p>
                ) : (
                  <div className="flex flex-wrap gap-1.5">
                    {role.permissions.map(p => (
                      <span key={p.id} className="inline-flex items-center gap-1 rounded-full bg-muted px-2.5 py-0.5 text-xs">
                        <span className="font-medium">{p.name}</span>
                        <span className="text-muted-foreground">· {p.resource}:{p.action}</span>
                        {!role.is_system && (
                          <button
                            type="button"
                            aria-label={`Revoke ${p.name}`}
                            className="ml-0.5 rounded-full p-0.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                            disabled={revoke.isPending}
                            onClick={() => revoke.mutate({ roleId: role.id, permissionId: p.id })}
                          >
                            <X className="size-3" />
                          </button>
                        )}
                      </span>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {revoke.isError && <InlineError message={(revoke.error as Error).message} />}

      {/* New role dialog */}
      <Dialog open={newRole} onOpenChange={setNewRole}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New role</DialogTitle>
            <DialogDescription>Create a custom operator role. Grant permissions after creating it.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-1.5"><Label htmlFor="role-name">Name</Label><Input id="role-name" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="Campaign Reviewer" /></div>
            <div className="space-y-1.5"><Label htmlFor="role-desc">Description</Label><Input id="role-desc" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} placeholder="Can review and moderate campaigns" /></div>
            {createRole.isError && <InlineError message={(createRole.error as Error).message} />}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setNewRole(false)}>Cancel</Button>
            <Button onClick={submitRole} disabled={createRole.isPending || form.name.trim().length < 2}>{createRole.isPending ? 'Creating…' : 'Create role'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Assign permission dialog */}
      <Dialog open={!!assignFor} onOpenChange={o => !o && setAssignFor(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Grant permission{assignFor ? ` — ${assignFor.name}` : ''}</DialogTitle>
            <DialogDescription>Add a permission to this role.</DialogDescription>
          </DialogHeader>
          <div className="space-y-1.5">
            <Label>Permission</Label>
            <Select value={selectedPerm} onValueChange={setSelectedPerm}>
              <SelectTrigger><SelectValue placeholder="Select a permission…" /></SelectTrigger>
              <SelectContent>
                {allPerms
                  .filter(p => !assignFor?.permissions.some(ap => ap.id === p.id))
                  .map(p => <SelectItem key={p.id} value={p.id}>{p.name} · {p.resource}:{p.action}</SelectItem>)}
              </SelectContent>
            </Select>
            {assign.isError && <InlineError message={(assign.error as Error).message} />}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setAssignFor(null)}>Cancel</Button>
            <Button
              disabled={!selectedPerm || assign.isPending}
              onClick={() => assignFor && assign.mutate({ roleId: assignFor.id, permissionId: selectedPerm }, { onSuccess: () => setAssignFor(null) })}
            >
              {assign.isPending ? 'Granting…' : 'Grant permission'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
