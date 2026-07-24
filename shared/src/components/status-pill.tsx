import { Badge } from './ui/badge';

export type StatusTone = 'success' | 'warning' | 'destructive' | 'neutral';

/** Maps common backend status strings to a muted 3-color scheme (+neutral). */
export function statusTone(status: string): StatusTone {
  const s = status?.toLowerCase?.() ?? '';
  if (['approved', 'active', 'passed', 'valid', 'live', 'ready', 'succeeded', 'success'].includes(s)) return 'success';
  if (['pending', 'pending_review', 'in_review', 'processing', 'queued', 'draft', 'unbilled'].includes(s)) return 'warning';
  if (['rejected', 'failed', 'suspended', 'inactive', 'error', 'cancelled', 'canceled', 'expired'].includes(s)) return 'destructive';
  return 'neutral';
}

const toneToVariant = {
  success: 'success',
  warning: 'warning',
  destructive: 'destructive',
  neutral: 'secondary',
} as const;

function label(s: string) {
  return s ? s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) : '—';
}

/** Status pill; pass a tone or let it be inferred from the status string. */
export function StatusPill({ status, tone }: { status: string; tone?: StatusTone }) {
  const resolved = tone ?? statusTone(status);
  return <Badge variant={toneToVariant[resolved]}>{label(status)}</Badge>;
}
