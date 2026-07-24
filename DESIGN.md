# Skykin Frontend Design System — "Signal"

One shared design system powers all three portals. This document is the source of
truth for tokens, type, the per-app identity rule, and the frontend architecture so a
new engineer can extend any portal without re-deriving these decisions.

## Apps & audiences
| App | Path prefix | Audience | Identity accent |
|---|---|---|---|
| `portal` | `/api/v1/portal` | **Developers** — SDK app registration + API keys | Steel blue `#3D6B99` |
| `ad-portal` | `/api/v1/ad-portal` | **Advertisers** — campaigns, subscription, audience | Warm bronze `#9C6B3F` |
| `admin-portal` | `/api/v1/ad-portal/admin` | **Operators** — analytics, moderation, catalog, RBAC | Muted plum `#6B5B95` |

## Shared package
Everything reusable lives in **`shared/` (`@skykin/ui`)**, imported by all three apps
(npm workspace). Nothing is duplicated per app.
- `styles/tokens.css` — the single token file (imported once per app as `@skykin/ui/styles.css`).
- `components/ui/*` — shadcn primitives (Button, Card, Badge, Input, Label, Dialog, Table,
  Tabs, Select, DropdownMenu, Tooltip, Avatar, Separator, Skeleton).
- `components/*` — `AppShell`, `KpiCard`, `StatusPill`, `SkykinMark`, `ThemeToggle`, and the
  state set (`LoadingState`, `EmptyState`, `ErrorState`, `InlineError`, `TableSkeleton`).
- `query/`, `providers.tsx`, `table/`, `form/`, `charts/` — the data + chart layer.

## Color
Tokens follow the **shadcn contract** (`--background`, `--foreground`, `--primary`,
`--muted`, `--border`, `--ring`, `--card`, `--popover`, `--destructive`, `--chart-1..6`,
`--sidebar-*`, `--radius`) plus Skykin additions (`--success`, `--warning`,
`--*-surface`, `--identity`). Components reference **only** these tokens via Tailwind
(`bg-primary`, `text-muted-foreground`, …) — never a raw hex or palette utility.

- **Core accent (shared):** deep muted teal `--primary #1F4E4A` (lifts to `#4F9A8F` on dark).
  Used for **every primary action in all three apps**.
- **Neutrals:** cool, teal-biased — paper `#FCFCFF`, ink `#131A22`, slate `#5C6773`, line `#E3E7EC`.
- **Semantic (separate from accent):** success `#2E7D5B`, warning `#B7791F`, destructive `#B4232A`,
  each with a low-chroma `*-surface` for pills/tiles. Status pills use `StatusPill` /
  `statusTone()` — a muted 3-color scheme (green/amber/red) + neutral. No rainbow.
- **Identity accent (`--identity`):** set by `data-app` on `<html>` (`portal|ad|admin`).
  **Wayfinding only** — the active-nav signal bar, active nav icon, avatars, wordmark, and
  nav-scoped badges. Never a primary button or large fill.
- **Charts:** one shared theme (`@skykin/ui` `chartAxis/chartGrid/chartTooltip/chartColor/
  CHART_COLORS`) drawing from `--chart-1..6`. Reused across every recharts view.

## Type (self-hosted via `@fontsource`, no CDN)
- **Display — Schibsted Grotesk** (`--font-display`): headings, wordmark, KPI numbers.
- **Body/UI — Hanken Grotesk** (`--font-sans`): all UI text and labels.
- **Numeric — IBM Plex Mono** (`--font-mono`): apply `.tabular-nums` to columns of digits
  (KPI tiles, tables, chart axes) so numbers align.

## Layout, elevation, motion
- Spacing 4/8/12/16/24/32; radius scale off `--radius: 0.5rem` (`--radius-sm/md/lg/xl`).
- **Hairline-border-first**: rely on `border-border` + subtle `shadow-sm`; two elevation
  levels only (cards, overlays). No heavy drop-shadow cards.
- Motion is near-instant (150–200ms) and purposeful; `prefers-reduced-motion` is honored
  globally in `tokens.css`.
- **Responsive:** the `AppShell` sidebar is a static rail on `lg+` and an off-canvas drawer
  below it (hamburger in the topbar, backdrop, auto-closes on navigation). Content and tables
  scroll within their own containers so the page body never scrolls sideways.

## Signature — the moving "signal bar"
The one animated element: a single vertical bar in the sidebar (`.signal-bar`, colored by
`--identity`) that **slides to the active nav item** (`AppShell`). It simultaneously encodes
wayfinding and which portal you're in, nodding to the product's "intent signal" idea.
Everything else in the shell is still.

## One-signature-per-app rule
Apps are identical in spacing, type, component shapes, motion, and chart style. **The only
per-app difference is the identity accent** (signal bar + wordmark sub-label + nav-scoped
badges) via `data-app`. Same family, different room.

## Frontend architecture (all three apps)
- **Server state → TanStack Query.** Each app has `lib/queries.ts` with a `qk` **query-key
  factory** and typed hooks. Mutations use **granular invalidation** (`invalidateQueries({
  queryKey: qk.<resource>.<...> })`) — never a blanket clear.
- **URL state → nuqs** (`useQueryState(s)`) for pagination, filters, tabs — survives refresh
  and is shareable. Wired via `AppProviders` (`NuqsAdapter` for React Router v7).
- **Forms → controlled inputs / TanStack Form** with the shared `Input`/`Label`/`InlineError`.
- **Tables → `DataTable`** (TanStack Table) with shared `exportToCsv` + `useListUrlState`.
- **Data states** are always surfaced: `LoadingState`/`TableSkeleton` while pending,
  `ErrorState`/`InlineError` on failure (no swallowed errors), `EmptyState` for empty sets.
- **API clients** stay in each app's `lib/api.ts`; all calls go through it (no raw `fetch` in
  pages).

## Adding a page
1. Add or reuse a hook in `lib/queries.ts` (key in `qk`, granular invalidation for mutations).
2. Build with `@skykin/ui` primitives + tokens — no hex, no palette utilities.
3. Put list/table state in nuqs; render `LoadingState`/`ErrorState`/`EmptyState`.
4. Add the route in `App.tsx` and a nav item in the app's `Layout.tsx`.

## Backend-contract notes (from the Phase 2 audit)
- Fixed: admin segment-candidates now calls `/admin/audience/segment-candidates` (was 404).
- Newly surfaced admin capabilities: SDK-users list (`/admin/sdk-users`, paginated), plan
  edit/suspend + billing-rate edit, segment suspend, and full **Roles & Permissions**
  (`/admin/roles`, `/admin/permissions`).
- `events`/`geofencing` backend routes are intentionally **not mounted** — no UI calls them.
