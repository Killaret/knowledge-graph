import { test, expect } from '@playwright/test';

/**
 * Public graph: unauthenticated users should see the read-only graph.
 */
test.describe('Public Graph', () => {
  test('loads graph for unauthenticated users', async ({ page }) => {
    // Ensure we are not in SKIP_AUTH mode
    await page.addInitScript(() => {
      localStorage.removeItem('__SKIP_AUTH__');
      if ('__SKIP_AUTH__' in window) {
        delete (window as any).__SKIP_AUTH__;
      }
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 15000 });

    const stats = page.locator('[data-testid="graph-stats"]');
    await expect(stats).toBeVisible({ timeout: 15000 });
    const statsText = await stats.textContent({ timeout: 5000 });
    expect(statsText).toMatch(/\d+\s+nodes?/i);
    expect(statsText).toMatch(/\d+\s+links?/i);
  });
});
