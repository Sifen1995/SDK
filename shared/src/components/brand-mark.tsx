import { cn } from '../lib/utils';

/** Skykin cloud/loop mark (inline SVG, inherits currentColor). */
export function SkykinMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 64 64" fill="none" aria-hidden className={cn('size-5', className)}>
      <path
        d="M23.5 20.5c-6.9 0-12.5 5.6-12.5 12.5S16.6 45.5 23.5 45.5c4.7 0 8.8-2.6 10.9-6.4 2.1 3.8 6.2 6.4 10.9 6.4 6.9 0 12.5-5.6 12.5-12.5S52.1 20.5 45.2 20.5c-4.7 0-8.8 2.6-10.9 6.4-2.1-3.8-6.1-6.4-10.8-6.4Zm0 7.5c2.8 0 5 2.2 5 5s-2.2 5-5 5-5-2.2-5-5 2.2-5 5-5Zm21.7 0c2.8 0 5 2.2 5 5s-2.2 5-5 5-5-2.2-5-5 2.2-5 5-5Z"
        fill="currentColor"
      />
    </svg>
  );
}
