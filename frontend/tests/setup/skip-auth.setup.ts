import { test as setup } from '@playwright/test';

/**
 * Setup file for SKIP_AUTH mode
 * Injects window.__SKIP_AUTH__ flag to bypass authentication in tests
 */

setup('configure skip auth', async ({ page }) => {
  // Inject SKIP_AUTH flag into page before navigation
  await page.addInitScript(() => {
    (window as any).__SKIP_AUTH__ = true;
  });
});
