import { test, expect } from '@playwright/test';

/**
 * Tests for 3D Graph Visualization
 * These tests verify that the graph page renders correctly
 * with either 2D or 3D canvas
 */

// Store created note IDs for cleanup
const createdNoteIds: string[] = [];

test.afterEach(async ({ request }) => {
  // Cleanup all created notes
  for (const noteId of createdNoteIds) {
    try {
      await request.delete(`http://localhost:8080/notes/${noteId}`);
    } catch {
      // Ignore cleanup errors
    }
  }
  createdNoteIds.length = 0;
});

test.describe('3D Graph Visualization', () => {
  
  test('should render graph page with canvas visible', async ({ page, request }) => {
    // Create a note via API
    const timestamp = Date.now();
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Graph Test ' + timestamp, content: 'Test content' }
    });
    const noteData = await note.json();
    const noteId = noteData.id;
    createdNoteIds.push(noteId);
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}?force3d=1`);
    await page.waitForLoadState('networkidle');
    // Wait for Graph3D or fallback to be ready (onReady or DOM markers)
    await page.waitForFunction(() => {
      try {
        return (window as any).__graphReady === true
          || !!document.querySelector('[data-testid="graph-canvas"]')
          || !!document.querySelector('.graph-wrapper')
          || !!document.querySelector('.graph-container');
      } catch { return false; }
    }, { timeout: 30000 });

    // Verify page title
    await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
    
    // Verify canvas or graph container is visible
    const canvasByTestId = page.locator('[data-testid="graph-canvas"]');
    const canvas = page.locator('canvas');
    const graphContainer = page.locator('.graph-container');
    if ((await canvasByTestId.count()) > 0) {
      await expect(canvasByTestId).toBeVisible({ timeout: 5000 });
    } else if ((await canvas.count()) > 0) {
      await expect(canvas).toBeVisible({ timeout: 5000 });
    } else {
      await expect(graphContainer).toBeVisible({ timeout: 5000 });
    }
    
    // Verify WebGL is available (if canvas exists)
    const hasWebGL = await page.evaluate(() => {
      const canvas = document.querySelector('canvas');
      if (!canvas) return false;
      const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
      return !!gl;
    });
    expect(hasWebGL).toBe(true);
  });

  test('should show graph container with correct styling', async ({ page, request }) => {
    // Create a note via API
    const timestamp = Date.now();
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Styling Test ' + timestamp, content: 'Testing styling' }
    });
    const noteData = await note.json();
    const noteId = noteData.id;
    createdNoteIds.push(noteId);
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}?force3d=1`);
    await page.waitForLoadState('networkidle');
    
    // Verify graph container has correct structure
    await expect(page.locator('.graph-page')).toBeVisible();
    await expect(page.locator('.graph-header')).toBeVisible();
    await expect(page.locator('.graph-header h1')).toHaveText('Knowledge Constellation');
    await expect(page.locator('.hint')).toBeVisible();
    await expect(page.locator('.graph-container')).toBeVisible();
    
    // Verify canvas has correct data-testid
    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 30000 });
  });

  test('should handle back button navigation from graph page', async ({ page, request }) => {
    // Create a note via API
    const timestamp = Date.now();
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Navigation Test ' + timestamp, content: 'Testing navigation' }
    });
    const noteData = await note.json();
    const noteId = noteData.id;
    createdNoteIds.push(noteId);
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}?force3d=1`);
    await page.waitForLoadState('networkidle');
    
    // Click back button
    await page.click('.back-button');
    
    // Should navigate back to home
    await page.waitForURL(/\/$/);
    
    // Verify we're on home page
    await expect(page.locator('h1')).toHaveText('My Notes');
  });

  test('should display performance mode indicator', async ({ page, request }) => {
    // Create a note via API
    const timestamp = Date.now();
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Performance Test ' + timestamp, content: 'Testing performance indicator' }
    });
    const noteData = await note.json();
    const noteId = noteData.id;
    createdNoteIds.push(noteId);
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}?force3d=1`);
    await page.waitForLoadState('networkidle');
    
    // Wait for performance hint to appear (with better selector)
    const performanceHint = page.locator('.performance-hint');
    await expect(performanceHint).toBeVisible({ timeout: 5000 });
    
    // Check for performance hint text
    const hintText = await performanceHint.textContent();
    expect(hintText).toMatch(/(3D Mode|2D Mode)/);
  });

  test('should handle graph page with no connections gracefully', async ({ page, request }) => {
    // Create an isolated note via API
    const timestamp = Date.now();
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Isolated Note ' + timestamp, content: 'No connections' }
    });
    const noteData = await note.json();
    const noteId = noteData.id;
    createdNoteIds.push(noteId);
    
    // Navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}?force3d=1`);
    await page.waitForLoadState('networkidle');
    
    // Page should render without errors even with no connections
    await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
    await expect(page.locator('.graph-container')).toBeVisible();
    
    // Should show canvas (single node)
    const canvas = page.locator('canvas');
    await expect(canvas).toBeVisible();
    
    // Verify no error message is shown
    const errorMessage = page.locator('.error-state');
    await expect(errorMessage).not.toBeVisible();
  });

  test('should handle empty data gracefully', async ({ page }) => {
    // Navigate to graph page with invalid ID (will result in empty graph)
    await page.goto('http://localhost:5173/graph/non-existent-id');
    await page.waitForLoadState('networkidle');
    
    // Page should show error or empty state
    const emptyState = page.locator('.empty-state');
    const errorState = page.locator('.error-state');
    
    // Either empty or error should be visible
    const isEmptyVisible = await emptyState.isVisible().catch(() => false);
    const isErrorVisible = await errorState.isVisible().catch(() => false);
    
    expect(isEmptyVisible || isErrorVisible).toBe(true);
  });
});

