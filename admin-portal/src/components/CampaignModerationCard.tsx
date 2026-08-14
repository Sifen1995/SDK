import { Card, CardContent, Button, StatusPill } from '@skykin/ui';
import { cn } from '@skykin/ui'; // adjust path if this file's utils import differs from KpiCard's
import { MapPin } from 'lucide-react';
import type { Campaign } from '../types';
import CampaignPreviewPanel from './CampaignPreviewPanel';
import { formatDate, formatLabel } from '../lib/campaignUtils';
import { useCampaignPreview, useCampaignZones, useActivateCampaignZones } from '../lib/queries';

export type ModerationAction = 'approve-only' | 'reject' | 'go-live' | 'approve-and-go-live';

interface CampaignModerationCardProps {
  campaign: Campaign;
  processing: boolean;
  mode: 'pending' | 'ready';
  onAction: (action: ModerationAction) => void;
}

export default function CampaignModerationCard({ campaign: c, processing, mode, onAction }: CampaignModerationCardProps) {
  const { data: preview, isPending, isError, error } = useCampaignPreview(c.id);
  const zones = useCampaignZones(c.id).data ?? [];
  const draftZoneCount = zones.filter(z => !z.is_active).length;
  const activateZones = useActivateCampaignZones();

  // left-rail accent: quick-scan status signal, independent of the brand blob
  const accent = c.isActive
    ? 'bg-success'
    : c.moderationStatus === 'rejected' || c.validationStatus === 'failed'
      ? 'bg-destructive'
      : 'bg-primary';

  return (
    <Card
      className={cn(
        'relative overflow-hidden p-0',
        'shadow-[0_2px_4px_rgba(8,38,62,0.08),0_16px_36px_-12px_rgba(8,38,62,0.22),inset_0_1px_0_0_rgba(255,255,255,0.7)]',
        'border border-primary/40',
      )}
    >
      {/* status accent rail */}
      <span aria-hidden className={cn('absolute inset-y-0 left-0 w-1', accent)} />

      {/* signature brand blob, same motif as KpiCard */}
      <span
        aria-hidden
        className="pointer-events-none absolute -right-12 -top-16 size-56 rounded-full bg-primary/10 blur-3xl"
      />

      <CardContent className="relative p-5 pl-6">
        <div className="relative grid gap-6 lg:grid-cols-2">
          {/* divider between info + preview, centered in the gap at lg */}
          <span
            aria-hidden
            className="absolute inset-y-0 left-1/2 hidden w-px -translate-x-1/2 bg-border lg:block"
          />

          <div className="min-w-0">
            <div className="mb-2 flex flex-wrap items-center gap-2">
              <h3 className="font-display text-lg font-bold">{c.name}</h3>
              <StatusPill status={c.moderationStatus} />
              <StatusPill status={c.validationStatus} />
              {c.isActive && <StatusPill status="active" />}
            </div>
            <p className="text-sm text-muted-foreground">
              {formatLabel(c.channelCode || c.creativeFormat || 'channel')} · {formatLabel(c.targetIntent)}
              {' · '}<span className="font-medium text-foreground">${c.totalBudgetCap} total</span>
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

            {zones.length > 0 && (
              <div className="mt-3 rounded-lg border border-border bg-muted/40 p-3">
                <div className="flex items-center gap-1.5 text-xs font-semibold">
                  <MapPin className="size-3.5" />
                  {zones.length} store zone{zones.length === 1 ? '' : 's'} linked
                </div>
                <ul className="mt-2 space-y-1">
                  {zones.map(zone => (
                    <li key={zone.id} className="flex items-center justify-between gap-2 text-xs">
                      <span className="font-mono tabular-nums text-muted-foreground">
                        {zone.latitude.toFixed(4)}, {zone.longitude.toFixed(4)} · {zone.radius_metres.toLocaleString()} m
                      </span>
                      <StatusPill status={zone.is_active ? 'active' : 'pending'} />
                    </li>
                  ))}
                </ul>
                <p className="mt-2 text-xs text-muted-foreground">
                  {draftZoneCount === 0
                    ? 'All linked zones are live.'
                    : c.moderationStatus === 'pending'
                      ? `Approving this campaign activates ${draftZoneCount} draft zone${draftZoneCount === 1 ? '' : 's'}.`
                      : `${draftZoneCount} zone${draftZoneCount === 1 ? ' was' : 's were'} linked after approval and ${draftZoneCount === 1 ? 'is' : 'are'} still inactive.`}
                </p>
                {/* Only offered once the campaign is past moderation — before
                    that, approving does this anyway. */}
                {draftZoneCount > 0 && c.moderationStatus !== 'pending' && (
                  <Button
                    variant="secondary"
                    size="sm"
                    className="mt-2"
                    disabled={activateZones.isPending}
                    onClick={() => activateZones.mutate(c.id)}
                  >
                    {activateZones.isPending ? 'Activating…' : 'Activate linked zones'}
                  </Button>
                )}
              </div>
            )}
          </div>

          {/* recessed surface for the preview — drop this wrapper if CampaignPreviewPanel already renders its own border/surface */}
          <div className="rounded-xl border border-border bg-muted/30 p-3 shadow-[inset_0_1px_3px_rgba(8,38,62,0.06)]">
            <CampaignPreviewPanel preview={preview} loading={isPending} error={isError ? (error as Error)?.message : undefined} />
          </div>
        </div>

        {/* full-bleed action strip — clipped to the card's actual radius via overflow-hidden above */}
        <div className="-mx-5 -mb-5 mt-5 flex flex-wrap items-center gap-2 border-t border-primary/15 bg-muted/30 px-5 py-4">
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