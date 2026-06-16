import { validationTone } from '../lib/campaignUtils';

const toneClass: Record<string, string> = {
  success: 'badge-success',
  warning: 'badge-warning',
  danger: 'badge-danger',
  neutral: 'badge-neutral',
  brand: 'badge-brand',
};

export default function StatusBadge({
  label,
  tone,
}: {
  label: string;
  tone: 'success' | 'warning' | 'danger' | 'neutral' | 'brand';
}) {
  return (
    <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ring-1 ring-inset ring-[var(--border)] ${toneClass[tone]}`}>
      {label}
    </span>
  );
}

export function ValidationBadge({ status }: { status: string }) {
  return <StatusBadge label={status} tone={validationTone(status)} />;
}

export function ActiveBadge({ active }: { active: boolean }) {
  return (
    <StatusBadge
      label={active ? 'Active' : 'Inactive'}
      tone={active ? 'success' : 'neutral'}
    />
  );
}
