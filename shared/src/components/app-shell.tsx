import * as React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Menu, X } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { cn } from '../lib/utils';

export type NavItem = {
  label: string;
  to: string;
  icon: LucideIcon;
  end?: boolean;
  badge?: React.ReactNode;
};
export type NavGroup = { label?: string; items: NavItem[] };

export type BrandLock = {
  name: string;
  sub: string;
  /** icon mark (SVG/img). */
  mark?: React.ReactNode;
};

function isActive(pathname: string, item: NavItem) {
  if (item.end) return pathname === item.to;
  return pathname === item.to || pathname.startsWith(item.to + '/');
}

/** Picks the single most-specific matching item so exactly one is active. */
function activeKey(pathname: string, groups: NavGroup[]) {
  let best = '';
  for (const g of groups) {
    for (const it of g.items) {
      if (isActive(pathname, it) && it.to.length >= best.length) best = it.to;
    }
  }
  return best;
}

/**
 * Shared collapsible sidebar + topbar shell for all three portals.
 * Signature element: a single "signal bar" (in the app's identity accent)
 * that slides to the active nav item.
 */
export function AppShell({
  brand,
  groups,
  title,
  topbarRight,
  sidebarFooter,
  children,
}: {
  brand: BrandLock;
  groups: NavGroup[];
  title?: React.ReactNode;
  topbarRight?: React.ReactNode;
  sidebarFooter?: React.ReactNode;
  children: React.ReactNode;
}) {
  const { pathname } = useLocation();
  const active = activeKey(pathname, groups);
  const [mobileOpen, setMobileOpen] = React.useState(false);

  // close the mobile drawer whenever the route changes
  React.useEffect(() => { setMobileOpen(false); }, [pathname]);

  const navRef = React.useRef<HTMLDivElement>(null);
  const itemRefs = React.useRef<Map<string, HTMLAnchorElement>>(new Map());
  const [bar, setBar] = React.useState<{ top: number; height: number; visible: boolean }>({
    top: 0,
    height: 0,
    visible: false,
  });

  React.useLayoutEffect(() => {
    const el = itemRefs.current.get(active);
    const nav = navRef.current;
    if (el && nav) {
      const top = el.offsetTop;
      const height = el.offsetHeight;
      setBar({ top, height, visible: true });
    } else {
      setBar(b => ({ ...b, visible: false }));
    }
  }, [active, groups]);

  return (
    <div className="flex min-h-screen bg-background">
      {/* mobile backdrop */}
      {mobileOpen && (
        <div
          className="fixed inset-0 z-40 bg-foreground/40 backdrop-blur-sm lg:hidden"
          aria-hidden
          onClick={() => setMobileOpen(false)}
        />
      )}

      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-50 flex h-screen w-60 shrink-0 flex-col border-r border-sidebar-border bg-sidebar transition-transform duration-200',
          'lg:sticky lg:top-0 lg:z-auto lg:translate-x-0',
          mobileOpen ? 'translate-x-0' : '-translate-x-full',
        )}
      >
        <div className="flex h-16 items-center gap-2.5 border-b border-sidebar-border px-5">
          {brand.mark && (
            <span className="flex size-9 items-center justify-center overflow-hidden rounded-md bg-identity/12">
              {brand.mark}
            </span>
          )}
          <div className="min-w-0">
            <p className="font-display text-sm font-semibold leading-tight text-sidebar-foreground">{brand.name}</p>
            <p className="text-[10px] uppercase tracking-wider text-sidebar-muted">{brand.sub}</p>
          </div>
          <button
            type="button"
            className="ml-auto rounded-md p-1.5 text-sidebar-muted hover:bg-sidebar-accent hover:text-sidebar-foreground lg:hidden"
            aria-label="Close menu"
            onClick={() => setMobileOpen(false)}
          >
            <X className="size-5" />
          </button>
        </div>

        <div ref={navRef} className="relative flex-1 overflow-y-auto px-3 py-4">
          {/* the signature moving signal bar */}
          <div
            className="signal-bar"
            aria-hidden
            style={{
              height: bar.height ? bar.height - 12 : 0,
              transform: `translateY(${bar.top + 6}px)`,
              opacity: bar.visible ? 1 : 0,
            }}
          />
          <nav className="space-y-6">
            {groups.map((group, gi) => (
              <div key={gi}>
                {group.label && (
                  <p className="px-3 pb-1.5 text-[10px] font-semibold uppercase tracking-wider text-sidebar-muted">
                    {group.label}
                  </p>
                )}
                <div className="space-y-0.5">
                  {group.items.map(item => {
                    const Icon = item.icon;
                    const on = active === item.to;
                    return (
                      <Link
                        key={item.to}
                        to={item.to}
                        ref={el => {
                          if (el) itemRefs.current.set(item.to, el);
                          else itemRefs.current.delete(item.to);
                        }}
                        aria-current={on ? 'page' : undefined}
                        className={cn(
                          'flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors',
                          on
                            ? 'bg-sidebar-accent font-medium text-sidebar-foreground'
                            : 'text-sidebar-muted hover:bg-sidebar-accent hover:text-sidebar-foreground',
                        )}
                      >
                        <Icon className={cn('size-4 shrink-0', on && 'text-identity')} strokeWidth={2} />
                        <span className="flex-1 truncate">{item.label}</span>
                        {item.badge}
                      </Link>
                    );
                  })}
                </div>
              </div>
            ))}
          </nav>
        </div>

        {sidebarFooter && <div className="border-t border-sidebar-border p-3">{sidebarFooter}</div>}
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-30 flex h-16 items-center justify-between gap-3 border-b border-border bg-background/85 px-4 backdrop-blur lg:px-8">
          <div className="flex min-w-0 items-center gap-2">
            <button
              type="button"
              className="-ml-1 rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground lg:hidden"
              aria-label="Open menu"
              onClick={() => setMobileOpen(true)}
            >
              <Menu className="size-5" />
            </button>
            <div className="min-w-0">{title}</div>
          </div>
          <div className="flex items-center gap-2.5">{topbarRight}</div>
        </header>
        <main className="flex-1 overflow-x-hidden px-6 py-6 lg:px-8">
          <div className="mx-auto w-full max-w-6xl">{children}</div>
        </main>
      </div>
    </div>
  );
}
