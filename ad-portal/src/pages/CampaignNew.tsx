import { useEffect, useMemo, useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import { useSubscription } from '../context/SubscriptionContext';
import ImageUrlInput from '../components/ImageUrlInput';
import { formatEtb } from '../lib/campaignUtils';
import {
  BILLING_MODELS,
  TARGET_INTENTS,
  channelNeedsImage,
  channelNeedsRichCopy,
  type AudienceSegment,
  type DeliveryChannel,
} from '../types';

const STEPS = ['Audience', 'Setup', 'Creative', 'Budget'] as const;

export default function CampaignNew() {
  const navigate = useNavigate();
  const { subscribed, subscription, loading: subLoading } = useSubscription();

  const [step, setStep] = useState(0);
  const [channels, setChannels] = useState<DeliveryChannel[]>([]);
  const [segments, setSegments] = useState<AudienceSegment[]>([]);
  const [audiencemartEnabled, setAudiencemartEnabled] = useState(false);
  const [catalogLoading, setCatalogLoading] = useState(true);

  const [segmentId, setSegmentId] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [channelId, setChannelId] = useState('');
  const [targetIntent, setTargetIntent] = useState('general_interest');
  const [billingModel, setBillingModel] = useState('CPC');
  const [title, setTitle] = useState('');
  const [bodyText, setBodyText] = useState('');
  const [imageUrl, setImageUrl] = useState('');
  const [destinationUrl, setDestinationUrl] = useState('');
  const [dailyBudget, setDailyBudget] = useState('200');
  const [totalBudget, setTotalBudget] = useState('2000');
  const [frequencyCap, setFrequencyCap] = useState('3');

  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const selectedChannel = channels.find(c => c.id === channelId);
  const selectedSegment = segments.find(s => s.id === segmentId) ?? null;
  const channelCode = selectedChannel?.code ?? '';

  const intentOptions = useMemo(() => {
    if (selectedSegment?.top_intent_signals?.length) {
      return selectedSegment.top_intent_signals.map(value => ({
        value,
        label: value.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()),
      }));
    }
    return TARGET_INTENTS.map(i => ({ value: i.value, label: i.label }));
  }, [selectedSegment]);

  useEffect(() => {
    if (intentOptions.length && !intentOptions.some(o => o.value === targetIntent)) {
      setTargetIntent(intentOptions[0].value);
    }
  }, [intentOptions, targetIntent]);

  useEffect(() => {
    if (!subscribed) return;
    let cancelled = false;
    (async () => {
      try {
        const [ch, seg] = await Promise.all([api.listChannels(), api.listSegments()]);
        if (cancelled) return;
        setChannels(ch);
        setSegments(seg.segments ?? []);
        setAudiencemartEnabled(seg.audiencemart_enabled);
        if (ch.length > 0) setChannelId(ch[0].id);
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Failed to load catalog');
      } finally {
        if (!cancelled) setCatalogLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [subscribed]);

  function validateStep(current: number): string | null {
    if (current === 1) {
      if (!name.trim() || name.trim().length < 3) return 'Campaign name must be at least 3 characters';
      if (!channelId) return 'Select a delivery channel';
    }
    if (current === 2) {
      if (!bodyText.trim()) return 'Body text is required';
      if (channelNeedsImage(channelCode) && !imageUrl.trim()) return 'Image URL is required for banner campaigns';
      if (channelNeedsRichCopy(channelCode) && !title.trim()) return 'Title is required for this channel';
      if (!destinationUrl.trim()) return 'Destination URL is required';
    }
    if (current === 3) {
      const daily = Number(dailyBudget);
      const total = Number(totalBudget);
      const freq = Number(frequencyCap);
      if (!daily || daily <= 0) return 'Daily budget must be greater than 0';
      if (!total || total <= 0) return 'Total budget must be greater than 0';
      if (total < daily) return 'Total budget must be at least the daily cap';
      if (!freq || freq < 1) return 'Frequency cap must be at least 1';
      const maxDaily = subscription?.plan.max_daily_budget_etb;
      if (maxDaily && daily > maxDaily) {
        return `Daily budget exceeds your plan limit of ${formatEtb(maxDaily)}`;
      }
    }
    return null;
  }

  function goNext() {
    const err = validateStep(step);
    if (err) {
      setError(err);
      return;
    }
    setError('');
    setStep(s => Math.min(s + 1, STEPS.length - 1));
  }

  function goBack() {
    setError('');
    setStep(s => Math.max(s - 1, 0));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const err = validateStep(3);
    if (err) {
      setError(err);
      return;
    }
    setError('');
    setLoading(true);
    try {
      const campaign = await api.createCampaign({
        name: name.trim(),
        target_intent: targetIntent,
        channel_id: channelId,
        segment_id: segmentId,
        title: title.trim() || undefined,
        body_text: bodyText.trim(),
        image_url: imageUrl.trim() || undefined,
        destination_url: destinationUrl.trim(),
        canvas_json: {},
        billing_model: billingModel,
        daily_budget_cap: Number(dailyBudget),
        total_budget_cap: Number(totalBudget),
        frequency_cap_per_day: Number(frequencyCap),
      });
      navigate(`/campaigns/${campaign.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create campaign');
    } finally {
      setLoading(false);
    }
  }

  if (subLoading || catalogLoading) {
    return <p className="text-muted">Preparing campaign builder…</p>;
  }

  if (!subscribed) {
    return (
      <div className="max-w-lg mx-auto text-center card p-10">
        <div className="text-4xl mb-4">🔒</div>
        <h1 className="text-xl font-bold text-primary">Subscription required</h1>
        <p className="text-muted mt-2">Subscribe to a plan before creating campaigns.</p>
        <Link to="/subscription" className="btn-primary mt-6 inline-flex">View plans</Link>
      </div>
    );
  }

  const needsImage = channelNeedsImage(channelCode);
  const needsRichCopy = channelNeedsRichCopy(channelCode);

  return (
    <div className="max-w-3xl">
      <Link to="/" className="text-sm text-brand-600 hover:text-brand-700 font-medium">
        ← Back to campaigns
      </Link>
      <h1 className="text-2xl font-bold text-primary mt-4">New campaign</h1>
      <p className="text-muted mt-1">
        Build your campaign in four steps. Segment purchase is included when you submit.
      </p>

      <div className="wizard-steps mt-8 mb-8">
        {STEPS.map((label, i) => (
          <div key={label} className={`wizard-step ${i === step ? 'wizard-step-active' : ''} ${i < step ? 'wizard-step-done' : ''}`}>
            <span className="wizard-step-num">{i < step ? '✓' : i + 1}</span>
            <span className="wizard-step-label">{label}</span>
          </div>
        ))}
      </div>

      <form onSubmit={handleSubmit} className="card p-6 sm:p-8 space-y-6">
        {error && (
          <div className="rounded-lg border border-red-200 dark:border-red-900/50 bg-red-50 dark:bg-red-950/30 text-red-700 dark:text-red-400 px-4 py-3 text-sm">
            {error}
          </div>
        )}

        {step === 0 && (
          <div className="space-y-6">
            <div>
              <h2 className="font-semibold text-primary">Audience targeting</h2>
              <p className="text-sm text-muted mt-1">
                Optionally attach an Audiencemart segment. The segment fee is charged when you create the campaign.
              </p>
            </div>

            {!audiencemartEnabled ? (
              <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-subtle)] p-5">
                <p className="font-medium text-primary">Intent-only targeting</p>
                <p className="text-sm text-muted mt-1">
                  Your {subscription?.plan.name ?? 'current'} plan targets users by predicted intent without purchasable segments.
                  <Link to="/subscription" className="text-brand-600 hover:underline ml-1">Upgrade</Link> for Audiencemart.
                </p>
              </div>
            ) : (
              <div className="grid gap-3">
                <button
                  type="button"
                  onClick={() => setSegmentId(null)}
                  className={`segment-card text-left ${segmentId === null ? 'segment-card-selected' : ''}`}
                >
                  <p className="font-semibold text-primary">No segment</p>
                  <p className="text-xs text-muted mt-1">Target by intent only — no Audiencemart purchase</p>
                </button>

                {segments.map(seg => (
                  <button
                    key={seg.id}
                    type="button"
                    onClick={() => setSegmentId(seg.id)}
                    className={`segment-card text-left ${segmentId === seg.id ? 'segment-card-selected' : ''}`}
                  >
                    <div className="flex justify-between items-start gap-3">
                      <div>
                        <p className="font-semibold text-primary">{seg.name}</p>
                        <p className="text-xs text-muted mt-1 line-clamp-2">{seg.description}</p>
                      </div>
                      <span className="segment-price shrink-0">{formatEtb(seg.estimated_price_etb)}</span>
                    </div>
                    <div className="flex flex-wrap gap-2 mt-3">
                      <span className="segment-meta">~{seg.approximate_size.toLocaleString()} users</span>
                      {seg.top_intent_signals.map(sig => (
                        <span key={sig} className="segment-meta">{sig.replace(/_/g, ' ')}</span>
                      ))}
                    </div>
                  </button>
                ))}

                {segments.length === 0 && (
                  <p className="text-sm text-muted">No segments available for purchase right now.</p>
                )}
              </div>
            )}
          </div>
        )}

        {step === 1 && (
          <div className="space-y-5">
            <div>
              <label className="block text-sm font-medium text-primary mb-1.5">Campaign name</label>
              <input required value={name} onChange={e => setName(e.target.value)} className="field-input" placeholder="Fashion Spring Sale" />
            </div>

            <div>
              <label className="block text-sm font-medium text-primary mb-2">Delivery channel</label>
              <div className="grid sm:grid-cols-2 gap-3">
                {channels.map(ch => (
                  <button
                    key={ch.id}
                    type="button"
                    onClick={() => setChannelId(ch.id)}
                    className={`channel-card text-left ${channelId === ch.id ? 'channel-card-selected' : ''}`}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <p className="font-semibold text-sm text-primary">{ch.name}</p>
                      {ch.is_premium && <span className="premium-badge">Premium</span>}
                    </div>
                    <p className="text-xs text-muted mt-1">{ch.description}</p>
                  </button>
                ))}
              </div>
            </div>

            <div className="grid sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-primary mb-1.5">Target intent</label>
                <select value={targetIntent} onChange={e => setTargetIntent(e.target.value)} className="field-input">
                  {intentOptions.map(i => (
                    <option key={i.value} value={i.value}>{i.label}</option>
                  ))}
                </select>
                {selectedSegment && (
                  <p className="text-xs text-faint mt-1">Must match your selected segment&apos;s intent signals.</p>
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1.5">Billing model</label>
                <select value={billingModel} onChange={e => setBillingModel(e.target.value)} className="field-input">
                  {BILLING_MODELS.map(m => (
                    <option key={m.value} value={m.value}>{m.label} — {m.description}</option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        )}

        {step === 2 && (
          <div className="space-y-5">
            <p className="text-sm text-muted">
              Creative for <strong className="text-primary">{selectedChannel?.name}</strong>
              {selectedChannel?.is_premium && ' (premium channel)'}
            </p>

            {(needsRichCopy || !needsImage) && (
              <div>
                <label className="block text-sm font-medium text-primary mb-1.5">
                  Title {needsRichCopy && <span className="text-brand-600">*</span>}
                </label>
                <input
                  value={title}
                  onChange={e => setTitle(e.target.value)}
                  maxLength={channelCode === 'SMS_PLUS' ? 40 : 50}
                  className="field-input"
                  placeholder="Headline"
                />
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-primary mb-1.5">
                Body text <span className="text-brand-600">*</span>
              </label>
              <textarea
                required
                value={bodyText}
                onChange={e => setBodyText(e.target.value)}
                maxLength={channelCode === 'SMS_PLUS' ? 160 : 120}
                rows={3}
                className="field-input resize-y"
                placeholder="Your ad copy…"
              />
            </div>

            {needsImage ? (
              <ImageUrlInput value={imageUrl} onChange={setImageUrl} required />
            ) : (
              <ImageUrlInput value={imageUrl} onChange={setImageUrl} label="Optional image" />
            )}

            <div>
              <label className="block text-sm font-medium text-primary mb-1.5">
                Destination URL <span className="text-brand-600">*</span>
              </label>
              <input
                type="url"
                required
                value={destinationUrl}
                onChange={e => setDestinationUrl(e.target.value)}
                className="field-input"
                placeholder="https://example.com/sale"
              />
            </div>
          </div>
        )}

        {step === 3 && (
          <div className="space-y-5">
            <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-subtle)] p-4 text-sm">
              <p className="font-medium text-primary mb-2">Review summary</p>
              <ul className="space-y-1 text-muted">
                <li><span className="text-primary">{name}</span> · {selectedChannel?.name} · {billingModel}</li>
                <li>Intent: {targetIntent.replace(/_/g, ' ')}</li>
                <li>Segment: {selectedSegment ? `${selectedSegment.name} (${formatEtb(selectedSegment.estimated_price_etb)})` : 'None'}</li>
              </ul>
            </div>

            <div className="grid sm:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium text-primary mb-1.5">Daily budget (ETB)</label>
                <input type="number" min="1" step="1" required value={dailyBudget} onChange={e => setDailyBudget(e.target.value)} className="field-input" />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1.5">Total budget (ETB)</label>
                <input type="number" min="1" step="1" required value={totalBudget} onChange={e => setTotalBudget(e.target.value)} className="field-input" />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1.5">Freq. cap / day</label>
                <input type="number" min="1" step="1" required value={frequencyCap} onChange={e => setFrequencyCap(e.target.value)} className="field-input" />
              </div>
            </div>

            <div className="alert-info text-sm">
              After submission your campaign enters <strong>pending moderation</strong>. An operator will review and activate it — you cannot self-activate.
            </div>
          </div>
        )}

        <div className="flex gap-3 pt-2 border-t border-[var(--border)]">
          {step > 0 && (
            <button type="button" onClick={goBack} className="btn-secondary">
              Back
            </button>
          )}
          {step < STEPS.length - 1 ? (
            <button type="button" onClick={goNext} className="btn-primary">
              Continue
            </button>
          ) : (
            <button type="submit" disabled={loading} className="btn-primary">
              {loading ? 'Creating…' : 'Submit for review'}
            </button>
          )}
          <Link to="/" className="btn-ghost ml-auto">Cancel</Link>
        </div>
      </form>
    </div>
  );
}
