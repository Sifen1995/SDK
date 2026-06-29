import { useState } from 'react';
import type { Campaign } from '../types';
import { formatLabel } from '../lib/campaignUtils';

interface ActivateCampaignPanelProps {
  campaign: Campaign;
  canWrite: boolean;
  onActivate: () => Promise<void>;
  highlight?: boolean;
}

export default function ActivateCampaignPanel({
  campaign,
  canWrite,
  onActivate,
  highlight = false,
}: ActivateCampaignPanelProps) {
  const [loading, setLoading] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  const isActive = campaign.isActive;
  const validation = campaign.validationStatus;
  const canActivate = canWrite && !isActive && validation === 'passed';

  async function handleConfirm() {
    setLoading(true);
    setError('');
    try {
      await onActivate();
      setConfirmOpen(false);
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Activation failed');
    } finally {
      setLoading(false);
    }
  }

  const steps = [
    { label: 'Created', done: true },
    { label: 'Validated', done: validation === 'passed' || validation === 'warning' },
    { label: 'Live', done: isActive },
  ];

  return (
    <section
      className={`activate-panel ${highlight ? 'activate-panel-highlight' : ''} ${isActive ? 'activate-panel-live' : ''}`}
    >
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wider text-muted mb-1">Campaign lifecycle</p>
          <h2 className="text-lg font-semibold text-primary">
            {isActive ? 'Campaign is live' : 'Ready to go live?'}
          </h2>
          <p className="text-sm text-muted mt-1 max-w-xl">
            {isActive
              ? 'This campaign is active and eligible for intent-matched ad delivery to connected SDK users.'
              : 'Activation publishes your creative to the delivery pipeline. Ads are pushed when user intent matches your target.'}
          </p>
        </div>

        {canWrite && !isActive && (
          <button
            type="button"
            onClick={() => setConfirmOpen(true)}
            disabled={!canActivate || loading}
            className="btn-primary shrink-0 min-w-[160px]"
          >
            {loading ? 'Activating…' : 'Activate campaign'}
          </button>
        )}
      </div>

      <div className="mt-6 flex items-center gap-2 sm:gap-4">
        {steps.map((step, i) => (
          <div key={step.label} className="flex items-center gap-2 flex-1 min-w-0">
            <div className={`activate-step ${step.done ? 'activate-step-done' : ''}`}>
              {step.done ? '✓' : i + 1}
            </div>
            <span className={`text-xs sm:text-sm truncate ${step.done ? 'text-primary font-medium' : 'text-muted'}`}>
              {step.label}
            </span>
            {i < steps.length - 1 && <div className="activate-step-line hidden sm:block" />}
          </div>
        ))}
      </div>

      {success && isActive && (
        <div className="alert-success mt-4">
          Campaign activated successfully. Matching SDK sessions may now receive this creative.
        </div>
      )}

      {error && (
        <div className="alert-error mt-4">{error}</div>
      )}

      {!canWrite && !isActive && (
        <p className="text-sm text-muted mt-4">Read-only access — you cannot activate campaigns.</p>
      )}

      {canWrite && !isActive && validation !== 'passed' && (
        <div className="alert-warning mt-4">
          {validation === 'failed' && (
            <>
              <strong>Cannot activate:</strong> validation failed.
              {campaign.validationNotes && ` ${campaign.validationNotes}`}
            </>
          )}
          {validation === 'warning' && (
            <>
              <strong>Review required:</strong> {campaign.validationNotes || 'Validation warning — fix issues or wait for review.'}
            </>
          )}
          {validation === 'pending' && 'Validation is still pending. Try again shortly.'}
          {validation !== 'failed' && validation !== 'warning' && validation !== 'pending' && (
            <>Activation requires validation status <code className="text-xs">passed</code> (current: {validation}).</>
          )}
        </div>
      )}

      {confirmOpen && (
        <div className="modal-backdrop" onClick={() => !loading && setConfirmOpen(false)}>
          <div className="modal-card" onClick={e => e.stopPropagation()} role="dialog" aria-modal="true">
            <h3 className="text-lg font-semibold text-primary">Activate campaign?</h3>
            <p className="text-sm text-muted mt-2">
              <strong className="text-primary">{campaign.name}</strong> will go live for{' '}
              <span className="font-medium">{formatLabel(campaign.targetIntent)}</span> on app{' '}
            </p>
            <ul className="text-sm text-muted mt-4 space-y-1 list-disc list-inside">
              <li>Inactive campaigns will not be delivered to the SDK</li>
              <li>You can create a new version if you need to change creative fields</li>
            </ul>
            <div className="flex gap-3 mt-6 justify-end">
              <button type="button" className="btn-secondary" disabled={loading} onClick={() => setConfirmOpen(false)}>
                Cancel
              </button>
              <button type="button" className="btn-primary" disabled={loading} onClick={handleConfirm}>
                {loading ? 'Activating…' : 'Confirm activation'}
              </button>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}
