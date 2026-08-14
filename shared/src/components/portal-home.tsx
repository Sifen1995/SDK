import type { ComponentType, ReactNode } from 'react';
import { Card, CardContent } from './ui/card';
import { SkykinMark } from './brand-mark';
import { cn } from '../lib/utils';

export interface HomeFeature {
  /** A lucide icon component. */
  icon: ComponentType<{ className?: string }>;
  title: string;
  description: string;
}

export interface PortalHomeProps {
  /** Small label above the headline, e.g. "Developer portal". */
  eyebrow: string;
  headline: string;
  /** One or two sentences. Keep it to what the portal actually does. */
  subhead: string;
  /** Sign-up / sign-in buttons. Portals differ: admin has no self-registration. */
  actions: ReactNode;
  features: HomeFeature[];
  /** Optional numbered "how it works" strip under the feature grid. */
  steps?: string[];
  /** Rendered top-right — the theme toggle. */
  utility?: ReactNode;
}

/**
 * The public landing page shell each portal renders at `/`.
 *
 * Shared *structure* only — every portal passes its own copy, because the three
 * audiences (SDK developers, advertisers, operators) have nothing in common.
 * Uses the established tokens: `brand-hero` for the header gradient, `lift` on
 * the feature cards, `--identity` for the per-app accent.
 */
export function PortalHome({
  eyebrow,
  headline,
  subhead,
  actions,
  features,
  steps,
  utility,
}: PortalHomeProps) {
  return (
    <div className="min-h-screen bg-background">
      <header className="brand-hero relative overflow-hidden text-white">
        <div
          className="pointer-events-none absolute inset-0"
          style={{
            background:
              'radial-gradient(circle at 12% 88%, rgb(255 255 255 / 0.14), transparent 46%)',
          }}
        />
        <div className="relative mx-auto flex max-w-5xl items-center justify-between px-6 py-5">
          <span className="flex items-center gap-2.5 font-display text-lg font-bold">
            <span className="flex size-9 items-center justify-center rounded-lg bg-white/15">
              <SkykinMark className="size-5 text-white" />
            </span>
            Skykin
          </span>
          {utility}
        </div>

        <div className="relative mx-auto max-w-5xl px-6 pb-20 pt-10 sm:pb-24 sm:pt-16">
          <p className="text-xs font-bold uppercase tracking-[0.18em] text-white/70">
            {eyebrow}
          </p>
          <h1 className="mt-3 max-w-2xl font-display text-4xl font-bold leading-tight sm:text-5xl">
            {headline}
          </h1>
          <p className="mt-5 max-w-xl text-lg leading-relaxed text-white/85">
            {subhead}
          </p>
          <div className="mt-8 flex flex-wrap gap-3">{actions}</div>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-6 py-14 sm:py-20">
        <div
          className={cn(
            'grid gap-4',
            features.length >= 3 ? 'sm:grid-cols-2 lg:grid-cols-3' : 'sm:grid-cols-2',
          )}
        >
          {features.map(({ icon: Icon, title, description }) => (
            <Card key={title} className="lift">
              <CardContent className="p-5">
                <span className="mb-3 flex size-10 items-center justify-center rounded-lg bg-identity/10">
                  <Icon className="size-5 text-identity" />
                </span>
                <h2 className="font-display font-semibold">{title}</h2>
                <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
                  {description}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>

        {steps && steps.length > 0 && (
          <div className="mt-12 rounded-xl border border-border bg-muted/40 p-6">
            <h2 className="font-display font-semibold">How it works</h2>
            <ol className="mt-4 grid gap-4 sm:grid-cols-3">
              {steps.map((step, i) => (
                <li key={step} className="flex gap-3 text-sm">
                  <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-identity text-xs font-bold text-identity-foreground">
                    {i + 1}
                  </span>
                  <span className="leading-relaxed text-muted-foreground">{step}</span>
                </li>
              ))}
            </ol>
          </div>
        )}
      </main>

      <footer className="border-t border-border">
        <div className="mx-auto max-w-5xl px-6 py-6 text-sm text-muted-foreground">
          Skykin — intent-driven ad delivery.
        </div>
      </footer>
    </div>
  );
}
