import { useState } from 'react';
import { useQueryState, parseAsString } from 'nuqs';
import {
  Card, CardContent, Button, Input, Label, Tabs, TabsList, TabsTrigger,
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
  LoadingState, ErrorState, EmptyState, InlineError, StatusPill,
} from '@skykin/ui';
import { CheckCircle, XCircle, BrainCircuit, Users, Calendar, BarChart3, FileCheck } from 'lucide-react';
import type { SegmentCandidate } from '../types';
import { useSegmentCandidates, useApproveCandidate, useRejectCandidate, useRunIntentConsistency } from '../lib/queries';

function titleCase(s: string) {
  return s.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
}

export default function AdminSegmentCandidates() {
  const [status, setStatus] = useQueryState('status', parseAsString.withDefault('pending'));
  const { data: candidates, isPending, isError, error, refetch } = useSegmentCandidates(status);

  const approve = useApproveCandidate();
  const reject = useRejectCandidate();
  const runAnalysis = useRunIntentConsistency();

  const [approveTarget, setApproveTarget] = useState<SegmentCandidate | null>(null);
  const [rejectTarget, setRejectTarget] = useState<SegmentCandidate | null>(null);
  const [form, setForm] = useState({ name: '', description: '', cpm: '4.50' });
  const [rejectNotes, setRejectNotes] = useState('');

  function openApprove(c: SegmentCandidate) {
    setForm({ name: titleCase(c.intent_name), description: '', cpm: '4.50' });
    approve.reset();
    setApproveTarget(c);
  }

  function submitApprove() {
    if (!approveTarget) return;
    approve.mutate(
      { id: approveTarget.id, data: { name: form.name, description: form.description, estimated_cpm: parseFloat(form.cpm) || 0 } },
      { onSuccess: () => setApproveTarget(null) },
    );
  }

  function submitReject() {
    if (!rejectTarget) return;
    reject.mutate(
      { id: rejectTarget.id, notes: rejectNotes },
      { onSuccess: () => { setRejectTarget(null); setRejectNotes(''); } },
    );
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="font-display text-lg font-semibold">Segment candidates</h2>
          <p className="text-sm text-muted-foreground">Review AI-discovered audience segments from intent analysis.</p>
        </div>
        <Button size="sm" onClick={() => runAnalysis.mutate()} disabled={runAnalysis.isPending}>
          <BrainCircuit className="size-4" />
          {runAnalysis.isPending ? 'Analyzing…' : 'Run intent analysis'}
        </Button>
      </div>

      {runAnalysis.isError && <InlineError message={(runAnalysis.error as Error).message} />}

      <Tabs value={status} onValueChange={setStatus}>
        <TabsList>
          <TabsTrigger value="pending">Pending</TabsTrigger>
          <TabsTrigger value="approved">Approved</TabsTrigger>
          <TabsTrigger value="rejected">Rejected</TabsTrigger>
        </TabsList>
      </Tabs>

      {isPending ? (
        <LoadingState label="Loading candidates…" />
      ) : isError ? (
        <ErrorState message={(error as Error)?.message} onRetry={() => refetch()} />
      ) : !candidates || candidates.length === 0 ? (
        <EmptyState
          icon={FileCheck}
          title={`No ${status} candidates`}
          description={status === 'pending' ? 'Run an intent-consistency analysis to discover new candidates.' : undefined}
        />
      ) : (
        <div className="space-y-3">
          {candidates.map(c => (
            <Card key={c.id}>
              <CardContent className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                  <h3 className="text-sm font-semibold">{titleCase(c.intent_name)}</h3>
                  <div className="mt-1.5 flex flex-wrap gap-3 text-xs text-muted-foreground">
                    <span className="inline-flex items-center gap-1 tabular-nums"><Users className="size-3" /> {c.user_count.toLocaleString()} users</span>
                    <span className="inline-flex items-center gap-1 tabular-nums"><BarChart3 className="size-3" /> {(c.avg_confidence * 100).toFixed(1)}% confidence</span>
                    <span className="inline-flex items-center gap-1 tabular-nums"><Calendar className="size-3" /> {c.avg_days_active.toFixed(1)} avg days</span>
                  </div>
                  <p className="mt-1 text-[11px] text-muted-foreground">Scanned {new Date(c.scanned_at).toLocaleDateString()}</p>
                </div>
                {status === 'pending' ? (
                  <div className="flex shrink-0 gap-2">
                    <Button variant="outline" size="sm" onClick={() => openApprove(c)}>
                      <CheckCircle className="size-4 text-success" /> Approve
                    </Button>
                    <Button variant="outline" size="sm" onClick={() => { reject.reset(); setRejectNotes(''); setRejectTarget(c); }}>
                      <XCircle className="size-4 text-destructive" /> Reject
                    </Button>
                  </div>
                ) : (
                  <StatusPill status={c.status} />
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Approve dialog */}
      <Dialog open={!!approveTarget} onOpenChange={o => !o && setApproveTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Approve &amp; publish segment</DialogTitle>
            <DialogDescription>
              Creates a purchasable audience segment{approveTarget ? ` from ${approveTarget.user_count.toLocaleString()} users` : ''}.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="seg-name">Segment name</Label>
              <Input id="seg-name" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="seg-desc">Description</Label>
              <Input id="seg-desc" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} placeholder="Users with sustained interest in…" />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="seg-cpm">Estimated CPM (ETB)</Label>
              <Input id="seg-cpm" type="number" step="0.01" value={form.cpm} onChange={e => setForm(f => ({ ...f, cpm: e.target.value }))} />
            </div>
            {approve.isError && <InlineError message={(approve.error as Error).message} />}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setApproveTarget(null)}>Cancel</Button>
            <Button onClick={submitApprove} disabled={!form.name.trim() || approve.isPending}>
              {approve.isPending ? 'Publishing…' : 'Approve & publish'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Reject dialog */}
      <Dialog open={!!rejectTarget} onOpenChange={o => !o && setRejectTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject candidate</DialogTitle>
            <DialogDescription>This candidate will not be published as a segment.</DialogDescription>
          </DialogHeader>
          <div className="space-y-1.5">
            <Label htmlFor="reject-notes">Rejection reason</Label>
            <Input id="reject-notes" value={rejectNotes} onChange={e => setRejectNotes(e.target.value)} placeholder="Insufficient user volume…" />
            {reject.isError && <InlineError message={(reject.error as Error).message} />}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setRejectTarget(null)}>Cancel</Button>
            <Button variant="destructive" onClick={submitReject} disabled={reject.isPending}>
              {reject.isPending ? 'Rejecting…' : 'Reject'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
