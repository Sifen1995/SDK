/**
 * Route constants.
 *
 * `/` is the public Home page, so the campaigns list — the signed-in landing
 * page — needs its own path. Kept out of App.tsx to avoid a cycle: App imports
 * Layout, and Layout needs this too.
 */
export const CAMPAIGNS_PATH = '/campaigns';
