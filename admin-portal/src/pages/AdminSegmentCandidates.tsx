import { useCallback, useEffect, useState } from 'react';
import { api } from '../lib/api';
import type { SegmentCandidate, ApproveSegmentCandidateRequest } from '../types';
import { NotesModal } from '../components/Modal';
import { CheckCircle, XCircle, Brain, Users, Calendar, BarChart3 } from 'lucide-react';

type ModalState =
  | { type: 'none' }
  | { type: 'approve'; candidate: SegmentCandidate }
  | { type: 'reject'; candidateId: string };

export default function AdminSegmentCandidates() {
  const [candidates, setCandidates] = useState<SegmentCandidate[]>([]);
  const [statusFilter, setStatusFilter] = useState<string>('pending');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionError, setActionError] = useState('');
  const [processingId, setProcessingId] = useState<string | null>(null);
  const [modal, setModal] = useState<ModalState>({ type: 'none' });
  const [runningAnalysis, setRunningAnalysis] = useState(false);

  // Approve form state
  const [approveName, setApproveName] = useState('');
  const [approveDesc, setApproveDesc] = useState('');
  const [approveCpm, setApproveCpm] = useState('');

  const loadCandidates = useCallback(async () => {
    try {
      setLoading(true);
      setActionError('');
      const list = await api.listSegmentCandidates(statusFilter);
      setCandidates(list ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load candidates');
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => { loadCandidates(); }, [loadCandidates]);

  async function handleRunAnalysis() {
    setRunningAnalysis(true);
    try {
      await api.runIntentConsistency();
      setActionError('');
      setTimeout(() => loadCandidates(), 2000);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Analysis failed');
    } finally {
      setRunningAnalysis(false);
    }
  }

  function openApprove(candidate: SegmentCandidate) {
    setApproveName(candidate.intent_name.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase()));
    setApproveDesc('');
    setApproveCpm('4.50');
    setModal({ type: 'approve', candidate });
  }

  async function handleApprove() {
    if (modal.type !== 'approve') return;
    const data: ApproveSegmentCandidateRequest = {
      name: approveName,
      description: approveDesc,
      estimated_cpm: parseFloat(approveCpm) || 0,
    };
    setProcessingId(modal.candidate.id);
    try {
      await api.approveSegmentCandidate(modal.candidate.id, data);
      setModal({ type: 'none' });
      await loadCandidates();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Approve failed');
    } finally {
      setProcessingId(null);
    }
  }

  async function handleReject(notes: string) {
    if (modal.type !== 'reject') return;
    setProcessingId(modal.candidateId);
    try {
      await api.rejectSegmentCandidate(modal.candidateId, notes);
      setModal({ type: 'none' });
      await loadCandidates();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Reject failed');
    } finally {
      setProcessingId(null);
    }
  }

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-5">
        <div>
          <h1 className="text-xl font-bold text-primary">Segment Candidates</h1>
          <p className="text-sm text-muted mt-0.5">Review AI-discovered audience segments from intent analysis.</p>
        </div>
        <button
          type="button"
          onClick={handleRunAnalysis}
          disabled={runningAnalysis}
          className="btn-primary text-xs"
        >
          <Brain size={14} />
          {runningAnalysis ? 'Analyzing…' : 'Run Intent Analysis'}
        </button>
      </div>

      <div className="tab-bar mb-5">
        {['pending', 'approved', 'rejected'].map(s => (
          <button
            key={s}
            type="button"
            className={`tab-btn text-xs ${statusFilter === s ? 'tab-btn-active' : ''}`}
            onClick={() => setStatusFilter(s)}
          >
            {s.charAt(0).toUpperCase() + s.slice(1)}
          </button>
        ))}
      </div>

      {actionError && <div className="alert-error mb-4 text-sm">{actionError}</div>}
      {error && <div className="alert-error mb-4 text-sm">{error}</div>}

      {loading ? (
        <div className="text-muted text-sm">Loading candidates…</div>
      ) : candidates.length === 0 ? (
        <div className="card-static p-10 text-center border-dashed">
          <p className="text-sm text-muted">No {statusFilter} segment candidates found.</p>
          {statusFilter === 'pending' && (
            <p className="text-xs text-faint mt-2">Run an intent consistency analysis to discover new candidates.</p>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {candidates.map(c => (
            <div key={c.id} className="card-static p-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="min-w-0">
                  <h3 className="text-sm font-semibold text-primary">{c.intent_name}</h3>
                  <div className="flex flex-wrap gap-3 mt-1.5 text-xs text-muted">
                    <span className="inline-flex items-center gap-1"><Users size={12} /> {c.user_count.toLocaleString()} users</span>
                    <span className="inline-flex items-center gap-1"><BarChart3 size={12} /> {(c.avg_confidence * 100).toFixed(1)}% confidence</span>
                    <span className="inline-flex items-center gap-1"><Calendar size={12} /> {c.avg_days_active.toFixed(1)} avg days</span>
                  </div>
                  <p className="text-[11px] text-faint mt-1">Scanned {new Date(c.scanned_at).toLocaleDateString()}</p>
                </div>
                {statusFilter === 'pending' && (
                  <div className="flex gap-2 shrink-0">
                    <button
                      type="button"
                      onClick={() => openApprove(c)}
                      disabled={processingId === c.id}
                      className="btn-success text-xs py-1.5 px-3"
                    >
                      <CheckCircle size={13} /> Approve
                    </button>
                    <button
                      type="button"
                      onClick={() => setModal({ type: 'reject', candidateId: c.id })}
                      disabled={processingId === c.id}
                      className="btn-danger-outline text-xs py-1.5 px-3"
                    >
                      <XCircle size={13} /> Reject
                    </button>
                  </div>
                )}
                {statusFilter !== 'pending' && (
                  <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                    c.status === 'approved'
                      ? 'bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                      : 'bg-red-50 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                  }`}>
                    {c.status}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Approve Modal */}
      {modal.type === 'approve' && (
        <div className="modal-backdrop" onClick={() => setModal({ type: 'none' })}>
          <div className="modal-card max-w-lg" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="text-base font-semibold text-primary">Approve & Publish Segment</h3>
              <p className="text-xs text-muted mt-1">This will create a purchasable audience segment from {modal.candidate.user_count} users.</p>
            </div>
            <div className="space-y-4 mb-4">
              <div>
                <label className="block text-xs font-medium text-primary mb-1">Segment Name</label>
                <input value={approveName} onChange={e => setApproveName(e.target.value)} className="field-input text-sm" />
              </div>
              <div>
                <label className="block text-xs font-medium text-primary mb-1">Description</label>
                <textarea value={approveDesc} onChange={e => setApproveDesc(e.target.value)} rows={2} className="field-input text-sm resize-y" placeholder="Users with sustained interest in…" />
              </div>
              <div>
                <label className="block text-xs font-medium text-primary mb-1">Estimated CPM (ETB)</label>
                <input type="number" step="0.01" value={approveCpm} onChange={e => setApproveCpm(e.target.value)} className="field-input text-sm" />
              </div>
            </div>
            <div className="modal-footer">
              <button type="button" className="btn-secondary text-xs" onClick={() => setModal({ type: 'none' })}>Cancel</button>
              <button type="button" className="btn-success text-xs" onClick={handleApprove} disabled={!approveName.trim() || processingId !== null}>
                {processingId ? 'Publishing…' : 'Approve & Publish'}
              </button>
            </div>
          </div>
        </div>
      )}

      <NotesModal
        open={modal.type === 'reject'}
        title="Reject candidate"
        description="This candidate will not be published as a segment."
        label="Rejection reason"
        placeholder="Insufficient user volume…"
        confirmLabel="Reject"
        variant="danger"
        loading={processingId !== null}
        onConfirm={handleReject}
        onCancel={() => setModal({ type: 'none' })}
      />
    </div>
  );
}
