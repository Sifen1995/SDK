import { type ReactNode } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Card, CardHeader, CardTitle, CardContent, Button, StatusPill, LoadingState, ErrorState, InlineError } from '@skykin/ui';
import { ArrowLeft } from 'lucide-react';
import CampaignPreviewPanel from '../components/CampaignPreviewPanel';
import { formatDate, formatEtb, formatLabel } from '../lib/campaignUtils';
import { useCampaign, useCampaignPreview, useCampaignZones } from '../lib/queries';
import { CAMPAIGNS_PATH } from '../routes';

export default function CampaignDetail() {
  const { id = '' } = useParams<{ id: string }>();
  const { data: campaign, isPending, isError, error, refetch } = useCampaign(id);
  const preview = useCampaignPreview(id);
  const zones = useCampaignZones(id);

  if (isPending) return <LoadingState label="Loading campaign…" />;
  if (isError || !campaign) return <ErrorState title="Not found" message={(error as Error)?.message || 'Campaign not found'} onRetry={() => refetch()} />;

  const cap = campaign.dailyBudgetCap || campaign.totalBudgetCap || 0;
  const pct = cap > 0 ? Math.min(100, Math.round((campaign.budgetSpent / cap) * 100)) : 0;

  return (
    <div className="space-y-6">
      <Button asChild variant="ghost" size="sm" className="-ml-2 w-fit">
        <Link to={CAMPAIGNS_PATH}><ArrowLeft className="size-4" /> Back to campaigns</Link>
      </Button>

      <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="font-display text-xl font-bold">{campaign.name}</h2>
            <StatusPill status={campaign.isActive ? 'active' : 'inactive'} />
            <StatusPill status={campaign.moderationStatus} />
            <StatusPill status={campaign.validationStatus} />
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            {formatLabel(campaign.channelCode || 'channel')} · {formatLabel(campaign.targetIntent)}
          </p>
        </div>
      </div>

      {campaign.validationStatus === 'failed' && campaign.validationNotes && (
        <InlineError message={`Validation failed: ${campaign.validationNotes}`} />
      )}

      <Card>
        <CardContent className="p-5">
          <div className="mb-1.5 flex items-center justify-between text-sm">
            <span className="font-medium">Budget spent</span>
            <span className="tabular-nums">{formatEtb(campaign.budgetSpent)} <span className="text-muted-foreground">/ {formatEtb(cap)}</span></span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-muted"><div className="h-full rounded-full bg-primary" style={{ width: `${pct}%` }} /></div>
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle>Campaign details</CardTitle></CardHeader>
          <CardContent>
            <dl className="space-y-3 text-sm">
              <Row label="ID" value={campaign.id} mono />
              <Row label="Channel" value={formatLabel(campaign.channelCode)} />
              <Row label="Frequency cap" value={`${campaign.frequencyCapPerDay}/day`} />
              <Row label="Schedule" value={campaign.scheduledStartAt
                ? `${formatDate(campaign.scheduledStartAt)} → ${campaign.scheduledEndAt ? formatDate(campaign.scheduledEndAt) : 'open-ended'}`
                : 'Continuous'} />
              <Row label="Destination" value={campaign.destinationUrl ? <a href={campaign.destinationUrl} target="_blank" rel="noreferrer" className="break-all text-identity hover:underline">{campaign.destinationUrl}</a> : '—'} />
              <Row label="Created" value={formatDate(campaign.createdAt)} />
              <Row label="Daily budget" value={formatEtb(campaign.dailyBudgetCap)} />
              <Row label="Total budget" value={formatEtb(campaign.totalBudgetCap)} />
              {campaign.segmentId && <Row label="Segment ID" value={campaign.segmentId} mono />}
              {campaign.title && <Row label="Title" value={campaign.title} />}
              {campaign.bodyText && <Row label="Body" value={campaign.bodyText} />}
              {campaign.moderationNotes && <Row label="Moderation notes" value={campaign.moderationNotes} />}
            </dl>
          </CardContent>
        </Card>

        <div className="space-y-6">
          <div>
            <h3 className="mb-3 font-display text-base font-semibold">Creative preview</h3>
            <CampaignPreviewPanel preview={preview.data} loading={preview.isPending} error={preview.isError ? (preview.error as Error)?.message : undefined} />
          </div>

          <Card>
            <CardHeader><CardTitle>Store zones</CardTitle></CardHeader>
            <CardContent>
              {zones.isPending ? (
                <p className="text-sm text-muted-foreground">Loading zones…</p>
              ) : zones.isError ? (
                <InlineError message={(zones.error as Error)?.message ?? 'Could not load zones'} />
              ) : (zones.data ?? []).length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No store zones linked. This campaign is delivered by intent alone.
                </p>
              ) : (
                <ul className="space-y-2">
                  {(zones.data ?? []).map(zone => (
                    <li key={zone.id} className="flex items-center justify-between gap-3 rounded-lg border border-border px-3 py-2">
                      <span className="font-mono text-sm tabular-nums">
                        {zone.latitude.toFixed(5)}, {zone.longitude.toFixed(5)}
                        <span className="ml-2 font-sans text-xs text-muted-foreground">
                          {zone.radius_metres.toLocaleString()} m
                        </span>
                      </span>
                      <StatusPill status={zone.is_active ? 'active' : 'pending'} />
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function Row({ label, value, mono = false }: { label: string; value: ReactNode; mono?: boolean }) {
  return (
    <div className="flex justify-between gap-4">
      <dt className="shrink-0 text-muted-foreground">{label}</dt>
      <dd className={`text-right ${mono ? 'break-all font-mono text-xs' : ''}`}>{value}</dd>
    </div>
  );
}
