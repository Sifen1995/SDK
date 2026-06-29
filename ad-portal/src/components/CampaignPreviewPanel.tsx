import type { CampaignPreview } from '../types';

interface CampaignPreviewPanelProps {
  preview: CampaignPreview | null;
  loading?: boolean;
}

export default function CampaignPreviewPanel({ preview, loading }: CampaignPreviewPanelProps) {
  if (loading) {
    return (
      <div className="card p-8 text-center text-muted">
        Loading preview…
      </div>
    );
  }

  if (!preview) {
    return (
      <div className="rounded-2xl border border-dashed border-[var(--border)] bg-[var(--bg-subtle)] p-8 text-center text-muted">
        Preview will appear here after the campaign is saved.
      </div>
    );
  }

  const { preview: content } = preview;
  const format = preview.format;

  return (
    <div className="card overflow-hidden">
      <div className="bg-[var(--accent)] text-[var(--accent-fg)] px-5 py-3 flex items-center justify-between">
        <div>
          <p className="opacity-70 text-xs uppercase tracking-wide">Simulator</p>
          <p className="font-semibold">{preview.channel_label}</p>
        </div>
        <span className="rounded-full border border-[var(--accent-fg)]/20 bg-[var(--accent-fg)]/10 px-3 py-1 text-xs">
          {preview.format}
        </span>
      </div>

      <div className="p-6">
        {format === 'BANNER' && (
          <div className="space-y-3">
            {content.image_url ? (
              <img
                src={content.image_url}
                alt={preview.campaign_name}
                className="w-full max-h-48 object-cover rounded-lg border border-[var(--border)]"
              />
            ) : (
              <div className="h-32 rounded-lg bg-[var(--bg-subtle)] flex items-center justify-center text-muted text-sm">
                No image
              </div>
            )}
            {(content.title || content.body_text) && (
              <div className="text-sm text-primary">
                {content.title && <p className="font-semibold">{content.title}</p>}
                {content.body_text && <p className="text-muted mt-1">{content.body_text}</p>}
              </div>
            )}
          </div>
        )}

        {format === 'PUSH_PLUS' && (
          <div className="max-w-sm mx-auto rounded-2xl border border-[var(--border)] shadow-lg overflow-hidden">
            <div className="bg-[var(--accent)] text-[var(--accent-fg)] px-4 py-2 text-xs opacity-80">Push notification</div>
            <div className="p-4 bg-[var(--surface)]">
              <p className="font-semibold text-primary">{content.title || 'Title'}</p>
              <p className="text-sm text-muted mt-1">{content.body_text || 'Body text'}</p>
              {content.image_url && (
                <img src={content.image_url} alt="" className="mt-3 rounded-lg w-full max-h-32 object-cover" />
              )}
            </div>
          </div>
        )}

        {format === 'SMS_PLUS' && (
          <div className="max-w-md mx-auto rounded-2xl border-2 border-[var(--border-strong)] bg-[var(--bg-subtle)] p-5">
            <p className="text-xs font-medium text-muted mb-2">SMS+ Interactive Canvas</p>
            {content.image_url && (
              <img src={content.image_url} alt="" className="w-full rounded-lg mb-3 max-h-36 object-cover" />
            )}
            <p className="font-bold text-primary">{content.title || 'Campaign title'}</p>
            <p className="text-sm text-muted mt-2 leading-relaxed">{content.body_text || 'Description'}</p>
            <button type="button" className="mt-4 w-full py-2.5 rounded-lg btn-primary text-sm">
              View offer
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
