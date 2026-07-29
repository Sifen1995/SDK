import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { Link, useParams } from 'react-router-dom';
import { api } from '../lib/api';
import CampaignStatusPanel from '../components/CampaignStatusPanel';
import CampaignPreviewPanel from '../components/CampaignPreviewPanel';
import { ActiveBadge, ModerationBadge, ValidationBadge } from '../components/StatusBadge';
import { formatDate, formatEtb, formatLabel } from '../lib/campaignUtils';
import type { Campaign, CampaignPreview } from '../types';

export default function CampaignDetail() {
  const { id } = useParams<{ id: string }>();
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
            <h1 className="text-xl font-bold text-primary">{campaign.name}</h1>
            <ActiveBadge active={campaign.isActive} />
            <ModerationBadge status={campaign.moderationStatus} />
            <ValidationBadge status={campaign.validationStatus} />
          </div>
          <p className="text-muted mt-1">
            {formatLabel(campaign.channelCode || 'channel')} · {formatLabel(campaign.targetIntent)}
          </p>
        </div>
      </div>

      {error && <div className="alert-error mt-4">{error}</div>}

      <div className="mt-6">
        <CampaignStatusPanel campaign={campaign} />
      </div>

      {campaign.validationStatus === 'failed' && campaign.validationNotes && (
        <div className="alert-error mt-4">
          <strong>Validation failed:</strong> {campaign.validationNotes}
          <Link to="/campaigns/new" className="block mt-2 underline font-medium">
            Create a corrected campaign →
          </Link>
        </div>
      )}

      <div className="grid lg:grid-cols-2 gap-8 mt-8">
        <div className="card p-6 space-y-4">
          <h2 className="font-semibold text-primary">Campaign details</h2>
          <dl className="space-y-3 text-sm">
            <DetailRow label="ID" value={campaign.id} mono />
            <DetailRow label="Channel" value={formatLabel(campaign.channelCode)} />
            <DetailRow label="Frequency cap" value={`${campaign.frequencyCapPerDay}/day`} />
            <DetailRow
              label="Destination"
              value={
                campaign.destinationUrl ? (
                  <a href={campaign.destinationUrl} target="_blank" rel="noreferrer" className="hover:underline break-all">
                    {campaign.destinationUrl}
                  </a>
                ) : '—'
              }
            />
            <DetailRow label="Created" value={formatDate(campaign.createdAt)} />
            <DetailRow label="Daily budget" value={formatEtb(campaign.dailyBudgetCap)} />
            <DetailRow label="Total budget" value={formatEtb(campaign.totalBudgetCap)} />
            <DetailRow label="Spent" value={formatEtb(campaign.budgetSpent)} />
            {campaign.segmentId && (
              <DetailRow label="Segment ID" value={campaign.segmentId} mono />
            )}
            {campaign.title && <DetailRow label="Title" value={campaign.title} />}
            {campaign.bodyText && <DetailRow label="Body" value={campaign.bodyText} />}
            {campaign.moderationNotes && (
              <DetailRow label="Moderation notes" value={campaign.moderationNotes} />
            )}
            {campaign.imageUrl && (
              <div>
                <dt className="text-muted mb-2">Image</dt>
                <dd>
                  <img
                    src={campaign.imageUrl}
                    alt=""
                    className="max-h-40 rounded-lg border border-[var(--border)] object-contain"
                  />
                </dd>
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

function DetailRow({
  label,
  value,
  mono = false,
}: {
  label: string;
  value: ReactNode;
  mono?: boolean;
}) {
  return (
    <div className="flex justify-between gap-4">
      <dt className="text-muted shrink-0">{label}</dt>
      <dd className={`text-primary text-right ${mono ? 'font-mono text-xs break-all' : ''}`}>{value}</dd>
    </div>
  );
}
