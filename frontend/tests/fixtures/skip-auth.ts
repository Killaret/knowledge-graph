import { test as base, expect } from "@playwright/test";

/**
 * Extended test fixture with SKIP_AUTH support
 * Automatically injects window.__SKIP_AUTH__ flag before page loads
 */

export const test = base.extend({
  page: async ({ browser }, use) => {
    // Create new context with script injection
    const context = await browser.newContext({
      bypassCSP: true,
    });

    // Add init script to set SKIP_AUTH flag before any page loads
    await context.addInitScript(() => {
      (window as any).__SKIP_AUTH__ = true;
      console.log("[SKIP_AUTH] Authentication bypass enabled");
    });

    const page = await context.newPage();

    // Use the page in tests
    await use(page);

    // Cleanup
    await context.close();
  },
});

export { expect };
