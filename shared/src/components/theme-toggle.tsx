import { Moon, Sun } from 'lucide-react';
import { Button } from './ui/button';

/**
 * Presentational theme toggle. The consuming app owns the theme state
 * (reads/writes .dark on <html>) and passes it in — keeps the shared package
 * free of app-specific storage keys.
 */
export function ThemeToggle({ isDark, onToggle }: { isDark: boolean; onToggle: () => void }) {
  return (
    <Button
      variant="outline"
      size="icon"
      onClick={onToggle}
      aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
      title={isDark ? 'Light mode' : 'Dark mode'}
    >
      {isDark ? <Sun className="size-4" /> : <Moon className="size-4" />}
    </Button>
  );
}
