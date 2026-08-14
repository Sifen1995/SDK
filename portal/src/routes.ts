/**
 * Route constants.
 *
 * `/` is the public Home page, so the signed-in landing page needs its own path.
 * Kept out of App.tsx to avoid a cycle: App imports Layout, and Layout needs
 * this too.
 */
export const DASHBOARD_PATH = '/dashboard';
