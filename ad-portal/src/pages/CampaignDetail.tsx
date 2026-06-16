import { useCallback, useEffect, useState } from 'react';
import { Link, useLocation, useParams } from 'react-router-dom';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import ActivateCampaignPanel from '../components/ActivateCampaignPanel';
import CampaignPreviewPanel from '../components/CampaignPreviewPanel';
import { ActiveBadge, ValidationBadge } from '../components/StatusBadge';
import { formatDate, formatLabel } from '../lib/campaignUtils';
import type { Campaign, CampaignPreview } from '../types';

export default function CampaignDetail() {
  const { id } = useParams<{ id: string }>();
  const location = useLocation();
  const highlightActivate = Boolean((location.state as { highlightActivate?: boolean } | null)?.highlightActivate);
  const { canWrite } = useAuth();
  const [campaign, setCampaign] = useState<Campaign | null>(null);
  const [preview, setPreview] = useState<CampaignPreview | null>(null);
  const [loading, setLoading] = useState(true);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [error, setError] = useState('');

  const loadPreview = useCallback(async (campaignId: string) => {
    setPreviewLoading(true);
    try {
      const p = await api.previewCampaign(campaignId);
      setPreview(p);
    } catch {
      setPreview(null);
    } finally {
      setPreviewLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;
    (async () => {
      try {
        const c = await api.getCampaign(id);
        if (cancelled) return;
        setCampaign(c);
        await loadPreview(id);
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Campaign not found');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [id, loadPreview]);

  async function handleActivate() {
    if (!id) throw new Error('Missing campaign id');
    const updated = await api.activateCampaign(id);
    setCampaign(updated);
    await loadPreview(id);
  }

  if (loading) {
    return <p className="text-muted">Loading campaign…</p>;
  }

  if (!campaign) {
    return (
      <div>
        <p className="text-red-600 dark:text-red-400">{error || 'Campaign not found'}</p>
        <Link to="/" className="text-sm mt-4 inline-block text-primary hover:underline">← Back</Link>
      </div>
    );
  }

  return (
    <div>
      <Link to="/" className="text-sm text-muted hover:text-primary font-medium transition">
        ← Back to campaigns
      </Link>

      <div className="mt-4 flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-bold text-primary">{campaign.name}</h1>
            <ActiveBadge active={campaign.isActive} />
            <ValidationBadge status={campaign.validationStatus} />
          </div>
          <p className="text-muted mt-1">
            {formatLabel(campaign.creativeFormat)} · {formatLabel(campaign.targetIntent)}
          </p>
        </div>
      </div>

      {error && (
        <div className="alert-error mt-4">{error}</div>
      )}

      <div className="mt-6">
        <ActivateCampaignPanel
          campaign={campaign}
          canWrite={canWrite}
          onActivate={handleActivate}
          highlight={highlightActivate && !campaign.isActive}
        />
      </div>

      {campaign.validationStatus === 'failed' && campaign.validationNotes && (
        <div className="alert-error mt-4">
          <strong>Validation failed:</strong> {campaign.validationNotes}
          {canWrite && (
            <Link to="/campaigns/new" className="block mt-2 underline font-medium">
              Create a corrected campaign →
            </Link>
          )}
        </div>
      )}

      <div className="grid lg:grid-cols-2 gap-8 mt-8">
        <div className="card p-6 space-y-4">
          <h2 className="font-semibold text-primary">Campaign details</h2>
          <dl className="space-y-3 text-sm">
            <div className="flex justify-between gap-4">
              <dt className="text-muted">ID</dt>
              <dd className="font-mono text-xs text-primary break-all text-right">{campaign.id}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-muted">Application ID</dt>
              <dd className="font-mono text-xs text-primary break-all text-right">{campaign.applicationId || '—'}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-muted">Destination URL</dt>
              <dd className="text-primary text-xs break-all text-right">
                {campaign.destinationUrl ? (
                  <a href={campaign.destinationUrl} target="_blank" rel="noreferrer" className="hover:underline">
                    {campaign.destinationUrl}
                  </a>
                ) : '—'}
              </dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-muted">Created</dt>
              <dd className="text-primary">{formatDate(campaign.createdAt)}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-muted">Daily budget</dt>
              <dd className="text-primary">{campaign.dailyBudgetCap || '—'}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-muted">Total budget</dt>
              <dd className="text-primary">{campaign.totalBudgetCap || '—'}</dd>
            </div>
            {campaign.title && (
              <div>
                <dt className="text-muted mb-1">Title</dt>
                <dd className="text-primary">{campaign.title}</dd>
              </div>
            )}
            {campaign.bodyText && (
              <div>
                <dt className="text-muted mb-1">Body</dt>
                <dd className="text-primary">{campaign.bodyText}</dd>
              </div>
            )}
            {campaign.imageUrl && (
              <div>
                <dt className="text-muted mb-2">Image URL</dt>
                <dd className="break-all text-xs font-mono text-muted">{campaign.imageUrl}</dd>
                <img
                  src={campaign.imageUrl}
                  alt=""
                  className="mt-2 max-h-40 rounded-lg border border-[var(--border)] object-contain"
                />
              </div>
            )}
          </dl>
        </div>

        <div>
          <h2 className="font-semibold text-primary mb-4">Creative preview</h2>
          <CampaignPreviewPanel preview={preview} loading={previewLoading} />
        </div>
      </div>
    </div>
  );
}
