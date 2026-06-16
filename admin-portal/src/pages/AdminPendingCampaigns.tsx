import { useEffect, useState } from 'react';
import { api } from '../lib/api';
import type { Campaign } from '../types';
import { formatDate } from '../lib/campaignUtils';

export default function AdminPendingCampaigns() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [processingId, setProcessingId] = useState<string | null>(null);

  async function loadPending() {
    try {
      setLoading(true);
      const data = await api.listPendingCampaigns();
      setCampaigns(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load pending campaigns');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadPending();
  }, []);

  async function handleValidate(id: string, action: 'approve' | 'reject') {
    const notes = action === 'reject' ? window.prompt('Reason for rejection:') : '';
    if (action === 'reject' && notes === null) return; // cancelled

    setProcessingId(id);
    try {
      await api.validateCampaign(id, action, notes || undefined);
      await loadPending();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Validation failed');
    } finally {
      setProcessingId(null);
    }
  }

  async function handleActivate(id: string) {
    if (!window.confirm('Are you sure you want to take this campaign LIVE?')) return;
    
    setProcessingId(id);
    try {
      await api.adminActivateCampaign(id);
      await loadPending();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Activation failed');
    } finally {
      setProcessingId(null);
    }
  }

  if (loading) return <div className="text-muted">Loading pending campaigns...</div>;
  if (error) return <div className="alert-error">{error}</div>;

  return (
    <div>
      <h1 className="text-2xl font-bold text-primary mb-2">Pending Campaigns</h1>
      <p className="text-muted mb-8">Review and moderate user-submitted campaigns.</p>

      {campaigns.length === 0 ? (
        <div className="card p-12 text-center border-dashed">
          <p className="text-muted">No pending campaigns to review.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {campaigns.map(c => (
            <div key={c.id} className="card p-6 bg-white dark:bg-[#1c1b22] border-[var(--border)]">
              <div className="flex flex-col md:flex-row justify-between gap-4">
                <div>
                  <h3 className="font-bold text-lg text-primary">{c.name}</h3>
                  <p className="text-sm text-muted mt-1">
                    Format: <span className="font-medium text-primary">{c.creativeFormat || 'N/A'}</span> &bull; 
                    Intent: <span className="font-medium text-primary">{c.targetIntent}</span> &bull; 
                    Budget: <span className="font-medium text-primary">${c.totalBudgetCap}</span>
                  </p>
                  <p className="text-xs text-faint mt-2">Advertiser ID: {c.advertiserId}</p>
                  <p className="text-xs text-faint">Submitted: {formatDate(c.createdAt)}</p>

                  <div className="mt-4 p-4 bg-[var(--bg-subtle)] rounded-lg border border-[var(--border)]">
                    <p className="text-xs font-semibold uppercase tracking-wider text-muted mb-2">Creative Payload</p>
                    {c.title && <p className="text-sm font-medium text-primary">{c.title}</p>}
                    {c.bodyText && <p className="text-sm text-muted mt-1">{c.bodyText}</p>}
                    {c.imageUrl && <img src={c.imageUrl} alt="creative" className="mt-2 max-h-32 rounded border border-[var(--border)]" />}
                    {c.destinationUrl && (
                      <p className="text-xs mt-2 truncate max-w-md">
                        <a href={c.destinationUrl} target="_blank" rel="noreferrer" className="text-brand-600 hover:underline">
                          {c.destinationUrl}
                        </a>
                      </p>
                    )}
                  </div>
                </div>

                <div className="flex flex-col gap-2 shrink-0 md:min-w-[150px]">
                  {c.validationStatus === 'pending' && (
                    <>
                      <button
                        onClick={() => handleValidate(c.id, 'approve')}
                        disabled={processingId === c.id}
                        className="btn-primary w-full bg-green-600 hover:bg-green-700 text-white"
                      >
                        Approve
                      </button>
                      <button
                        onClick={() => handleValidate(c.id, 'reject')}
                        disabled={processingId === c.id}
                        className="btn-secondary w-full text-red-600 hover:bg-red-50 hover:border-red-200 dark:hover:bg-red-900/20"
                      >
                        Reject
                      </button>
                    </>
                  )}
                  {c.validationStatus === 'passed' && !c.isActive && (
                    <button
                      onClick={() => handleActivate(c.id)}
                      disabled={processingId === c.id}
                      className="btn-primary w-full"
                    >
                      Go Live
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
