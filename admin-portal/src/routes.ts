/**
 * Route constants.
 *
 * `/` is the public Home page, so the operator dashboard needs its own path.
 * Kept out of App.tsx to avoid a cycle: App imports Layout, and Layout needs
 * this too.
 */
export const OVERVIEW_PATH = '/overview';
