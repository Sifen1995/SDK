import type { CampaignPreview } from '../types';

interface CampaignPreviewPanelProps {
  preview: CampaignPreview | null;
  loading?: boolean;
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

function isBannerFormat(format: string): boolean {
  const f = format.toUpperCase();
  return f === 'BANNER' || f === 'IN_APP_BANNER' || f === 'NATIVE_FEED';
}

function isPushFormat(format: string): boolean {
  const f = format.toUpperCase();
  return f === 'PUSH' || f === 'PUSH_PLUS';
}

function isSmsFormat(format: string): boolean {
  return format.toUpperCase() === 'SMS_PLUS';
}

export default function CampaignPreviewPanel({ preview, loading }: CampaignPreviewPanelProps) {
  if (loading) {
    return (
      <div className="preview-panel preview-panel-loading">
        <div className="preview-skeleton" />
        <p className="text-sm text-muted text-center mt-4">Loading preview…</p>
      </div>
    );
  }

  if (!preview) {
    return (
      <div className="preview-panel preview-panel-empty">
        <p className="text-muted text-sm">Preview unavailable.</p>
      </div>
    );
  }

  const format = (preview.format || '').toUpperCase();
  const content = normalizePreviewContent(preview.preview);
  const hasImage = Boolean(content.image_url?.trim());

  return (
    <div className="preview-panel overflow-hidden">
      <div className="preview-panel-header">
        <div>
          <p className="text-xs uppercase tracking-wide opacity-80">Creative preview</p>
          <p className="font-semibold">{preview.channel_label || format}</p>
        </div>
        <span className="preview-format-badge">{preview.format}</span>
      </div>

      <div className="p-5">
        {isBannerFormat(format) && (
          <div className="space-y-3">
            {hasImage ? (
              <img
                src={content.image_url}
                alt={preview.campaign_name}
                className="w-full max-h-52 object-contain rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]"
                onError={e => {
                  (e.target as HTMLImageElement).style.display = 'none';
                }}
              />
            ) : (
              <div className="preview-image-placeholder">No image provided</div>
            )}
            {(content.title || content.body_text) && (
              <div className="text-sm">
                {content.title && <p className="font-semibold text-primary">{content.title}</p>}
                {content.body_text && <p className="text-muted mt-1">{content.body_text}</p>}
              </div>
            )}
          </div>
        )}

        {isPushFormat(format) && (
          <div className="max-w-sm mx-auto rounded-2xl border border-[var(--border)] shadow-lg overflow-hidden">
            <div className="bg-[var(--accent)] text-[var(--accent-fg)] px-4 py-2 text-xs opacity-90">
              Push notification
            </div>
            <div className="p-4 bg-[var(--surface)]">
              <p className="font-semibold text-primary">{content.title || 'Title'}</p>
              <p className="text-sm text-muted mt-1">{content.body_text || 'Body text'}</p>
              {hasImage && (
                <img
                  src={content.image_url}
                  alt=""
                  className="mt-3 rounded-lg w-full max-h-36 object-cover border border-[var(--border)]"
                />
              )}
            </div>
          </div>
        )}

        {isSmsFormat(format) && (
          <div className="max-w-md mx-auto rounded-2xl border-2 border-[var(--border-strong)] bg-[var(--bg-subtle)] p-5">
            <p className="text-xs font-medium text-muted mb-2">SMS+ Interactive Canvas</p>
            {hasImage && (
              <img
                src={content.image_url}
                alt=""
                className="w-full rounded-lg mb-3 max-h-40 object-contain border border-[var(--border)] bg-[var(--surface)]"
              />
            )}
            <p className="font-bold text-primary">{content.title || 'Campaign title'}</p>
            <p className="text-sm text-muted mt-2 leading-relaxed">{content.body_text || 'Description'}</p>
          </div>
        )}

        {!isBannerFormat(format) && !isPushFormat(format) && !isSmsFormat(format) && (
          <div className="space-y-3">
            {hasImage && (
              <img
                src={content.image_url}
                alt=""
                className="w-full max-h-52 object-contain rounded-lg border border-[var(--border)]"
              />
            )}
            {content.title && <p className="font-semibold text-primary">{content.title}</p>}
            {content.body_text && <p className="text-sm text-muted">{content.body_text}</p>}
          </div>
        )}

        {content.destination_url && (
          <p className="text-xs text-faint mt-4 truncate">
            → {content.destination_url}
          </p>
        )}
      </div>
    </div>
  );
}
