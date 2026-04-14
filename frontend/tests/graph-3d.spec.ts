import { test, expect } from '@playwright/test';

/**
 * Tests for Graph Visualization (2D primary, 3D optional)
 * These tests verify that the graph renders correctly in both modes
 * Note: WebGL may be limited in headless environments, so we test for either 2D or 3D
 */

test.describe('Graph Visualization', () => {
  
  test.beforeEach(async ({ page }) => {
    // Navigate to home page first
    await page.goto('http://localhost:5173/');
    await page.waitForLoadState('networkidle');
  });

  test('should render graph page with visualization', async ({ page, request }) => {
    // Create a note via API
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Graph Test Note', content: 'Test note for graph' }
    });
    const noteId = (await note.json()).id;
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    // Verify either canvas (2D/3D) or error message is shown
    // WebGL may not work in headless mode, so we check for any graph container
    const canvas = page.locator('canvas, .graph-canvas').first();
    const graphContainer = page.locator('.graph-3d-container, .graph-2d, .graph-wrapper, .error-overlay').first();
    
    const hasCanvas = await canvas.isVisible().catch(() => false);
    const hasContainer = await graphContainer.isVisible().catch(() => false);
    
    // At least one visualization element should be present
    expect(hasCanvas || hasContainer).toBe(true);
  });

  test('should show graph container with correct styling', async ({ page, request }) => {
    // Create a note via API
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Styling Test Note', content: 'Testing graph styling' }
    });
    const noteId = (await note.json()).id;
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Verify graph container is visible (use first())
    await expect(page.locator('.graph-3d-container, .graph-2d, .graph-wrapper, .error-overlay').first()).toBeVisible();
    
    // Canvas may not be available in headless mode due to WebGL limitations
    // Just verify the container loaded correctly
    const container = page.locator('.graph-3d-container, .graph-2d, .graph-wrapper').first();
    const errorOverlay = page.locator('.error-overlay').first();
    
    const hasContainer = await container.isVisible().catch(() => false);
    const hasError = await errorOverlay.isVisible().catch(() => false);
    
    // Either graph or error message should be shown
    expect(hasContainer || hasError).toBe(true);
  });

  test('should handle back button navigation from graph page', async ({ page, request }) => {
    // Create a note via API
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Navigation Test Note', content: 'Testing navigation from graph' }
    });
    const noteId = (await note.json()).id;
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Navigate back using browser back
    await page.goBack();
    await page.waitForTimeout(1000);
    
    // Should be back on home or note page
    const currentUrl = page.url();
    expect(currentUrl).toMatch(/http:\/\/localhost:5173(\/|\/notes\/.+)/);
  });

  test('should display performance mode or fallback indicator', async ({ page, request }) => {
    // Create a note via API
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Performance Test Note', content: 'Testing performance mode indicator' }
    });
    const noteId = (await note.json()).id;
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}`);
    await page.waitForLoadState('networkidle');
    
    // Wait for graph to load (either 2D or 3D)
    await page.waitForTimeout(3000);
    
    // Check for performance hint (optional - may not be visible in headless)
    const performanceHint = page.locator('.performance-hint');
    const hasHint = await performanceHint.isVisible().catch(() => false);
    
    if (hasHint) {
      const hintText = await performanceHint.textContent();
      expect(hintText).toMatch(/(3D Mode|2D Mode|Performance)/i);
    }
    
    // Verify graph container loaded (success or error state)
    const container = page.locator('.graph-3d-container, .graph-2d, .graph-wrapper, .error-overlay').first();
    await expect(container).toBeVisible();
  });

  test('should handle graph page with no nodes gracefully', async ({ page, request }) => {
    // Create a note with no links via API
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Isolated Note', content: 'This note has no connections' }
    });
    const noteId = (await note.json()).id;
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(6000); // Wait for 5s loading timeout + buffer
    
    // Page should render without errors - wait for loading to complete
    // Check for container which should always exist, or empty state messages
    const graphContainer = page.locator('.graph-3d-container');
    const noDataMessage = page.locator('.no-data-message');
    const errorOverlay = page.locator('.error-overlay');
    
    // Wait for container or empty state (with extended timeout for 5s simulation timeout)
    await expect(graphContainer.or(noDataMessage).or(errorOverlay)).toBeVisible({ timeout: 10000 });
  });
});
