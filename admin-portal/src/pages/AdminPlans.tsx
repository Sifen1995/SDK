import { useEffect, useState } from 'react';
import { useQueryState, parseAsStringLiteral } from 'nuqs';
import {
  Card, CardContent, Button, Badge, Input, Label, Separator,
  Tabs, TabsList, TabsTrigger,
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
  StatusPill, LoadingState, ErrorState, EmptyState, InlineError,
} from '@skykin/ui';
import { Plus, Layers, PackageOpen, Pencil, Ban, Receipt } from 'lucide-react';
import type { Plan, CreatePlanRequest, CreateSegmentRequest, AudienceSegment } from '../types';
import {
  usePlans, useSegments, useBillingRates,
  useCreatePlan, useUpdatePlan, useSuspendPlan, useCreateSegment, useSuspendSegment, useUpdateBillingRate,
} from '../lib/queries';

const emptyPlan: CreatePlanRequest = {
  name: '', monthly_fee_etb: 0, max_active_campaigns: 5, max_daily_budget_etb: 100,
  included_impressions: 10000, sms_plus_enabled: false, audiencemart_enabled: false, cpc_discount_pct: 0,
};

function PlanDialog({ open, onOpenChange, initial, planId }: { open: boolean; onOpenChange: (o: boolean) => void; initial: CreatePlanRequest; planId?: string }) {
  const create = useCreatePlan();
  const update = useUpdatePlan();
  const [form, setForm] = useState<CreatePlanRequest>(initial);
  useEffect(() => { if (open) setForm(initial); }, [open, initial]);
  const busy = create.isPending || update.isPending;
  const err = (create.error || update.error) as Error | null;

  function set<K extends keyof CreatePlanRequest>(k: K, v: CreatePlanRequest[K]) { setForm(f => ({ ...f, [k]: v })); }
  function submit() {
    const done = { onSuccess: () => onOpenChange(false) };
    if (planId) update.mutate({ id: planId, data: form }, done);
    else create.mutate(form, done);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>{planId ? 'Edit plan' : 'New plan'}</DialogTitle>
          <DialogDescription>Subscription tier limits and feature flags.</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1.5"><Label>Plan name</Label><Input value={form.name} onChange={e => set('name', e.target.value)} placeholder="Pro" /></div>
          <div className="space-y-1.5"><Label>Monthly fee (ETB)</Label><Input type="number" step="0.01" value={form.monthly_fee_etb || ''} onChange={e => set('monthly_fee_etb', parseFloat(e.target.value) || 0)} /></div>
          <div className="space-y-1.5"><Label>Max active campaigns</Label><Input type="number" value={form.max_active_campaigns} onChange={e => set('max_active_campaigns', parseInt(e.target.value) || 1)} /></div>
          <div className="space-y-1.5"><Label>Max daily budget (ETB)</Label><Input type="number" step="0.01" value={form.max_daily_budget_etb || ''} onChange={e => set('max_daily_budget_etb', parseFloat(e.target.value) || 0)} /></div>
          <div className="space-y-1.5"><Label>Included impressions</Label><Input type="number" value={form.included_impressions} onChange={e => set('included_impressions', parseInt(e.target.value) || 0)} /></div>
          <div className="space-y-1.5"><Label>CPC discount %</Label><Input type="number" step="0.1" value={form.cpc_discount_pct || ''} onChange={e => set('cpc_discount_pct', parseFloat(e.target.value) || 0)} /></div>
        </div>
        <div className="flex gap-4 pt-1">
          <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={form.sms_plus_enabled} onChange={e => set('sms_plus_enabled', e.target.checked)} /> SMS+ enabled</label>
          <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={form.audiencemart_enabled} onChange={e => set('audiencemart_enabled', e.target.checked)} /> AudienceMart enabled</label>
        </div>
        {err && <InlineError message={err.message} />}
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={submit} disabled={busy || !form.name.trim()}>{busy ? 'Saving…' : planId ? 'Save changes' : 'Create plan'}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function BillingRatesDialog({ planId, planName, onClose }: { planId: string; planName: string; onClose: () => void }) {
  const { data, isPending, isError, error } = useBillingRates(planId);
  const rates = data?.rates ?? [];
  const update = useUpdateBillingRate(planId);
  const [edits, setEdits] = useState<Record<string, string>>({});

  return (
    <Dialog open onOpenChange={o => !o && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Billing rates — {planName}</DialogTitle>
          <DialogDescription>Per-model rates charged against this plan.</DialogDescription>
        </DialogHeader>
        {isPending ? (
          <LoadingState label="Loading rates…" />
        ) : isError ? (
          <ErrorState message={(error as Error)?.message} />
        ) : !rates || rates.length === 0 ? (
          <p className="py-6 text-center text-sm text-muted-foreground">No billing rates configured for this plan.</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead>Model</TableHead>
                <TableHead>Rate (ETB)</TableHead>
                <TableHead>Status</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {rates.map(r => (
                <TableRow key={r.id}>
                  <TableCell className="font-medium">{r.billing_model}</TableCell>
                  <TableCell>
                    <Input
                      className="h-8 w-24"
                      value={edits[r.id] ?? String(r.rate_etb)}
                      onChange={e => setEdits(s => ({ ...s, [r.id]: e.target.value }))}
                    />
                  </TableCell>
                  <TableCell><StatusPill status={r.is_active ? 'active' : 'inactive'} /></TableCell>
                  <TableCell className="text-right">
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={update.isPending}
                      onClick={() => update.mutate({ id: r.id, data: { rate_etb: parseFloat(edits[r.id] ?? String(r.rate_etb)) || 0, is_active: r.is_active } })}
                    >
                      Save
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
        {update.isError && <InlineError message={(update.error as Error).message} />}
        <DialogFooter><Button variant="ghost" onClick={onClose}>Close</Button></DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function AdminPlans() {
  const [tab, setTab] = useQueryState('tab', parseAsStringLiteral(['plans', 'segments'] as const).withDefault('plans'));
  const plans = usePlans();
  const segments = useSegments();
  const suspendPlan = useSuspendPlan();
  const suspendSegment = useSuspendSegment();
  const createSegment = useCreateSegment();

  const [planDialog, setPlanDialog] = useState<{ open: boolean; initial: CreatePlanRequest; planId?: string }>({ open: false, initial: emptyPlan });
  const [ratesFor, setRatesFor] = useState<Plan | null>(null);
  const [suspendTarget, setSuspendTarget] = useState<{ kind: 'plan' | 'segment'; id: string; name: string } | null>(null);
  const [segDialog, setSegDialog] = useState(false);

  return (
    <div className="space-y-5">
      <div>
        <h2 className="font-display text-lg font-semibold">Plans &amp; catalog</h2>
        <p className="text-sm text-muted-foreground">Manage subscription plans, billing rates, and audience segments.</p>
      </div>

      <Tabs value={tab} onValueChange={v => setTab(v as 'plans' | 'segments')}>
        <TabsList>
          <TabsTrigger value="plans">Plans &amp; billing</TabsTrigger>
          <TabsTrigger value="segments">Audience segments</TabsTrigger>
        </TabsList>
      </Tabs>

      {tab === 'plans' && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold">Subscription plans</h3>
            <Button size="sm" onClick={() => setPlanDialog({ open: true, initial: emptyPlan })}><Plus className="size-4" /> New plan</Button>
          </div>

          {plans.isPending ? (
            <LoadingState label="Loading plans…" />
          ) : plans.isError ? (
            <ErrorState message={(plans.error as Error)?.message} onRetry={() => plans.refetch()} />
          ) : !plans.data || plans.data.length === 0 ? (
            <EmptyState icon={PackageOpen} title="No subscription plans yet" description="Create a plan to define tier limits and pricing." />
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {plans.data.map(p => (
                <Card key={p.id}>
                  <CardContent className="p-4">
                    <h3 className="text-sm font-semibold">{p.name}</h3>
                    <p className="mt-1 font-display text-lg font-bold tabular-nums text-identity">ETB {p.monthly_fee_etb}/mo</p>
                    <div className="mt-2 space-y-0.5 text-xs text-muted-foreground tabular-nums">
                      <p>Max {p.max_active_campaigns} active campaigns</p>
                      <p>Daily cap ETB {p.max_daily_budget_etb}</p>
                      <p>{p.included_impressions.toLocaleString()} impressions</p>
                    </div>
                    <div className="mt-3 flex flex-wrap gap-1.5">
                      {p.sms_plus_enabled && <Badge variant="secondary">SMS+</Badge>}
                      {p.audiencemart_enabled && <Badge variant="secondary">AudienceMart</Badge>}
                    </div>
                    <Separator className="my-3" />
                    <div className="flex flex-wrap gap-1.5">
                      <Button size="sm" variant="outline" onClick={() => setPlanDialog({ open: true, initial: { ...emptyPlan, ...p }, planId: p.id })}><Pencil className="size-3.5" /> Edit</Button>
                      <Button size="sm" variant="outline" onClick={() => setRatesFor(p)}><Receipt className="size-3.5" /> Rates</Button>
                      <Button size="sm" variant="outline" onClick={() => setSuspendTarget({ kind: 'plan', id: p.id, name: p.name })}><Ban className="size-3.5" /> Suspend</Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      )}

      {tab === 'segments' && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold">Audience segments {segments.data ? `(${segments.data.length})` : ''}</h3>
            <Button size="sm" onClick={() => setSegDialog(true)}><Plus className="size-4" /> New segment</Button>
          </div>
          {segments.isPending ? (
            <LoadingState label="Loading segments…" />
          ) : segments.isError ? (
            <ErrorState message={(segments.error as Error)?.message} onRetry={() => segments.refetch()} />
          ) : !segments.data || segments.data.length === 0 ? (
            <EmptyState icon={Layers} title="No audience segments yet" description="Create a segment or approve a candidate to populate the catalog." />
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {segments.data.map((seg: AudienceSegment) => (
                <Card key={seg.id}>
                  <CardContent className="p-4">
                    <h3 className="text-sm font-semibold">{seg.name}</h3>
                    <p className="mt-0.5 break-all font-mono text-[10px] text-muted-foreground">{seg.id}</p>
                    {seg.description && <p className="mt-1.5 text-xs text-muted-foreground">{seg.description}</p>}
                    <Separator className="my-3" />
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-muted-foreground">Price (ETB)</span>
                      <span className="font-display text-sm font-bold tabular-nums text-identity">{seg.price_etb}</span>
                    </div>
                    <Button size="sm" variant="outline" className="mt-3 w-full" onClick={() => setSuspendTarget({ kind: 'segment', id: seg.id, name: seg.name })}>
                      <Ban className="size-3.5" /> Suspend
                    </Button>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      )}

      <PlanDialog open={planDialog.open} onOpenChange={o => setPlanDialog(s => ({ ...s, open: o }))} initial={planDialog.initial} planId={planDialog.planId} />
      {ratesFor && <BillingRatesDialog planId={ratesFor.id} planName={ratesFor.name} onClose={() => setRatesFor(null)} />}
      <SegmentDialog open={segDialog} onOpenChange={setSegDialog} mutation={createSegment} />

      {/* Suspend confirm */}
      <Dialog open={!!suspendTarget} onOpenChange={o => !o && setSuspendTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Suspend {suspendTarget?.kind}?</DialogTitle>
            <DialogDescription>“{suspendTarget?.name}” will be suspended and no longer available.</DialogDescription>
          </DialogHeader>
          {(suspendPlan.isError || suspendSegment.isError) && (
            <InlineError message={((suspendPlan.error || suspendSegment.error) as Error).message} />
          )}
          <DialogFooter>
            <Button variant="ghost" onClick={() => setSuspendTarget(null)}>Cancel</Button>
            <Button
              variant="destructive"
              disabled={suspendPlan.isPending || suspendSegment.isPending}
              onClick={() => {
                if (!suspendTarget) return;
                const done = { onSuccess: () => setSuspendTarget(null) };
                if (suspendTarget.kind === 'plan') suspendPlan.mutate(suspendTarget.id, done);
                else suspendSegment.mutate(suspendTarget.id, done);
              }}
            >
              Suspend
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function SegmentDialog({ open, onOpenChange, mutation }: { open: boolean; onOpenChange: (o: boolean) => void; mutation: ReturnType<typeof useCreateSegment> }) {
  const [form, setForm] = useState<CreateSegmentRequest>({ name: '', description: '', top_intent_signals: [], approximate_size: 0, estimated_cpm: 0, is_active: true });
  const [signals, setSignals] = useState('');
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => { if (open) { setForm({ name: '', description: '', top_intent_signals: [], approximate_size: 0, estimated_cpm: 0, is_active: true }); setSignals(''); mutation.reset(); } }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>New segment</DialogTitle>
          <DialogDescription>Create a purchasable audience segment.</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1.5"><Label>Segment name</Label><Input value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="Coffee Lovers" /></div>
          <div className="space-y-1.5"><Label>Estimated CPM (ETB)</Label><Input type="number" step="0.01" value={form.estimated_cpm || ''} onChange={e => setForm(f => ({ ...f, estimated_cpm: parseFloat(e.target.value) || 0 }))} /></div>
        </div>
        <div className="space-y-1.5"><Label>Description</Label><Input value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} /></div>
        <div className="space-y-1.5"><Label>Intent signals (comma separated)</Label><Input value={signals} onChange={e => setSignals(e.target.value)} placeholder="crypto_interest, fintech_interest" /></div>
        {mutation.isError && <InlineError message={(mutation.error as Error).message} />}
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button
            disabled={mutation.isPending || !form.name.trim()}
            onClick={() => mutation.mutate(
              { ...form, top_intent_signals: signals.split(',').map(s => s.trim()).filter(Boolean) },
              { onSuccess: () => onOpenChange(false) },
            )}
          >
            {mutation.isPending ? 'Creating…' : 'Create segment'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
