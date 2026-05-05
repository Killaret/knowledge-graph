import { test, expect } from '@playwright/test';
import { setupSkipAuth } from './helpers/testUtils';

/**
 * Test to verify SKIP_AUTH is working correctly
 */

test.describe('SKIP_AUTH Verification', () => {
  test('should access main page without authentication', async ({ page }) => {
    // Setup SKIP_AUTH first
    await setupSkipAuth(page);
    
    // Now navigate to main page
    await page.goto('/');
    await page.waitForTimeout(500);
    
    // Should NOT redirect to login
    await expect(page).toHaveURL('/');
    
    // Should show graph or notes (or at least not redirect)
    const currentUrl = page.url();
    expect(currentUrl).not.toContain('/auth/login');
  });
  
  test('should access graph page without authentication', async ({ page }) => {
    // Setup SKIP_AUTH first
    await setupSkipAuth(page);
    
    // Navigate to graph page
    await page.goto('/graph');
    await page.waitForTimeout(500);
    
    // Should NOT redirect to login
    await expect(page).toHaveURL('/graph');
    
    // Should show graph canvas
    const canvas = page.locator('canvas').first();
    await expect(canvas).toBeVisible({ timeout: 10000 });
  });
});
