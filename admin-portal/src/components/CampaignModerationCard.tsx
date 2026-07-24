import { Card, CardContent, Button, StatusPill } from '@skykin/ui';
import type { Campaign } from '../types';
import CampaignPreviewPanel from './CampaignPreviewPanel';
import { formatDate, formatLabel } from '../lib/campaignUtils';
import { useCampaignPreview } from '../lib/queries';

export type ModerationAction = 'approve-only' | 'reject' | 'go-live' | 'approve-and-go-live';

interface CampaignModerationCardProps {
  campaign: Campaign;
  processing: boolean;
  mode: 'pending' | 'ready';
  onAction: (action: ModerationAction) => void;
}

export default function CampaignModerationCard({ campaign: c, processing, mode, onAction }: CampaignModerationCardProps) {
  const { data: preview, isPending, isError, error } = useCampaignPreview(c.id);

  return (
    <Card>
      <CardContent className="p-5">
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="min-w-0">
            <div className="mb-2 flex flex-wrap items-center gap-2">
              <h3 className="font-display text-lg font-bold">{c.name}</h3>
              <StatusPill status={c.moderationStatus} />
              <StatusPill status={c.validationStatus} />
              {c.isActive && <StatusPill status="active" />}
            </div>
            <p className="text-sm text-muted-foreground">
              {formatLabel(c.channelCode || c.creativeFormat || 'channel')} · {formatLabel(c.targetIntent)}
              {c.billingModel && ` · ${c.billingModel}`} · <span className="font-medium text-foreground">${c.totalBudgetCap} total</span>
            </p>
            <p className="mt-2 text-xs text-muted-foreground">Advertiser {c.advertiserId} · Submitted {formatDate(c.createdAt)}</p>

            {c.moderationNotes && (
              <p className="mt-3 rounded-lg border border-border bg-muted/50 p-3 text-xs text-muted-foreground">
                <span className="font-medium text-foreground">Moderator notes:</span> {c.moderationNotes}
              </p>
            )}
            {c.validationNotes && c.validationStatus === 'failed' && (
              <p className="mt-2 text-xs text-destructive">Validation: {c.validationNotes}</p>
            )}
          </div>

          <CampaignPreviewPanel preview={preview} loading={isPending} error={isError ? (error as Error)?.message : undefined} />
        </div>

        <div className="mt-5 flex flex-wrap gap-2 border-t border-border pt-4">
          {mode === 'pending' && c.moderationStatus === 'pending' && (
            <>
              <Button onClick={() => onAction('approve-and-go-live')} disabled={processing}>Approve &amp; go live</Button>
              <Button variant="secondary" onClick={() => onAction('approve-only')} disabled={processing}>Approve only</Button>
              <Button variant="outline" onClick={() => onAction('reject')} disabled={processing}>Reject</Button>
            </>
          )}
          {mode === 'ready' && (
            <Button onClick={() => onAction('go-live')} disabled={processing}>
              {processing ? 'Activating…' : 'Go live'}
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
