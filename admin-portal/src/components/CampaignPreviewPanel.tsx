import { Skeleton } from '@skykin/ui';
import type { CampaignPreview } from '../types';

interface CampaignPreviewPanelProps {
  preview: CampaignPreview | null | undefined;
  loading?: boolean;
  error?: string;
}

function normalizePreviewContent(content: CampaignPreview['preview'] | Record<string, unknown>) {
  const raw = content as Record<string, unknown>;
  return {
    title: String(raw.title ?? raw.Title ?? ''),
    body_text: String(raw.body_text ?? raw.BodyText ?? ''),
    image_url: String(raw.image_url ?? raw.ImageURL ?? raw.imageUrl ?? ''),
    destination_url: String(raw.destination_url ?? raw.DestinationURL ?? ''),
  };
}

const isBanner = (f: string) => ['BANNER', 'IN_APP_BANNER', 'NATIVE_FEED'].includes(f.toUpperCase());
const isPush = (f: string) => ['PUSH', 'PUSH_PLUS'].includes(f.toUpperCase());
const isSms = (f: string) => f.toUpperCase() === 'SMS_PLUS';

export default function CampaignPreviewPanel({ preview, loading, error }: CampaignPreviewPanelProps) {
  if (loading) {
    return (
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <div className="border-b border-border bg-muted/50 px-4 py-3">
          <Skeleton className="h-4 w-32" />
        </div>
        <div className="p-5"><Skeleton className="h-40 w-full" /></div>
      </div>
    );
  }

  if (error || !preview) {
    return (
      <div className="flex items-center justify-center rounded-lg border border-dashed border-border bg-muted/40 p-8 text-center text-sm text-muted-foreground">
        {error ? `Preview unavailable — ${error}` : 'Preview unavailable.'}
      </div>
    );
  }

  const format = (preview.format || '').toUpperCase();
  const content = normalizePreviewContent(preview.preview);
  const hasImage = Boolean(content.image_url?.trim());

  return (
    <div className="overflow-hidden rounded-lg border border-border bg-card">
      <div className="flex items-center justify-between bg-primary px-4 py-3 text-primary-foreground">
        <div>
          <p className="text-[11px] uppercase tracking-wide opacity-80">Creative preview</p>
          <p className="font-semibold">{preview.channel_label || format}</p>
        </div>
        <span className="rounded-full border border-white/25 bg-white/15 px-2.5 py-0.5 text-[11px] font-medium">{preview.format}</span>
      </div>

      <div className="p-5">
        {isBanner(format) && (
          <div className="space-y-3">
            {hasImage ? (
              <img src={content.image_url} alt={preview.campaign_name} className="max-h-52 w-full rounded-lg border border-border bg-muted object-contain" onError={e => ((e.target as HTMLImageElement).style.display = 'none')} />
            ) : (
              <div className="flex h-32 items-center justify-center rounded-lg border border-dashed border-border bg-muted text-sm text-muted-foreground">No image provided</div>
            )}
            {(content.title || content.body_text) && (
              <div className="text-sm">
                {content.title && <p className="font-semibold">{content.title}</p>}
                {content.body_text && <p className="mt-1 text-muted-foreground">{content.body_text}</p>}
              </div>
            )}
          </div>
        )}

        {isPush(format) && (
          <div className="mx-auto max-w-sm overflow-hidden rounded-2xl border border-border shadow-md">
            <div className="bg-primary px-4 py-2 text-xs text-primary-foreground opacity-90">Push notification</div>
            <div className="bg-card p-4">
              <p className="font-semibold">{content.title || 'Title'}</p>
              <p className="mt-1 text-sm text-muted-foreground">{content.body_text || 'Body text'}</p>
              {hasImage && <img src={content.image_url} alt="" className="mt-3 max-h-36 w-full rounded-lg border border-border object-cover" />}
            </div>
          </div>
        )}

        {isSms(format) && (
          <div className="mx-auto max-w-md rounded-2xl border-2 border-border bg-muted/50 p-5">
            <p className="mb-2 text-xs font-medium text-muted-foreground">SMS+ interactive canvas</p>
            {hasImage && <img src={content.image_url} alt="" className="mb-3 max-h-40 w-full rounded-lg border border-border bg-card object-contain" />}
            <p className="font-bold">{content.title || 'Campaign title'}</p>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{content.body_text || 'Description'}</p>
          </div>
        )}

        {!isBanner(format) && !isPush(format) && !isSms(format) && (
          <div className="space-y-3">
            {hasImage && <img src={content.image_url} alt="" className="max-h-52 w-full rounded-lg border border-border object-contain" />}
            {content.title && <p className="font-semibold">{content.title}</p>}
            {content.body_text && <p className="text-sm text-muted-foreground">{content.body_text}</p>}
          </div>
        )}

        {content.destination_url && <p className="mt-4 truncate text-xs text-muted-foreground">→ {content.destination_url}</p>}
      </div>
    </div>
  );
}
