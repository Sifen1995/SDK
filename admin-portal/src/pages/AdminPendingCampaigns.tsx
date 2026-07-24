import { useState } from 'react';
import { useQueryState, parseAsStringLiteral } from 'nuqs';
import {
  Tabs, TabsList, TabsTrigger, Badge, Button, Input, Label,
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
  LoadingState, ErrorState, EmptyState, InlineError,
} from '@skykin/ui';
import { ShieldCheck } from 'lucide-react';
import type { Campaign } from '../types';
import CampaignModerationCard, { type ModerationAction } from '../components/CampaignModerationCard';
import { usePendingCampaigns, useReadyCampaigns, useModerateCampaign, useActivateCampaign, useApproveAndGoLive } from '../lib/queries';

type Modal =
  | { type: 'none' }
  | { type: 'confirm'; action: 'go-live' | 'approve-and-go-live'; c: Campaign }
  | { type: 'approve'; c: Campaign }
  | { type: 'reject'; c: Campaign };

export default function AdminPendingCampaigns() {
  const [tab, setTab] = useQueryState('tab', parseAsStringLiteral(['pending', 'ready'] as const).withDefault('pending'));
  const pending = usePendingCampaigns();
  const ready = useReadyCampaigns();
  const moderate = useModerateCampaign();
  const activate = useActivateCampaign();
  const approveGoLive = useApproveAndGoLive();

  const [modal, setModal] = useState<Modal>({ type: 'none' });
  const [notes, setNotes] = useState('');

  const busy = moderate.isPending || activate.isPending || approveGoLive.isPending;
  const list = tab === 'pending' ? pending : ready;
  const rows = (list.data ?? []) as Campaign[];

  function openAction(c: Campaign, action: ModerationAction) {
    setNotes('');
    if (action === 'approve-and-go-live') setModal({ type: 'confirm', action, c });
    else if (action === 'go-live') setModal({ type: 'confirm', action: 'go-live', c });
    else if (action === 'approve-only') setModal({ type: 'approve', c });
    else setModal({ type: 'reject', c });
  }

  function confirmAction() {
    if (modal.type !== 'confirm') return;
    const done = { onSuccess: () => setModal({ type: 'none' }) };
    if (modal.action === 'go-live') activate.mutate(modal.c.id, done);
    else approveGoLive.mutate(modal.c.id, { onSuccess: () => { setModal({ type: 'none' }); setTab('ready'); } });
  }

  function submitApprove() {
    if (modal.type !== 'approve') return;
    moderate.mutate(
      { id: modal.c.id, action: 'approve', notes: notes || undefined },
      { onSuccess: () => { setModal({ type: 'none' }); setTab('ready'); } },
    );
  }

  function submitReject() {
    if (modal.type !== 'reject') return;
    moderate.mutate(
      { id: modal.c.id, action: 'reject', notes },
      { onSuccess: () => setModal({ type: 'none' }) },
    );
  }

  const mutationError = (moderate.error || activate.error || approveGoLive.error) as Error | null;

  return (
    <div className="space-y-5">
      <div>
        <h2 className="font-display text-lg font-semibold">Campaign moderation</h2>
        <p className="text-sm text-muted-foreground">Review pending submissions, approve moderation, and activate approved campaigns.</p>
      </div>

      <Tabs value={tab} onValueChange={v => setTab(v as 'pending' | 'ready')}>
        <TabsList>
          <TabsTrigger value="pending">
            Awaiting review
            {(pending.data?.length ?? 0) > 0 && <Badge variant="warning" className="ml-1.5">{pending.data!.length}</Badge>}
          </TabsTrigger>
          <TabsTrigger value="ready">
            Ready to go live
            {(ready.data?.length ?? 0) > 0 && <Badge variant="success" className="ml-1.5">{ready.data!.length}</Badge>}
          </TabsTrigger>
        </TabsList>
      </Tabs>

      {mutationError && <InlineError message={mutationError.message} />}

      {list.isPending ? (
        <LoadingState label="Loading campaigns…" />
      ) : list.isError ? (
        <ErrorState message={(list.error as Error)?.message} onRetry={() => list.refetch()} />
      ) : rows.length === 0 ? (
        <EmptyState
          icon={ShieldCheck}
          title={tab === 'pending' ? 'Nothing awaiting moderation' : 'Nothing ready to activate'}
          description={tab === 'pending' ? 'New submissions will appear here for review.' : 'Approve a campaign first, then activate it here.'}
        />
      ) : (
        <div className="space-y-4">
          {rows.map(c => (
            <CampaignModerationCard key={c.id} campaign={c} processing={busy} mode={tab} onAction={a => openAction(c, a)} />
          ))}
        </div>
      )}

      {/* Confirm (go-live / approve+go-live) */}
      <Dialog open={modal.type === 'confirm'} onOpenChange={o => !o && setModal({ type: 'none' })}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{modal.type === 'confirm' && modal.action === 'go-live' ? 'Activate campaign?' : 'Approve & go live?'}</DialogTitle>
            <DialogDescription>
              {modal.type === 'confirm' && (modal.action === 'go-live'
                ? `"${modal.c.name}" will go live and start delivering to matched users.`
                : `"${modal.c.name}" will be approved and activated immediately.`)}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setModal({ type: 'none' })}>Cancel</Button>
            <Button onClick={confirmAction} disabled={busy}>
              {busy ? 'Working…' : modal.type === 'confirm' && modal.action === 'go-live' ? 'Go live' : 'Approve & go live'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Approve with optional notes */}
      <Dialog open={modal.type === 'approve'} onOpenChange={o => !o && setModal({ type: 'none' })}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Approve campaign</DialogTitle>
            <DialogDescription>Creative validation runs again; the campaign moves to “Ready to go live”.</DialogDescription>
          </DialogHeader>
          <div className="space-y-1.5">
            <Label htmlFor="approve-notes">Approval notes (optional)</Label>
            <Input id="approve-notes" value={notes} onChange={e => setNotes(e.target.value)} placeholder="Looks good — brand guidelines met…" />
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setModal({ type: 'none' })}>Cancel</Button>
            <Button onClick={submitApprove} disabled={busy}>{busy ? 'Approving…' : 'Approve'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Reject with required reason */}
      <Dialog open={modal.type === 'reject'} onOpenChange={o => !o && setModal({ type: 'none' })}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject campaign</DialogTitle>
            <DialogDescription>The advertiser will see this campaign as rejected.</DialogDescription>
          </DialogHeader>
          <div className="space-y-1.5">
            <Label htmlFor="reject-notes">Rejection reason</Label>
            <Input id="reject-notes" value={notes} onChange={e => setNotes(e.target.value)} placeholder="Creative does not meet guidelines…" />
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setModal({ type: 'none' })}>Cancel</Button>
            <Button variant="destructive" onClick={submitReject} disabled={busy || !notes.trim()}>{busy ? 'Rejecting…' : 'Reject campaign'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
