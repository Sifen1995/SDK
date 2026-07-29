import type { ReactNode } from 'react';
import { SkykinMark } from '@skykin/ui';

/**
 * Developer-portal landing hero — the bold "front door" panel shown beside the
 * auth form on lg+. Distinct from the app shell in boldness (deep navy→blue
 * brand gradient), but clearly the same product family. Grounded in what the
 * developer portal is for: registering apps for the Skykin Flutter Ad SDK and
 * issuing API keys, so it leads with a real SDK snippet.
 */
export function AuthHero({
  eyebrow,
  title,
  blurb,
  chips,
  children,
}: {
  eyebrow: string;
  title: ReactNode;
  blurb: string;
  chips: string[];
  children?: ReactNode;
}) {
  return (
    <div className="brand-hero hidden w-[48%] flex-col justify-between overflow-hidden p-12 text-white lg:flex">
      {/* ambient highlights */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-90"
        style={{
          background:
            'radial-gradient(circle at 12% 18%, rgb(255 255 255 / 0.10), transparent 42%), radial-gradient(circle at 88% 92%, rgb(0 131 213 / 0.35), transparent 46%)',
        }}
      />
      <div className="relative">
        <span className="mb-9 flex size-11 items-center justify-center rounded-xl bg-white/12 ring-1 ring-white/20 [&_svg]:!text-white">
          <SkykinMark className="size-6" />
        </span>
        <p className="mb-3 text-xs font-semibold uppercase tracking-[0.18em] text-white/60">{eyebrow}</p>
        <h1 className="font-display text-[2rem] font-bold leading-[1.12]">{title}</h1>
        <p className="mt-4 max-w-md leading-relaxed text-white/75">{blurb}</p>
        {children}
      </div>
      <div className="relative flex flex-wrap gap-2 text-xs">
        {chips.map(t => (
          <span key={t} className="rounded-full border border-white/20 bg-white/10 px-3 py-1 font-medium text-white/90">
            {t}
          </span>
        ))}
      </div>
    </div>
  );
}

/**
 * Stylized SDK snippet card for the hero — the one glass surface allowed on a
 * landing hero (never over data). Static, decorative, developer-flavored.
 */
export function SdkSnippet() {
  const kw = 'text-[#8fd0f5]';
  const str = 'text-white';
  const dim = 'text-white/45';
  const id = 'text-white/85';
  return (
    <div className="mt-8 max-w-md overflow-hidden rounded-xl border border-white/15 bg-[#061a2c]/60 shadow-2xl backdrop-blur-md">
      <div className="flex items-center gap-1.5 border-b border-white/10 px-4 py-2.5">
        <span className="size-2.5 rounded-full bg-white/25" />
        <span className="size-2.5 rounded-full bg-white/25" />
        <span className="size-2.5 rounded-full bg-white/25" />
        <span className="ml-2 font-mono text-[11px] text-white/50">main.dart</span>
      </div>
      <pre className="overflow-x-auto px-4 py-4 font-mono text-[12.5px] leading-relaxed">
        <code>
          <span className={kw}>import</span> <span className={str}>'package:skykin_sdk/skykin_sdk.dart'</span><span className={dim}>;</span>
          {'\n\n'}
          <span className={kw}>await</span> <span className={id}>Skykin</span><span className={dim}>.</span><span className={id}>initialize</span><span className={dim}>(</span>
          {'\n'}
          {'  '}apiKey<span className={dim}>:</span> <span className={str}>'sk_live_••••'</span><span className={dim}>,</span>
          {'\n'}
          {'  '}intentSignals<span className={dim}>:</span> <span className={kw}>true</span><span className={dim}>,</span>
          {'\n'}
          <span className={dim}>);</span>
          {'\n\n'}
          <span className={dim}>// one call renders an intent-matched ad</span>
          {'\n'}
          <span className={kw}>final</span> <span className={id}>ad</span> <span className={dim}>=</span> <span className={kw}>await</span> <span className={id}>Skykin</span><span className={dim}>.</span><span className={id}>loadAd</span><span className={dim}>(</span>slot<span className={dim}>:</span> <span className={str}>'home_feed'</span><span className={dim}>);</span>
        </code>
      </pre>
    </div>
  );
}
