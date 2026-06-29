import type { Campaign } from '../types';
import { formatLabel } from '../lib/campaignUtils';

interface CampaignStatusPanelProps {
  campaign: Campaign;
}

export default function CampaignStatusPanel({ campaign }: CampaignStatusPanelProps) {
  const validation = campaign.validationStatus;
  const moderation = campaign.moderationStatus;
  const isLive = campaign.isActive;

  const steps = [
    { label: 'Submitted', done: true },
    { label: 'Creative check', done: validation === 'passed' || validation === 'warning' },
    { label: 'Moderation', done: moderation === 'approved' },
    { label: 'Live', done: isLive },
  ];

  let headline = 'Awaiting review';
  let description =
    'Your campaign has been submitted. Our team will review your creative and activate it when approved.';

  if (moderation === 'rejected') {
    headline = 'Campaign rejected';
    description =
      campaign.moderationNotes ||
      'This campaign was rejected during moderation. Create a new campaign with updated creative.';
  } else if (validation === 'failed') {
    headline = 'Creative validation failed';
    description =
      campaign.validationNotes ||
      'Automatic creative checks failed. Update your assets and submit a new campaign.';
  } else if (isLive) {
    headline = 'Campaign is live';
    description =
      'Your campaign is active and eligible for intent-matched delivery to connected SDK users.';
  } else if (moderation === 'approved' && validation === 'passed') {
    headline = 'Approved — pending activation';
    description =
      'Moderation and creative checks passed. An operator will activate your campaign to go live.';
  } else if (moderation === 'pending') {
    headline = 'Pending moderation';
    description = `Targeting ${formatLabel(campaign.targetIntent)}. You'll be notified once an operator reviews this campaign.`;
  }

  return (
    <section className={`campaign-status-panel ${isLive ? 'campaign-status-live' : ''}`}>
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wider text-muted mb-1">Campaign lifecycle</p>
          <h2 className="text-lg font-semibold text-primary">{headline}</h2>
          <p className="text-sm text-muted mt-1 max-w-xl">{description}</p>
        </div>
        {isLive && (
          <span className="live-pulse shrink-0">● Live</span>
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

      {moderation === 'rejected' && campaign.moderationNotes && (
        <div className="alert-error mt-4">
          <strong>Moderator notes:</strong> {campaign.moderationNotes}
        </div>
      )}

      {!isLive && moderation !== 'rejected' && validation !== 'failed' && (
        <div className="alert-info mt-4">
          Campaign activation is handled by Skykin operators after moderation. Advertisers cannot self-activate.
        </div>
      )}
    </section>
  );
}
