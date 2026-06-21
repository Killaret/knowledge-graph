import { test, expect } from '@playwright/test';

/**
 * Visual Regression Tests with Argos
 * Captures screenshots of key UI states for comparison
 * 
 * Key fixes:
 * - Use backend graph API instead of graph service (proxy not working in dev)
 * - Use note ID from first note for graph visualization
 * - Wait for link data to load
 * - Increase wait times for proper rendering
 * 
 * Requires: ARGOS_TOKEN environment variable
 * Screenshots saved to: argos-screenshots/
 */

test.describe('Visual Regression @visual', { tag: ['@visual'] }, () => {
  
  // Setup: Inject SKIP_AUTH flag before each test
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      (window as any).__SKIP_AUTH__ = true;
    });
  });

  test('Home page - default view', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Wait for actual content to load
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(5000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/home-default.png',
      fullPage: true 
    });
  });

  test('Home page - list view', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    
    // Try to click list toggle
    try {
      const listButton = await page.$('[data-testid="list-toggle"]');
      if (listButton) {
        await listButton.click();
        await page.waitForTimeout(3000);
      }
    } catch {
      console.log('List toggle not found, skipping');
    }
    
    await page.screenshot({ 
      path: 'argos-screenshots/home-list-view.png',
      fullPage: true 
    });
  });

  test('Home page - with filter', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    
    // Try to apply filter
    try {
      const filterButton = await page.$('[data-testid="filter-star"]');
      if (filterButton) {
        await filterButton.click();
        await page.waitForTimeout(3000);
      }
    } catch {
      console.log('Filter button not found, skipping');
    }
    
    await page.screenshot({ 
      path: 'argos-screenshots/home-filtered-stars.png',
      fullPage: true 
    });
  });

  test('2D Graph - loading state', async ({ page }) => {
    await page.goto('/graph');
    
    // Capture loading state quickly
    await page.waitForTimeout(1000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/2d-loading-state.png',
      fullPage: true 
    });
  });

  test('2D Graph - full view with links', async ({ page }) => {
    await page.goto('/graph');
    await page.waitForLoadState('networkidle');
    
    // Wait longer for graph to load
    await page.waitForTimeout(5000);
    
    // Enable full graph mode to show all links
    try {
      const fullGraphToggle = await page.$('input[type="checkbox"]');
      if (fullGraphToggle) {
        await fullGraphToggle.click();
        console.log('Enabled full graph mode');
        // Wait for full graph data to load with links
        await page.waitForTimeout(8000);
      }
    } catch {
      console.log('Full graph toggle not found, may already be enabled');
    }
    
    // Wait for graph to fully render with links
    await page.waitForTimeout(5000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/2d-full-view-with-links.png',
      fullPage: true 
    });
  });

  test('3D Graph - frozen notice (redirects to 2D)', async ({ page }) => {
    await page.goto('/graph/3d');
    
    // 3D is frozen, will redirect to 2D. Wait for redirect
    await page.waitForTimeout(2000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/3d-frozen-notice.png',
      fullPage: true 
    });
  });

  test('Search page', async ({ page }) => {
    await page.goto('/search');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(3000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/search-page.png',
      fullPage: true 
    });
  });

  test('Search with query', async ({ page }) => {
    await page.goto('/search');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    
    // Enter search query
    const searchInput = await page.$('input[type="search"], input[placeholder*="search"], [data-testid="search-input"]');
    if (searchInput) {
      await searchInput.fill('star');
      await page.keyboard.press('Enter');
      await page.waitForTimeout(5000);
    }
    
    await page.screenshot({ 
      path: 'argos-screenshots/search-with-query.png',
      fullPage: true 
    });
  });

  test('Empty state', async ({ page }) => {
    await page.goto('/search?q=nonexistentquery123456789');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(3000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/empty-state.png',
      fullPage: true 
    });
  });

  test('Responsive - desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(5000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/responsive-desktop.png',
      fullPage: true 
    });
  });

  test('Responsive - tablet', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(5000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/responsive-tablet.png',
      fullPage: true 
    });
  });

  test('Responsive - mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(5000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/responsive-mobile.png',
      fullPage: true 
    });
  });
});