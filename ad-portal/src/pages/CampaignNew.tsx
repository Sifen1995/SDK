import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import ImageUrlInput from '../components/ImageUrlInput';
import { CREATIVE_FORMATS, TARGET_INTENTS } from '../types';

export default function CampaignNew() {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [targetIntent, setTargetIntent] = useState('crypto_interest');
  const [creativeFormat, setCreativeFormat] = useState('BANNER');
  const [title, setTitle] = useState('');
  const [bodyText, setBodyText] = useState('');
  const [imageUrl, setImageUrl] = useState('');
  const [destinationUrl, setDestinationUrl] = useState('');
  const [dailyBudget, setDailyBudget] = useState('');
  const [totalBudget, setTotalBudget] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const needsImage = creativeFormat === 'BANNER';
  const needsCopy = creativeFormat === 'PUSH_PLUS' || creativeFormat === 'SMS_PLUS';

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');

    if (needsImage && !imageUrl.trim()) {
      setError('Image URL is required for banner campaigns');
      return;
    }
    if (!destinationUrl.trim()) {
      setError('Destination URL is required for all campaigns');
      return;
    }
    if (needsCopy && !title.trim()) {
      setError('Title is required for this format');
      return;
    }
    if (needsCopy && !bodyText.trim()) {
      setError('Body text is required for this format');
      return;
    }

    setLoading(true);
    try {
      const campaign = await api.createCampaign({
        name,
        target_intent: targetIntent,
        creative_format: creativeFormat,
        title: title || undefined,
        body_text: bodyText || undefined,
        image_url: imageUrl || undefined,
        destination_url: destinationUrl,
        canvas_json: {},
        daily_budget_cap: dailyBudget ? Number(dailyBudget) : undefined,
        total_budget_cap: totalBudget ? Number(totalBudget) : undefined,
      });
      navigate(`/campaigns/${campaign.id}`, { state: { highlightActivate: true } });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create campaign');
    } finally {
      setLoading(false);
    }
  }

  const selectedFormat = CREATIVE_FORMATS.find(f => f.value === creativeFormat);

  return (
    <div className="max-w-3xl">
      <Link to="/" className="text-sm text-brand-600 hover:text-brand-700 font-medium">
        ← Back to campaigns
      </Link>
      <h1 className="text-2xl font-bold text-primary mt-4">New campaign</h1>
      <p className="text-muted mt-1">Creative fields are embedded in one campaign row per backend schema.</p>

      <form onSubmit={handleSubmit} className="card mt-8 p-6 sm:p-8 space-y-6">
        {error && (
          <div className="rounded-lg border border-red-200 dark:border-red-900/50 bg-red-50 dark:bg-red-950/30 text-red-700 dark:text-red-400 px-4 py-3 text-sm">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-primary mb-1.5">Campaign name</label>
          <input required value={name} onChange={e => setName(e.target.value)} className="field-input" placeholder="Crypto Promo Q2" />
        </div>

        <div className="grid sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-primary mb-1.5">Target intent</label>
            <select value={targetIntent} onChange={e => setTargetIntent(e.target.value)} className="field-input">
              {TARGET_INTENTS.map(i => (
                <option key={i.value} value={i.value}>{i.label}</option>
              ))}
            </select>
            <p className="text-xs text-faint mt-1">Must match ML intent labels for ad delivery.</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-primary mb-1.5">Daily budget cap</label>
            <input type="number" min="0" step="0.01" value={dailyBudget} onChange={e => setDailyBudget(e.target.value)} className="field-input" placeholder="Optional" />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-primary mb-2">Creative format</label>
          <div className="grid sm:grid-cols-3 gap-3">
            {CREATIVE_FORMATS.map(f => (
              <button
                key={f.value}
                type="button"
                onClick={() => setCreativeFormat(f.value)}
                className={`rounded-xl border p-4 text-left transition cursor-pointer ${
                  creativeFormat === f.value
                    ? 'border-brand-500 bg-brand-50 dark:bg-brand-950/20 ring-2 ring-brand-200 dark:ring-brand-500'
                    : 'border-[var(--border)] hover:border-brand-200'
                }`}
              >
                <p className="font-semibold text-sm text-primary">{f.label}</p>
                <p className="text-xs text-muted mt-1">{f.description}</p>
              </button>
            ))}
          </div>
        </div>

        {(needsCopy || creativeFormat === 'BANNER') && (
          <div className="grid sm:grid-cols-2 gap-4">
            <div className={needsCopy ? '' : 'sm:col-span-2'}>
              <label className="block text-sm font-medium text-primary mb-1.5">
                Title {needsCopy && <span className="text-brand-600">*</span>}
              </label>
              <input
                value={title}
                onChange={e => setTitle(e.target.value)}
                maxLength={creativeFormat === 'SMS_PLUS' ? 40 : 50}
                className="field-input"
                placeholder="Headline"
              />
            </div>
            <div className="sm:col-span-2">
              <label className="block text-sm font-medium text-primary mb-1.5">
                Body text {needsCopy && <span className="text-brand-600">*</span>}
              </label>
              <textarea
                value={bodyText}
                onChange={e => setBodyText(e.target.value)}
                maxLength={creativeFormat === 'SMS_PLUS' ? 160 : 120}
                rows={3}
                className="field-input resize-y"
                placeholder={selectedFormat?.description}
              />
              <p className="text-xs text-faint mt-1">
                {creativeFormat === 'PUSH_PLUS' && '1–120 characters required for validation.'}
                {creativeFormat === 'SMS_PLUS' && '1–160 characters required for validation.'}
              </p>
            </div>
          </div>
        )}

        {needsImage && (
          <ImageUrlInput value={imageUrl} onChange={setImageUrl} required />
        )}

        {!needsImage && (
          <div>
            <ImageUrlInput value={imageUrl} onChange={setImageUrl} label="Optional image" />
          </div>
        )}

        <div className="sm:col-span-2">
          <label className="block text-sm font-medium text-primary mb-1.5">
            Destination URL <span className="text-brand-600">*</span>
          </label>
          <input 
            type="url" 
            required 
            value={destinationUrl} 
            onChange={e => setDestinationUrl(e.target.value)} 
            className="field-input" 
            placeholder="https://example.com/promo" 
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-primary mb-1.5">Total budget cap</label>
          <input type="number" min="0" step="0.01" value={totalBudget} onChange={e => setTotalBudget(e.target.value)} className="field-input max-w-xs" placeholder="Optional" />
        </div>

        <div className="flex gap-3 pt-2">
          <button type="submit" disabled={loading} className="btn-primary">
            {loading ? 'Creating…' : 'Create campaign'}
          </button>
          <Link to="/" className="btn-secondary">Cancel</Link>
        </div>
      </form>
    </div>
  );
}
