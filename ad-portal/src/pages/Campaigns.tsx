import { useEffect, useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import { ActiveBadge, ValidationBadge } from '../components/StatusBadge';
import { formatDate, formatLabel } from '../lib/campaignUtils';
import type { Campaign } from '../types';

export default function Campaigns() {
  const { canWrite } = useAuth();
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activatingId, setActivatingId] = useState<string | null>(null);
  const [actionError, setActionError] = useState('');

  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 3;
  const [totalCount, setTotalCount] = useState<number | undefined>(undefined);
  
  const totalPages = totalCount !== undefined 
    ? Math.max(1, Math.ceil(totalCount / itemsPerPage))
    : Math.max(1, Math.ceil(campaigns.length / itemsPerPage));

  useEffect(() => {
    if (currentPage > totalPages && totalPages > 0) {
      setCurrentPage(totalPages);
    }
  }, [totalPages, currentPage]);

  const paginatedCampaigns = useMemo(() => {
    if (totalCount !== undefined) {
      return campaigns;
    }
    const startIndex = (currentPage - 1) * itemsPerPage;
    return campaigns.slice(startIndex, startIndex + itemsPerPage);
  }, [campaigns, currentPage, totalCount]);

  async function loadCampaigns(page = currentPage) {
    const offset = (page - 1) * itemsPerPage;
    const res = await api.listCampaigns(offset, itemsPerPage);
    setCampaigns(res.campaigns);
    if (res.total !== undefined) {
      setTotalCount(res.total);
    }
  }

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        setLoading(true);
        await loadCampaigns(currentPage);
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Failed to load campaigns');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [currentPage]);

  async function handleQuickActivate(c: Campaign) {
    if (!window.confirm(`Activate "${c.name}"? It will go live for ${formatLabel(c.targetIntent)}.`)) {
      return;
    }
    setActivatingId(c.id);
    setActionError('');
    try {
      await api.activateCampaign(c.id);
      await loadCampaigns();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Activation failed');
    } finally {
      setActivatingId(null);
    }
  }

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
        <div>
          <h1 className="text-2xl font-bold text-primary">Campaigns</h1>
          <p className="text-muted mt-1">Manage intent-targeted ad creatives</p>
        </div>
        {canWrite && (
          <Link to="/campaigns/new" className="btn-primary shrink-0">
            + New campaign
          </Link>
        )}
      </div>

      {!canWrite && (
        <div className="mb-6 rounded-xl border border-[var(--border)] bg-[var(--bg-subtle)] px-4 py-3 text-sm text-muted">
          You have read-only access. Contact an advertiser on your team to create or activate campaigns.
        </div>
      )}

      {error && <div className="alert-error mb-6">{error}</div>}
      {actionError && <div className="alert-error mb-6">{actionError}</div>}

      {loading ? (
        <p className="text-muted">Loading campaigns…</p>
      ) : campaigns.length === 0 ? (
        <div className="card p-12 text-center">
          <div className="logo-mark mx-auto h-16 w-16 rounded-2xl flex items-center justify-center text-2xl mb-4">
            ◆
          </div>
          <h2 className="text-lg font-semibold text-primary">No campaigns yet</h2>
          <p className="text-muted mt-2 max-w-sm mx-auto">
            Create your first campaign to target subscribers by predicted intent.
          </p>
          {canWrite && (
            <Link to="/campaigns/new" className="btn-primary mt-6 inline-flex">
              Create campaign
            </Link>
          )}
        </div>
      ) : (
        <div className="grid gap-4">
          {paginatedCampaigns.map(c => {
            const canActivate =
              canWrite && !c.isActive && c.validationStatus === 'passed';

            return (
              <div key={c.id} className="card p-5 hover:border-[var(--border-strong)] transition">
                <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
                  <Link to={`/campaigns/${c.id}`} className="min-w-0 flex-1 group">
                    <div className="flex flex-wrap items-center gap-2 mb-1">
                      <h2 className="font-semibold text-primary group-hover:underline truncate">
                        {c.name}
                      </h2>
                      <ActiveBadge active={c.isActive} />
                      <ValidationBadge status={c.validationStatus} />
                    </div>
                    <p className="text-sm text-muted">
                      {formatLabel(c.creativeFormat)} · {formatLabel(c.targetIntent)}
                    </p>
                    {c.validationNotes && c.validationStatus === 'failed' && (
                      <p className="text-xs text-red-600 dark:text-red-400 mt-1">{c.validationNotes}</p>
                    )}
                    <p className="text-xs text-faint mt-2">Created {formatDate(c.createdAt)}</p>
                  </Link>

                  <div className="flex flex-col sm:items-end gap-2 shrink-0">
                    {c.imageUrl && c.creativeFormat === 'BANNER' && (
                      <img
                        src={c.imageUrl}
                        alt=""
                        className="h-12 w-auto max-w-[120px] rounded border border-[var(--border)] object-cover"
                      />
                    )}
                    <div className="flex gap-2">
                      <Link to={`/campaigns/${c.id}`} className="btn-secondary text-xs py-2">
                        View
                      </Link>
                      {canActivate && (
                        <button
                          type="button"
                          className="btn-primary text-xs py-2"
                          disabled={activatingId === c.id}
                          onClick={() => handleQuickActivate(c)}
                        >
                          {activatingId === c.id ? 'Activating…' : 'Activate'}
                        </button>
                      )}
                      {canWrite && !c.isActive && c.validationStatus !== 'passed' && (
                        <Link
                          to={`/campaigns/${c.id}`}
                          className="btn-ghost text-xs"
                          title="Review validation before activating"
                        >
                          Review
                        </Link>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {!loading && (totalCount !== undefined ? totalCount > itemsPerPage : campaigns.length > itemsPerPage) && (
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4 border-t border-[var(--border)] pt-6 mt-6">
          <p className="text-sm text-muted">
            Showing <span className="font-medium text-primary">{(currentPage - 1) * itemsPerPage + 1}</span> to <span className="font-medium text-primary">{Math.min(currentPage * itemsPerPage, totalCount !== undefined ? totalCount : campaigns.length)}</span> of <span className="font-medium text-primary">{totalCount !== undefined ? totalCount : campaigns.length}</span> campaigns
          </p>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
              disabled={currentPage === 1}
              className="btn-secondary px-4 py-2 text-sm disabled:opacity-50"
            >
              ← Previous
            </button>
            <button
              onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
              className="btn-secondary px-4 py-2 text-sm disabled:opacity-50"
            >
              Next →
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
