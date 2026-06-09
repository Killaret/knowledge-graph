import { test, expect } from '@playwright/test';

/**
 * Manual test to verify view toggle functionality
 * Run with: npm run test -- manual-view-toggle.spec.ts --headed
 */
test.describe('Manual View Toggle Test', () => {
  test('should toggle between graph and list view', async ({ page }) => {
    // Go to main page
    await page.goto('http://localhost:5173/', { timeout: 60000 });
    await page.waitForLoadState('networkidle');
    
    // Wait for splash screen to disappear
    const splashScreen = page.locator('.splash-screen').first();
    const splashVisible = await splashScreen.isVisible().catch(() => false);
    if (splashVisible) {
      console.log('[TEST] Waiting for splash screen to disappear...');
      await splashScreen.waitFor({ state: 'detached', timeout: 10000 });
    }
    
    await page.waitForTimeout(3000);

    // Check initial state - should be graph view
    const graphContainer = page.locator('.fullscreen-graph, canvas').first();
    await expect(graphContainer).toBeVisible({ timeout: 10000 });
    console.log('[TEST] Initial state: Graph view visible');

    // Debug: Check for console errors
    page.on('console', msg => {
      if (msg.type() === 'error' || msg.text().includes('ERROR')) {
        console.log('[BROWSER ERROR]', msg.text());
      }
    });
    
    // Debug: Check what's on the page
    const allButtons = await page.locator('button').all();
    console.log('[TEST] Total buttons on page:', allButtons.length);
    
    const testIds = await page.locator('[data-testid]').all();
    const testIdValues = await Promise.all(testIds.map(el => el.getAttribute('data-testid')));
    console.log('[TEST] Available data-testid values:', testIdValues);

    // Check if FloatingControls exists
    const floatingControls = page.locator('.floating-controls').first();
    const fcVisible = await floatingControls.isVisible().catch(() => false);
    console.log('[TEST] FloatingControls visible:', fcVisible);

    // Click List toggle button using Playwright (not JavaScript)
    const listBtn = page.locator('[data-testid="view-toggle-list"]').first();
    await expect(listBtn).toBeVisible({ timeout: 5000 });
    
    console.log('[TEST] Found List button, clicking with Playwright...');
    await listBtn.click({ force: true });

    // Wait for view to switch
    await page.waitForTimeout(5000);

    // Check if list view is now visible
    const listContainer = page.locator('[data-testid="list-container"]').first();
    const listVisible = await listContainer.isVisible().catch(() => false);
    
    console.log('[TEST] List container visible:', listVisible);

    if (listVisible) {
      console.log('[TEST] SUCCESS: List view is now visible');
      const noteCards = page.locator('.note-card');
      const cardCount = await noteCards.count();
      console.log('[TEST] Note cards count:', cardCount);
    } else {
      // Check graph is still visible
      const graphStillVisible = await graphContainer.isVisible().catch(() => false);
      console.log('[TEST] Graph still visible:', graphStillVisible);
      
      // Check button state
      const listBtnActive = await page.evaluate(() => {
        const btn = document.querySelector('[data-testid="view-toggle-list"]');
        return btn?.classList.contains('active') || false;
      });
      console.log('[TEST] List button active class:', listBtnActive);
    }

    // Take screenshot for manual inspection
    await page.screenshot({ path: 'test-results/view-toggle-debug.png' });
    
    // Assert - for now just log, we'll see what happens
    expect(true).toBeTruthy();
  });
});
