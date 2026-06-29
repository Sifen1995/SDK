import type { PortalRole } from '../types';
import { ROLE_META } from '../types';

const styles: Record<PortalRole, string> = {
  operator_admin: 'badge-brand',
  advertiser: 'badge-success',
  read_only_analyst: 'badge-neutral',
};

export default function RoleBadge({ role, size = 'md' }: { role: PortalRole; size?: 'sm' | 'md' }) {
  const meta = ROLE_META[role];
  const sizeClass = size === 'sm' ? 'text-xs px-2 py-0.5' : 'text-xs px-2.5 py-1';
  return (
    <span
      className={`inline-flex items-center rounded-full font-medium border border-[var(--border)] ${styles[role]} ${sizeClass}`}
      title={meta.description}
    >
      {meta.label}
    </span>
  );
}
