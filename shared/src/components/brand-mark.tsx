import { cn } from '../lib/utils';
import logo from '../assets/skykin_logo.png';

/** Skykin cloud/loop mark (inline SVG, inherits currentColor). */
export function SkykinMark({ className }: { className?: string }) {
  return (
    <img src={logo} alt="Skykin Logo" aria-hidden className={cn('h-10 w-auto object-contain', className)} />
  );
}
