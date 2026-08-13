import type { Campaign } from '../types';

/** Go serializes Campaign with PascalCase field names — normalize for the UI. */
export function normalizeCampaign(raw: Record<string, unknown>): Campaign {
  const canvas = raw.CanvasJSON ?? raw.canvas_json ?? raw.canvasJson ?? {};
  let canvasJson: Record<string, unknown> = {};
  if (typeof canvas === 'string') {
    try {
      canvasJson = JSON.parse(canvas) as Record<string, unknown>;
    } catch {
      canvasJson = {};
    }
  } else if (canvas && typeof canvas === 'object') {
    canvasJson = canvas as Record<string, unknown>;
  }

  return {
    id: String(raw.ID ?? raw.id ?? ''),
    advertiserId: String(raw.AdvertiserID ?? raw.advertiser_id ?? raw.advertiserId ?? ''),
    name: String(raw.Name ?? raw.name ?? ''),
    targetIntent: String(raw.TargetIntent ?? raw.target_intent ?? raw.targetIntent ?? ''),
    creativeFormat: String(raw.CreativeFormat ?? raw.creative_format ?? raw.creativeFormat ?? ''),
    title: String(raw.Title ?? raw.title ?? ''),
    bodyText: String(raw.BodyText ?? raw.body_text ?? raw.bodyText ?? ''),
    imageUrl: String(raw.ImageURL ?? raw.image_url ?? raw.imageUrl ?? ''),
    destinationUrl: String(raw.DestinationURL ?? raw.destination_url ?? raw.destinationUrl ?? ''),
    canvasJson,
    dailyBudgetCap: Number(raw.DailyBudgetCap ?? raw.daily_budget_cap ?? raw.dailyBudgetCap ?? 0),
    totalBudgetCap: Number(raw.TotalBudgetCap ?? raw.total_budget_cap ?? raw.totalBudgetCap ?? 0),
    isActive: Boolean(raw.IsActive ?? raw.is_active ?? raw.isActive ?? false),
    validationStatus: String(raw.ValidationStatus ?? raw.validation_status ?? raw.validationStatus ?? 'pending'),
    validationNotes: String(raw.ValidationNotes ?? raw.validation_notes ?? raw.validationNotes ?? ''),
    moderationStatus: String(raw.ModerationStatus ?? raw.moderation_status ?? raw.moderationStatus ?? 'pending'),
    moderationNotes: String(raw.ModerationNotes ?? raw.moderation_notes ?? raw.moderationNotes ?? ''),
    channelId: String(raw.ChannelID ?? raw.channel_id ?? raw.channelId ?? ''),
    channelCode: String(raw.ChannelCode ?? raw.channel_code ?? raw.channelCode ?? ''),
    createdAt: String(raw.CreatedAt ?? raw.created_at ?? raw.createdAt ?? ''),
    updatedAt: String(raw.UpdatedAt ?? raw.updated_at ?? raw.updatedAt ?? ''),
  };
}

export function formatLabel(value: string): string {
  return value.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
}

export function formatDate(iso: string): string {
  if (!iso) return '—';
  try {
    return new Intl.DateTimeFormat(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(new Date(iso));
  } catch {
    return iso;
  }
}

export function validationTone(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'passed':
      return 'success';
    case 'failed':
      return 'danger';
    case 'pending':
      return 'warning';
    default:
      return 'neutral';
  }
}

export function moderationTone(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'approved':
      return 'success';
    case 'rejected':
      return 'danger';
    case 'pending':
      return 'warning';
    default:
      return 'neutral';
  }
}

export function campaignsFromList(raw: unknown): Campaign[] {
  if (!raw || typeof raw !== 'object') return [];
  const obj = raw as { campaigns?: unknown[] };
  if (!Array.isArray(obj.campaigns)) return [];
  return obj.campaigns.map(item => normalizeCampaign(item as Record<string, unknown>));
}
