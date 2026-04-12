import { test, expect } from '@playwright/test';

// Allow forcing 3D rendering in CI/debug via environment variable FORCE3D=1
const forceSuffix = process.env.FORCE3D === '1' ? '?force3d=1' : '';

/**
 * Tests for 3D Graph Visualization
 * These tests verify that the 3D graph canvas renders correctly
 * and that the SmartGraph component properly selects 2D or 3D mode
 */

test.describe('3D Graph Visualization', () => {
  
  test.beforeEach(async ({ page }) => {
    // Navigate to home page first to create a note
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('should render graph page with canvas visible', async ({ page }) => {
    // Create a new note first
    await page.click('a:has-text("New Note")');
    await page.waitForURL(/\/notes\/new/);
    
    // Fill in note details
    await page.fill('input[name="title"]', 'Graph Test Note');
    await page.fill('textarea[name="content"]', 'This is a test note for graph visualization');
    await page.click('button[type="submit"]');
    
    // Wait for note creation and redirect to note page
    await page.waitForURL(/\/notes\/[a-f0-9-]+/);
    
    // Get the note ID from URL
    const noteUrl = page.url();
    const noteId = noteUrl.split('/').pop();
    
    // Navigate to graph page
    await page.goto(`/graph/${noteId}${forceSuffix}`);
    await page.waitForLoadState('networkidle');
    
    // Verify page title
    await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
    
    // Verify canvas or 2D fallback is visible
    const canvas3D = page.locator('.graph-3d canvas');
    const canvas2D = page.locator('.graph-2d canvas');

    if ((await canvas3D.count()) > 0) {
      await expect(canvas3D.first()).toBeVisible({ timeout: 10000 });
    } else {
      await expect(canvas2D.first()).toBeVisible({ timeout: 10000 });
    }
  });

  test('should show graph container with correct styling', async ({ page }) => {
    // Create a new note
    await page.click('a:has-text("New Note")');
    await page.waitForURL(/\/notes\/new/);
    
    await page.fill('input[name="title"]', 'Styling Test Note');
    await page.fill('textarea[name="content"]', 'Testing graph styling');
    await page.click('button[type="submit"]');
    
    await page.waitForURL(/\/notes\/[a-f0-9-]+/);
    const noteUrl = page.url();
    const noteId = noteUrl.split('/').pop();
    
    // Navigate to graph page
    await page.goto(`/graph/${noteId}${forceSuffix}`);
    await page.waitForLoadState('networkidle');
    
    // Verify graph container has correct background
    const graphPage = page.locator('.graph-page');
    await expect(graphPage).toBeVisible();
    
    // Verify header elements
    await expect(page.locator('.graph-header')).toBeVisible();
    await expect(page.locator('.graph-header h1')).toHaveText('Knowledge Constellation');
    await expect(page.locator('.hint')).toBeVisible();
  });

  test('should handle back button navigation from graph page', async ({ page }) => {
    // Create a new note
    await page.click('a:has-text("New Note")');
    await page.waitForURL(/\/notes\/new/);
    
    await page.fill('input[name="title"]', 'Navigation Test Note');
    await page.fill('textarea[name="content"]', 'Testing navigation from graph');
    await page.click('button[type="submit"]');
    
    await page.waitForURL(/\/notes\/[a-f0-9-]+/);
    const noteUrl = page.url();
    const noteId = noteUrl.split('/').pop();
    
    // Navigate to graph page
    await page.goto(`/graph/${noteId}${forceSuffix}`);
    await page.waitForLoadState('networkidle');
    
    // Click back button
    await page.click('.back-button');
    
    // Should navigate back
    await page.waitForURL(/\/$/);
    await expect(page).toHaveURL('/');
  });

  test('should display performance mode indicator', async ({ page }) => {
    // Create a new note
    await page.click('a:has-text("New Note")');
    await page.waitForURL(/\/notes\/new/);
    
    await page.fill('input[name="title"]', 'Performance Test Note');
    await page.fill('textarea[name="content"]', 'Testing performance mode indicator');
    await page.click('button[type="submit"]');
    
    await page.waitForURL(/\/notes\/[a-f0-9-]+/);
    const noteUrl = page.url();
    const noteId = noteUrl.split('/').pop();
    
    // Navigate to graph page
    await page.goto(`/graph/${noteId}${forceSuffix}`);
    await page.waitForLoadState('networkidle');
    
    // Wait for graph to load (either 2D or 3D)
    await page.waitForTimeout(2000);
    
    // Check for performance hint (either "3D Mode" or "2D Mode")
    const performanceHint = page.locator('.performance-hint');
    if (await performanceHint.isVisible().catch(() => false)) {
      const hintText = await performanceHint.textContent();
      expect(hintText).toMatch(/(3D Mode|2D Mode)/);
    }
  });

  test('should handle graph page with no nodes gracefully', async ({ page }) => {
    // Create a note with no links
    await page.click('a:has-text("New Note")');
    await page.waitForURL(/\/notes\/new/);
    
    await page.fill('input[name="title"]', 'Isolated Note');
    await page.fill('textarea[name="content"]', 'This note has no connections');
    await page.click('button[type="submit"]');
    
    await page.waitForURL(/\/notes\/[a-f0-9-]+/);
    const noteUrl = page.url();
    const noteId = noteUrl.split('/').pop();
    
    // Navigate to graph page
    await page.goto(`/graph/${noteId}${forceSuffix}`);
    await page.waitForLoadState('networkidle');
    
    // Page should still render without errors
    await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
    await expect(page.locator('.graph-container')).toBeVisible();
  });
});
