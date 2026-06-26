import { useCallback, useEffect, useMemo, useState } from 'react';
import { api } from '../lib/api';
import type { Campaign } from '../types';
import CampaignModerationCard, { type ModerationAction } from '../components/CampaignModerationCard';
import { ConfirmModal, NotesModal } from '../components/Modal';
import OffsetPagination, { paginateSlice, PAGE_SIZE } from '../components/Pagination';

type Tab = 'pending' | 'ready';

type ModalState =
  | { type: 'none' }
  | { type: 'confirm'; action: 'go-live' | 'approve-and-go-live'; campaignId: string; campaignName: string }
  | { type: 'notes-approve'; campaignId: string }
  | { type: 'notes-reject'; campaignId: string };

export default function AdminPendingCampaigns() {
  const [tab, setTab] = useState<Tab>('pending');
  const [pending, setPending] = useState<Campaign[]>([]);
  const [ready, setReady] = useState<Campaign[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionError, setActionError] = useState('');
  const [processingId, setProcessingId] = useState<string | null>(null);
  const [pendingPage, setPendingPage] = useState(1);
  const [readyPage, setReadyPage] = useState(1);
  const [modal, setModal] = useState<ModalState>({ type: 'none' });

  const loadAll = useCallback(async () => {
    try {
      setLoading(true);
      setActionError('');
      const [pendingList, allRes] = await Promise.all([
        api.listPendingCampaigns(),
        api.listCampaigns(0, 500),
      ]);
      setPending(pendingList);
      setReady(
        allRes.campaigns.filter(
          c =>
            c.moderationStatus === 'approved' &&
            c.validationStatus === 'passed' &&
            !c.isActive,
        ),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load campaigns');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadAll();
  }, [loadAll]);

  const activeList = tab === 'pending' ? pending : ready;
  const activePage = tab === 'pending' ? pendingPage : readyPage;
  const setActivePage = tab === 'pending' ? setPendingPage : setReadyPage;
  const paginated = useMemo(
    () => paginateSlice(activeList, activePage, PAGE_SIZE),
    [activeList, activePage],
  );

  useEffect(() => {
    setActivePage(1);
  }, [tab, setActivePage]);

  function openAction(campaign: Campaign, action: ModerationAction) {
    setActionError('');
    switch (action) {
      case 'approve-and-go-live':
        setModal({
          type: 'confirm',
          action: 'approve-and-go-live',
          campaignId: campaign.id,
          campaignName: campaign.name,
        });
        break;
      case 'go-live':
        setModal({
          type: 'confirm',
          action: 'go-live',
          campaignId: campaign.id,
          campaignName: campaign.name,
        });
        break;
      case 'approve-only':
        setModal({ type: 'notes-approve', campaignId: campaign.id });
        break;
      case 'reject':
        setModal({ type: 'notes-reject', campaignId: campaign.id });
        break;
    }
  }

  async function runValidate(campaignId: string, action: 'approve' | 'reject', notes?: string) {
    setProcessingId(campaignId);
    try {
      await api.validateCampaign(campaignId, action, notes || undefined);
      setModal({ type: 'none' });
      await loadAll();
      if (action === 'approve') setTab('ready');
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setProcessingId(null);
    }
  }

  async function runActivate(campaignId: string) {
    setProcessingId(campaignId);
    try {
      await api.adminActivateCampaign(campaignId);
      setModal({ type: 'none' });
      await loadAll();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Activation failed');
    } finally {
      setProcessingId(null);
    }
  }

  async function handleConfirm() {
    if (modal.type !== 'confirm') return;
    const { campaignId, action } = modal;
    if (action === 'go-live') {
      await runActivate(campaignId);
      return;
    }
    setProcessingId(campaignId);
    try {
      await api.validateCampaign(campaignId, 'approve');
      await api.adminActivateCampaign(campaignId);
      setModal({ type: 'none' });
      await loadAll();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Approve & go live failed');
    } finally {
      setProcessingId(null);
    }
  }

  if (loading) return <div className="text-muted">Loading campaigns…</div>;
  if (error) return <div className="alert-error">{error}</div>;

  return (
    <div>
      <h1 className="text-xl font-bold text-primary mb-1">Campaign Moderation</h1>
      <p className="text-sm text-muted mb-5">
        Review pending submissions, approve moderation, and activate approved campaigns.
      </p>

      <div className="tab-bar mb-6">
        <button
          type="button"
          className={`tab-btn ${tab === 'pending' ? 'tab-btn-active' : ''}`}
          onClick={() => setTab('pending')}
        >
          Awaiting review
          {pending.length > 0 && <span className="tab-count">{pending.length}</span>}
        </button>
        <button
          type="button"
          className={`tab-btn ${tab === 'ready' ? 'tab-btn-active' : ''}`}
          onClick={() => setTab('ready')}
        >
          Ready to go live
          {ready.length > 0 && <span className="tab-count tab-count-ready">{ready.length}</span>}
        </button>
      </div>

      {actionError && <div className="alert-error mb-6">{actionError}</div>}

      {activeList.length === 0 ? (
        <div className="card-static p-12 text-center border-dashed">
          <p className="text-muted">
            {tab === 'pending'
              ? 'No campaigns awaiting moderation.'
              : 'No approved campaigns waiting for activation. Approve a campaign first, then activate it here.'}
          </p>
        </div>
      ) : (
        <>
          <div className="space-y-4">
            {paginated.map(c => (
              <CampaignModerationCard
                key={c.id}
                campaign={c}
                processing={processingId === c.id}
                mode={tab}
                onAction={action => openAction(c, action)}
              />
            ))}
          </div>
          <OffsetPagination
            page={activePage}
            totalItems={activeList.length}
            onPageChange={setActivePage}
          />
        </>
      )}

      <ConfirmModal
        open={modal.type === 'confirm'}
        title={
          modal.type === 'confirm' && modal.action === 'go-live'
            ? 'Activate campaign?'
            : 'Approve & go live?'
        }
        description={
          modal.type === 'confirm'
            ? modal.action === 'go-live'
              ? `"${modal.campaignName}" will go live and start delivering to matched users.`
              : `"${modal.campaignName}" will be approved and activated immediately.`
            : ''
        }
        confirmLabel={modal.type === 'confirm' && modal.action === 'go-live' ? 'Go Live' : 'Approve & Go Live'}
        variant="success"
        loading={processingId !== null}
        onConfirm={handleConfirm}
        onCancel={() => setModal({ type: 'none' })}
      />

      <NotesModal
        open={modal.type === 'notes-approve'}
        title="Approve campaign"
        description="Creative validation will run again. The campaign moves to Ready to go live for activation."
        label="Approval notes (optional)"
        placeholder="Looks good — brand guidelines met…"
        confirmLabel="Approve"
        variant="primary"
        loading={processingId !== null}
        onConfirm={notes =>
          modal.type === 'notes-approve' && runValidate(modal.campaignId, 'approve', notes)
        }
        onCancel={() => setModal({ type: 'none' })}
      />

      <NotesModal
        open={modal.type === 'notes-reject'}
        title="Reject campaign"
        description="The advertiser will see this campaign as rejected."
        label="Rejection reason"
        placeholder="Creative does not meet guidelines…"
        required
        confirmLabel="Reject campaign"
        variant="danger"
        loading={processingId !== null}
        onConfirm={notes =>
          modal.type === 'notes-reject' && runValidate(modal.campaignId, 'reject', notes)
        }
        onCancel={() => setModal({ type: 'none' })}
      />
    </div>
  );
}
