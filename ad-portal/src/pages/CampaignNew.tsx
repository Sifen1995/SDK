import { useEffect, useMemo, useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  Card, CardContent, Button, Input, Label, Badge,
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
  LoadingState, EmptyState, InlineError, cn,
} from '@skykin/ui';
import { ArrowLeft, Check, Lock } from 'lucide-react';
import { useSubscription } from '../context/SubscriptionContext';
import { useChannels, useSegments, useCreateCampaign, useZones, useLinkCampaignZones } from '../lib/queries';
import { formatEtb } from '../lib/campaignUtils';
import { googleDriveToDirectImageUrl } from '../lib/googleDrive';
import { TARGET_INTENTS, channelNeedsImage, channelNeedsRichCopy } from '../types';
import { CAMPAIGNS_PATH } from '../routes';

const STEPS = ['Audience', 'Setup', 'Creative', 'Budget'] as const;

export default function CampaignNew() {
  const navigate = useNavigate();
  const { subscribed, subscription, loading: subLoading } = useSubscription();
  const channelsQ = useChannels();
  const segmentsQ = useSegments();
  const create = useCreateCampaign();
  const zonesQ = useZones();
  const linkZones = useLinkCampaignZones();

  const zones = zonesQ.data ?? [];
  const channels = channelsQ.data ?? [];
  const segments = segmentsQ.data?.segments ?? [];
  const audiencemartEnabled = segmentsQ.data?.audiencemart_enabled ?? false;

  const [step, setStep] = useState(0);
  const [segmentId, setSegmentId] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [channelId, setChannelId] = useState('');
  const [targetIntent, setTargetIntent] = useState('general_interest');
  const [title, setTitle] = useState('');
  const [bodyText, setBodyText] = useState('');
  const [imageUrl, setImageUrl] = useState('');
  const [destinationUrl, setDestinationUrl] = useState('');
  const [dailyBudget, setDailyBudget] = useState('200');
  const [totalBudget, setTotalBudget] = useState('2000');
  const [frequencyCap, setFrequencyCap] = useState('3');
  const [scheduledStartAt, setScheduledStartAt] = useState('');
  const [scheduledEndAt, setScheduledEndAt] = useState('');
  const [zoneIds, setZoneIds] = useState<string[]>([]);
  const [stepError, setStepError] = useState('');

  useEffect(() => { if (channels.length && !channelId) setChannelId(channels[0].id); }, [channels, channelId]);

  const selectedChannel = channels.find(c => c.id === channelId);
  const selectedSegment = segments.find(s => s.id === segmentId) ?? null;
  const channelCode = selectedChannel?.code ?? '';
  const needsImage = channelNeedsImage(channelCode);
  const needsRichCopy = channelNeedsRichCopy(channelCode);

  const intentOptions = useMemo(() => {
    if (selectedSegment?.top_intent_signals?.length) {
      return selectedSegment.top_intent_signals.map(v => ({ value: v, label: v.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) }));
    }
    return TARGET_INTENTS.map(i => ({ value: i.value, label: i.label }));
  }, [selectedSegment]);

  useEffect(() => {
    if (intentOptions.length && !intentOptions.some(o => o.value === targetIntent)) setTargetIntent(intentOptions[0].value);
  }, [intentOptions, targetIntent]);

  function validateStep(current: number): string | null {
    if (current === 1) {
      if (name.trim().length < 3) return 'Campaign name must be at least 3 characters';
      if (!channelId) return 'Select a delivery channel';
    }
    if (current === 2) {
      if (!bodyText.trim()) return 'Body text is required';
      if (needsImage && !imageUrl.trim()) return 'Image URL is required for banner campaigns';
      if (needsRichCopy && !title.trim()) return 'Title is required for this channel';
      if (!destinationUrl.trim()) return 'Destination URL is required';
    }
    if (current === 3) {
      const daily = Number(dailyBudget), total = Number(totalBudget), freq = Number(frequencyCap);
      if (!daily || daily <= 0) return 'Daily budget must be greater than 0';
      if (!total || total <= 0) return 'Total budget must be greater than 0';
      if (total < daily) return 'Total budget must be at least the daily cap';
      if (!freq || freq < 1) return 'Frequency cap must be at least 1';
      const maxDaily = subscription?.plan.max_daily_budget_etb;
      if (maxDaily && daily > maxDaily) return `Daily budget exceeds your plan limit of ${formatEtb(maxDaily)}`;
      // Mirrors the server's `gtfield=ScheduledStartAt` binding so the error
      // shows here rather than as a 400 after submit.
      if (scheduledEndAt && !scheduledStartAt) return 'Set a start date before an end date';
      if (scheduledStartAt && scheduledEndAt && new Date(scheduledEndAt) <= new Date(scheduledStartAt)) {
        return 'End date must be after the start date';
      }
    }
    return null;
  }

  /// `datetime-local` yields a zone-less "YYYY-MM-DDTHH:mm"; the backend binds
  /// *time.Time, so send a real RFC3339 instant.
  function toRfc3339(local: string): string | undefined {
    if (!local) return undefined;
    const date = new Date(local);
    return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
  }

  function goNext() {
    const err = validateStep(step);
    if (err) return setStepError(err);
    setStepError('');
    setStep(s => Math.min(s + 1, STEPS.length - 1));
  }
  function goBack() { setStepError(''); setStep(s => Math.max(s - 1, 0)); }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const err = validateStep(3);
    if (err) return setStepError(err);
    setStepError('');
    create.mutate(
      {
        name: name.trim(), target_intent: targetIntent, channel_id: channelId, segment_id: segmentId,
        title: title.trim() || undefined, body_text: bodyText.trim(), image_url: imageUrl.trim() || undefined,
        destination_url: destinationUrl.trim(), canvas_json: {},
        daily_budget_cap: Number(dailyBudget), total_budget_cap: Number(totalBudget), frequency_cap_per_day: Number(frequencyCap),
        scheduled_start_at: toRfc3339(scheduledStartAt), scheduled_end_at: toRfc3339(scheduledEndAt),
      },
      {
        onSuccess: async c => {
          // Zones are linked in a second call — the create DTO has no zone
          // field. A link failure must not strand the user on the form: the
          // campaign already exists, so send them to it either way and let the
          // detail page show which zones are actually attached.
          if (zoneIds.length > 0) {
            try {
              await linkZones.mutateAsync({ campaignId: c.id, zoneIds });
            } catch {
              /* surfaced on the detail page's zone list */
            }
          }
          navigate(`/campaigns/${c.id}`);
        },
      },
    );
  }

  if (subLoading || channelsQ.isPending || segmentsQ.isPending) return <LoadingState label="Preparing campaign builder…" />;

  if (!subscribed) {
    return (
      <div className="mx-auto max-w-lg">
        <EmptyState icon={Lock} title="Subscription required" description="Subscribe to a plan before creating campaigns."
          action={<Button asChild><Link to="/subscription">View plans</Link></Button>} />
      </div>
    );
  }

  const error = stepError || (create.isError ? (create.error as Error).message : '');

  return (
    <div className="max-w-3xl space-y-6">
      <Button asChild variant="ghost" size="sm" className="-ml-2 w-fit"><Link to={CAMPAIGNS_PATH}><ArrowLeft className="size-4" /> Back to campaigns</Link></Button>
      <div>
        <h2 className="font-display text-xl font-bold">New campaign</h2>
        <p className="text-sm text-muted-foreground">Build your campaign in four steps. Segment purchase is included when you submit.</p>
      </div>

      <div className="grid grid-cols-4 gap-2">
        {STEPS.map((label, i) => (
          <div key={label} className={cn('flex flex-col items-center gap-1.5 rounded-lg border p-3 text-center transition-colors', i === step ? 'border-identity bg-identity/5' : i < step ? 'border-border' : 'border-border opacity-60')}>
            <span className={cn('flex size-7 items-center justify-center rounded-full text-xs font-bold', i <= step ? 'bg-identity text-identity-foreground' : 'bg-muted text-muted-foreground')}>
              {i < step ? <Check className="size-4" /> : i + 1}
            </span>
            <span className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{label}</span>
          </div>
        ))}
      </div>

      <form onSubmit={handleSubmit}>
        <Card>
          <CardContent className="space-y-6 p-6">
            {error && <InlineError message={error} />}

            {step === 0 && (
              <div className="space-y-5">
                <div>
                  <h3 className="font-display font-semibold">Audience targeting</h3>
                  <p className="mt-1 text-sm text-muted-foreground">Optionally attach an AudienceMart segment. The fee is charged when you create the campaign.</p>
                </div>
                {!audiencemartEnabled ? (
                  <div className="rounded-lg border border-border bg-muted/50 p-5">
                    <p className="font-medium">Intent-only targeting</p>
                    <p className="mt-1 text-sm text-muted-foreground">
                      Your {subscription?.plan.name ?? 'current'} plan targets by predicted intent without purchasable segments.
                      <Link to="/subscription" className="ml-1 text-identity hover:underline">Upgrade</Link> for AudienceMart.
                    </p>
                  </div>
                ) : (
                  <div className="grid gap-3">
                    <SegmentButton selected={segmentId === null} onClick={() => setSegmentId(null)}>
                      <p className="font-semibold">No segment</p>
                      <p className="mt-1 text-xs text-muted-foreground">Target by intent only — no AudienceMart purchase</p>
                    </SegmentButton>
                    {segments.map(seg => (
                      <SegmentButton key={seg.id} selected={segmentId === seg.id} onClick={() => setSegmentId(seg.id)}>
                        <div className="flex items-start justify-between gap-3">
                          <div className="min-w-0">
                            <p className="font-semibold">{seg.name}</p>
                            <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{seg.description}</p>
                          </div>
                          <span className="shrink-0 rounded-full bg-identity/10 px-2.5 py-0.5 text-sm font-bold tabular-nums text-identity">{formatEtb(seg.estimated_price_etb)}</span>
                        </div>
                        <div className="mt-3 flex flex-wrap gap-1.5">
                          <Badge variant="secondary">~{seg.approximate_size.toLocaleString()} users</Badge>
                          {seg.top_intent_signals.map(sig => <Badge key={sig} variant="outline">{sig.replace(/_/g, ' ')}</Badge>)}
                        </div>
                      </SegmentButton>
                    ))}
                    {segments.length === 0 && <p className="text-sm text-muted-foreground">No segments available for purchase right now.</p>}
                  </div>
                )}

                <div className="border-t border-border pt-5">
                  <h3 className="font-display font-semibold">Store zones (optional)</h3>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Attach store locations so customers who walk within range receive this
                    campaign. Draft zones go live automatically when an operator approves it.
                  </p>
                  {zones.length === 0 ? (
                    <div className="mt-3 rounded-lg border border-border bg-muted/50 p-4 text-sm text-muted-foreground">
                      You have no store zones yet.
                      <Link to="/zones" className="ml-1 text-identity hover:underline">Create one</Link> to
                      target customers near a location.
                    </div>
                  ) : (
                    <div className="mt-3 grid gap-2">
                      {zones.map(zone => {
                        const selected = zoneIds.includes(zone.id);
                        return (
                          <button
                            key={zone.id}
                            type="button"
                            onClick={() => setZoneIds(ids =>
                              selected ? ids.filter(id => id !== zone.id) : [...ids, zone.id])}
                            className={cn(
                              'flex items-center justify-between gap-3 rounded-lg border p-3 text-left transition-colors',
                              selected ? 'border-identity bg-identity/5 ring-1 ring-identity' : 'border-border hover:bg-muted/50',
                            )}
                          >
                            <span className="min-w-0">
                              <span className="block font-mono text-sm tabular-nums">
                                {zone.latitude.toFixed(5)}, {zone.longitude.toFixed(5)}
                              </span>
                              <span className="mt-0.5 block text-xs text-muted-foreground">
                                {zone.radius_metres.toLocaleString()} m radius
                              </span>
                            </span>
                            <Badge variant={zone.is_active ? 'secondary' : 'outline'}>
                              {zone.is_active ? 'Active' : 'Draft'}
                            </Badge>
                          </button>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            )}

            {step === 1 && (
              <div className="space-y-5">
                <div className="space-y-1.5"><Label htmlFor="cname">Campaign name</Label><Input id="cname" value={name} onChange={e => setName(e.target.value)} placeholder="Fashion Spring Sale" /></div>
                <div>
                  <Label className="mb-2 block">Delivery channel</Label>
                  <div className="grid gap-3 sm:grid-cols-2">
                    {channels.map(ch => (
                      <SegmentButton key={ch.id} selected={channelId === ch.id} onClick={() => setChannelId(ch.id)}>
                        <div className="flex items-center justify-between gap-2">
                          <p className="text-sm font-semibold">{ch.name}</p>
                          {ch.is_premium && <Badge variant="warning">Premium</Badge>}
                        </div>
                        <p className="mt-1 text-xs text-muted-foreground">{ch.description}</p>
                      </SegmentButton>
                    ))}
                  </div>
                </div>
                <div className="space-y-1.5">
                  <Label>Target intent</Label>
                  <Select value={targetIntent} onValueChange={setTargetIntent}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>{intentOptions.map(i => <SelectItem key={i.value} value={i.value}>{i.label}</SelectItem>)}</SelectContent>
                  </Select>
                  <p className="text-xs text-muted-foreground">All campaigns are billed per click (CPC).</p>
                </div>
              </div>
            )}

            {step === 2 && (
              <div className="space-y-5">
                <p className="text-sm text-muted-foreground">Creative for <strong className="text-foreground">{selectedChannel?.name}</strong>{selectedChannel?.is_premium && ' (premium channel)'}</p>
                {(needsRichCopy || !needsImage) && (
                  <div className="space-y-1.5">
                    <Label htmlFor="title">Title {needsRichCopy && <span className="text-identity">*</span>}</Label>
                    <Input id="title" value={title} onChange={e => setTitle(e.target.value)} maxLength={channelCode === 'SMS_PLUS' ? 40 : 50} placeholder="Headline" />
                  </div>
                )}
                <div className="space-y-1.5">
                  <Label htmlFor="body">Body text <span className="text-identity">*</span></Label>
                  <textarea id="body" required value={bodyText} onChange={e => setBodyText(e.target.value)} maxLength={channelCode === 'SMS_PLUS' ? 160 : 120} rows={3}
                    className="flex w-full resize-y rounded-md border border-input bg-card px-3 py-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" placeholder="Your ad copy…" />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="image">{needsImage ? <>Image URL <span className="text-identity">*</span></> : 'Image URL (optional)'}</Label>
                  <Input id="image" value={imageUrl} onChange={e => setImageUrl(e.target.value)}
                    onBlur={e => { const d = googleDriveToDirectImageUrl(e.target.value); if (d) setImageUrl(d); }}
                    placeholder="https://… or a Google Drive share link" />
                  {imageUrl.trim() && <img src={imageUrl} alt="" className="mt-2 max-h-40 rounded-lg border border-border object-contain" onError={e => ((e.target as HTMLImageElement).style.display = 'none')} />}
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="dest">Destination URL <span className="text-identity">*</span></Label>
                  <Input id="dest" type="url" required value={destinationUrl} onChange={e => setDestinationUrl(e.target.value)} placeholder="https://example.com/sale" />
                </div>
              </div>
            )}

            {step === 3 && (
              <div className="space-y-5">
                <div className="rounded-lg border border-border bg-muted/50 p-4 text-sm">
                  <p className="mb-2 font-medium">Review summary</p>
                  <ul className="space-y-1 text-muted-foreground">
                    <li><span className="text-foreground">{name}</span> · {selectedChannel?.name}</li>
                    <li>Intent: {targetIntent.replace(/_/g, ' ')}</li>
                    <li>Segment: {selectedSegment ? `${selectedSegment.name} (${formatEtb(selectedSegment.estimated_price_etb)})` : 'None'}</li>
                    <li>Store zones: {zoneIds.length === 0 ? 'None' : `${zoneIds.length} linked`}</li>
                  </ul>
                </div>
                <div className="grid gap-4 sm:grid-cols-3">
                  <div className="space-y-1.5"><Label>Daily budget (ETB)</Label><Input type="number" min="1" required value={dailyBudget} onChange={e => setDailyBudget(e.target.value)} /></div>
                  <div className="space-y-1.5"><Label>Total budget (ETB)</Label><Input type="number" min="1" required value={totalBudget} onChange={e => setTotalBudget(e.target.value)} /></div>
                  <div className="space-y-1.5"><Label>Freq. cap / day</Label><Input type="number" min="1" required value={frequencyCap} onChange={e => setFrequencyCap(e.target.value)} /></div>
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-1.5">
                    <Label htmlFor="start-at">Start (optional)</Label>
                    <Input id="start-at" type="datetime-local" value={scheduledStartAt} onChange={e => setScheduledStartAt(e.target.value)} />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="end-at">End (optional)</Label>
                    <Input id="end-at" type="datetime-local" value={scheduledEndAt} onChange={e => setScheduledEndAt(e.target.value)} />
                  </div>
                </div>
                <p className="-mt-2 text-xs text-muted-foreground">Leave both empty to run continuously once approved.</p>
                <div className="rounded-md border border-identity/30 bg-identity/5 px-3 py-2 text-sm text-foreground">
                  After submission your campaign enters <strong>pending moderation</strong>. An operator reviews and activates it — you cannot self-activate.
                </div>
              </div>
            )}

            <div className="flex gap-3 border-t border-border pt-4">
              {step > 0 && <Button type="button" variant="secondary" onClick={goBack}>Back</Button>}
              {step < STEPS.length - 1 ? (
                <Button type="button" onClick={goNext}>Continue</Button>
              ) : (
                <Button type="submit" disabled={create.isPending}>{create.isPending ? 'Creating…' : 'Submit for review'}</Button>
              )}
              <Button asChild variant="ghost" className="ml-auto"><Link to={CAMPAIGNS_PATH}>Cancel</Link></Button>
            </div>
          </CardContent>
        </Card>
      </form>
    </div>
  );
}

function SegmentButton({ selected, onClick, children }: { selected: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button type="button" onClick={onClick} className={cn('rounded-lg border p-4 text-left transition-colors', selected ? 'border-identity bg-identity/5 ring-1 ring-identity' : 'border-border hover:bg-muted/50')}>
      {children}
    </button>
  );
}
