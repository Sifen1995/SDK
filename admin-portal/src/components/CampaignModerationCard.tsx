import { useEffect, useState } from 'react';
import { api } from '../lib/api';
import type { Campaign, CampaignPreview } from '../types';
import { ActiveBadge, ModerationBadge, ValidationBadge } from './StatusBadge';
import CampaignPreviewPanel from './CampaignPreviewPanel';
import { formatDate, formatLabel } from '../lib/campaignUtils';

export type ModerationAction = 'approve-only' | 'reject' | 'go-live' | 'approve-and-go-live';

interface CampaignModerationCardProps {
  campaign: Campaign;
  processing: boolean;
  mode: 'pending' | 'ready';
  onAction: (action: ModerationAction) => void;
}

export default function CampaignModerationCard({
  campaign: c,
  processing,
  mode,
  onAction,
}: CampaignModerationCardProps) {
  const [preview, setPreview] = useState<CampaignPreview | null>(null);
  const [previewLoading, setPreviewLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setPreviewLoading(true);
    api.previewCampaign(c.id)
      .then(p => { if (!cancelled) setPreview(p); })
      .catch(() => { if (!cancelled) setPreview(null); })
      .finally(() => { if (!cancelled) setPreviewLoading(false); });
    return () => { cancelled = true; };
  }, [c.id]);

  return (
    <div className="moderation-card">
      <div className="grid lg:grid-cols-2 gap-6">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2 mb-2">
            <h3 className="font-bold text-lg text-primary">{c.name}</h3>
            <ModerationBadge status={c.moderationStatus} />
            <ValidationBadge status={c.validationStatus} />
            {c.isActive && <ActiveBadge active />}
          </div>

          <p className="text-sm text-muted">
            {formatLabel(c.channelCode || c.creativeFormat || 'channel')} · {formatLabel(c.targetIntent)}
            {c.billingModel && ` · ${c.billingModel}`}
            {' · '}
            <span className="font-medium text-primary">${c.totalBudgetCap} total</span>
          </p>
          <p className="text-xs text-faint mt-2">
            Advertiser {c.advertiserId} · Submitted {formatDate(c.createdAt)}
          </p>

          {c.moderationNotes && (
            <p className="text-xs text-muted mt-3 p-3 rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)]">
              <span className="font-medium text-primary">Moderator notes:</span> {c.moderationNotes}
            </p>
          )}

          {c.validationNotes && c.validationStatus === 'failed' && (
            <p className="text-xs text-red-600 dark:text-red-400 mt-2">Validation: {c.validationNotes}</p>
          )}
        </div>

        <CampaignPreviewPanel preview={preview} loading={previewLoading} />
      </div>

      <div className="moderation-card-actions">
        {mode === 'pending' && c.moderationStatus === 'pending' && (
          <>
            <button
              type="button"
              onClick={() => onAction('approve-and-go-live')}
              disabled={processing}
              className="btn-primary"
            >
              Approve &amp; Go Live
            </button>
            <button
              type="button"
              onClick={() => onAction('approve-only')}
              disabled={processing}
              className="btn-secondary"
            >
              Approve only
            </button>
            <button
              type="button"
              onClick={() => onAction('reject')}
              disabled={processing}
              className="btn-danger-outline"
            >
              Reject
            </button>
          </>
        )}

        {mode === 'ready' && (
          <button
            type="button"
            onClick={() => onAction('go-live')}
            disabled={processing}
            className="btn-success"
          >
            {processing ? 'Activating…' : 'Go Live'}
          </button>
        )}
      </div>
    </div>
  );
}
