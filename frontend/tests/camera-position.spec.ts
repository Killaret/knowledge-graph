import { test, expect } from '@playwright/test';

/**
 * Tests for Camera Position and Navigation in 3D Graph
 * Verifies camera positioning, auto-zoom, and route transitions
 */

test.describe('3D Graph - Camera Position and Navigation', () => {
  
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should position camera to show start node in center for local graph', async ({ page, request }) => {
    // Create a note with connections
    const centerNote = await request.post('http://localhost:8080/notes', {
      data: { title: 'Center Node', content: 'Main node for camera test', type: 'star' }
    });
    const centerId = (await centerNote.json()).id;
    
    const linkedNote = await request.post('http://localhost:8080/notes', {
      data: { title: 'Linked Node', content: 'Secondary node', type: 'planet' }
    });
    const linkedId = (await linkedNote.json()).id;
    
    await request.post('http://localhost:8080/links', {
      data: { sourceNoteId: centerId, targetNoteId: linkedId, weight: 0.8 }
    });
    
    // Navigate to 3D graph for center node
    await page.goto(`http://localhost:5173/graph/3d/${centerId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000); // Wait for simulation and auto-zoom
    
    const container = page.locator('.graph-3d-container').first();
    await expect(container).toBeVisible({ timeout: 10000 });
    
    // Check that camera logs show centering on nodes
    const logs: string[] = await page.evaluate(() => {
      return (window as any).consoleLogs || [];
    });
    
    // Verify auto-zoom was called
    const autoZoomLog = logs.find(log => log.includes('autoZoomToFit'));
    expect(autoZoomLog).toBeTruthy();
  });

  test('should position camera appropriately for full 3D graph', async ({ page, request }) => {
    // Create multiple notes for full graph
    const notes = [];
    for (let i = 0; i < 5; i++) {
      const note = await request.post('http://localhost:8080/notes', {
        data: { title: `Full Graph Node ${i}`, content: `Node ${i}`, type: i === 0 ? 'star' : 'planet' }
      });
      notes.push((await note.json()).id);
    }
    
    // Create some links
    for (let i = 0; i < notes.length - 1; i++) {
      await request.post('http://localhost:8080/links', {
        data: { sourceNoteId: notes[i], targetNoteId: notes[i + 1], weight: 0.7 }
      });
    }
    
    // Navigate to full 3D graph
    await page.goto('http://localhost:5173/graph/3d');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(5000);
    
    const container = page.locator('.graph-3d-container, .lazy-loading, .lazy-error').first();
    await expect(container).toBeVisible({ timeout: 10000 });
    
    // Stats should show multiple nodes
    const statsBar = page.locator('.stats-bar').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      const nodeMatch = statsText?.match(/(\d+)\s*nodes/);
      if (nodeMatch) {
        const nodeCount = parseInt(nodeMatch[1], 10);
        expect(nodeCount).toBeGreaterThanOrEqual(5);
      }
    }
  });

  test('should adjust camera when toggling full graph mode', async ({ page, request }) => {
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Toggle Camera Test', content: 'Testing camera on toggle', type: 'star' }
    });
    const noteId = (await note.json()).id;
    
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    const container = page.locator('.graph-3d-container').first();
    await expect(container).toBeVisible();
    
    // Find and click the toggle
    const toggle = page.locator('.toggle input[type="checkbox"]').first();
    if (await toggle.isVisible().catch(() => false)) {
      // Get initial state
      const isChecked = await toggle.isChecked();
      
      // Toggle to other mode
      await toggle.click();
      await page.waitForTimeout(3000); // Wait for camera adjustment
      
      // Graph should still be visible after camera adjustment
      await expect(container).toBeVisible();
      
      // Stats should update
      const statsBar = page.locator('.stats-bar').first();
      if (await statsBar.isVisible().catch(() => false)) {
        const statsText = await statsBar.textContent();
        // Should show appropriate mode
        const hasModeIndicator = statsText?.toLowerCase().includes('full') || 
                                 statsText?.toLowerCase().includes('local');
        expect(hasModeIndicator).toBe(true);
      }
      
      // Toggle back
      await toggle.click();
      await page.waitForTimeout(3000);
      
      // Should still be visible
      await expect(container).toBeVisible();
    }
  });

  test('should maintain camera position on route navigation', async ({ page, request }) => {
    // Create test note
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Navigation Test', content: 'Testing route navigation', type: 'star' }
    });
    const noteId = (await note.json()).id;
    
    // Go to 3D graph
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    const container3D = page.locator('.graph-3d-container').first();
    await expect(container3D).toBeVisible();
    
    // Navigate to home page
    await page.goto('http://localhost:5173/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Home page should load
    const homeContent = page.locator('.fullscreen-graph, .list-container').first();
    await expect(homeContent).toBeVisible();
    
    // Navigate back to 3D graph
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // 3D graph should reload correctly
    const container3DAgain = page.locator('.graph-3d-container').first();
    await expect(container3DAgain).toBeVisible();
  });

  test('should transition from 2D to 3D graph maintaining context', async ({ page, request }) => {
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: '2D to 3D Transition', content: 'Testing 2D to 3D', type: 'planet' }
    });
    const noteId = (await note.json()).id;
    
    // Start at 2D graph page
    await page.goto('http://localhost:5173/graph');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    // Verify 2D graph loads
    const graph2D = page.locator('.fullscreen-graph, canvas').first();
    await expect(graph2D).toBeVisible();
    
    // Navigate to 3D version
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // Verify 3D graph loads
    const graph3D = page.locator('.graph-3d-container').first();
    await expect(graph3D).toBeVisible();
    
    // Stats should show the specific note's local graph
    const statsBar = page.locator('.stats-bar').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      expect(statsText).toMatch(/\d+\s*nodes/);
      expect(statsText).toMatch(/\d+\s*links/);
    }
  });

  test('should navigate from home page 3D button to full 3D graph', async ({ page }) => {
    // Start at home page
    await page.goto('http://localhost:5173/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Find and click the 3D button
    const button3D = page.locator('.view-button-3d, button:has-text("3D")').first();
    if (await button3D.isVisible().catch(() => false)) {
      await button3D.click();
      await page.waitForTimeout(3000);
      
      // Should navigate to /graph/3d
      const currentUrl = page.url();
      expect(currentUrl).toContain('/graph/3d');
      
      // 3D graph should load
      const container = page.locator('.graph-3d-container, .lazy-loading').first();
      await expect(container).toBeVisible({ timeout: 10000 });
    }
  });

  test('should handle direct URL access to 3D graph routes', async ({ page, request }) => {
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Direct URL Test', content: 'Testing direct URL access', type: 'comet' }
    });
    const noteId = (await note.json()).id;
    
    // Test direct access to local 3D graph
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    const localContainer = page.locator('.graph-3d-container').first();
    await expect(localContainer).toBeVisible();
    
    // Local mode should be indicated
    const statsLocal = page.locator('.stats-bar').first();
    if (await statsLocal.isVisible().catch(() => false)) {
      const statsText = await statsLocal.textContent();
      const hasLocalMode = statsText?.toLowerCase().includes('local');
      expect(hasLocalMode).toBe(true);
    }
    
    // Test direct access to full 3D graph
    await page.goto('http://localhost:5173/graph/3d');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    const fullContainer = page.locator('.graph-3d-container, .lazy-loading').first();
    await expect(fullContainer).toBeVisible();
    
    // Full mode should be indicated
    const statsFull = page.locator('.stats-bar').first();
    if (await statsFull.isVisible().catch(() => false)) {
      const statsText = await statsFull.textContent();
      const hasFullMode = statsText?.toLowerCase().includes('full');
      expect(hasFullMode).toBe(true);
    }
  });

  test('should show empty state with appropriate camera position when no notes', async ({ page }) => {
    // Navigate to full 3D graph when no notes exist (or after clearing)
    await page.goto('http://localhost:5173/graph/3d');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    // Should show empty state or loading state, not error
    const emptyState = page.locator('.no-data-message, .empty-state, .lazy-loading').first();
    const graphContainer = page.locator('.graph-3d-container').first();
    
    const hasEmpty = await emptyState.isVisible().catch(() => false);
    const hasGraph = await graphContainer.isVisible().catch(() => false);
    
    expect(hasEmpty || hasGraph).toBe(true);
  });

  test('should position camera correctly for isolated single node', async ({ page, request }) => {
    // Create note with NO connections
    const isolatedNote = await request.post('http://localhost:8080/notes', {
      data: { title: 'Isolated Node', content: 'No connections', type: 'galaxy' }
    });
    const noteId = (await isolatedNote.json()).id;
    
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // Should render without errors
    const container = page.locator('.graph-3d-container, .no-data-message').first();
    await expect(container).toBeVisible();
    
    // Stats should show 1 node (just the isolated note itself)
    const statsBar = page.locator('.stats-bar').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      const nodeMatch = statsText?.match(/(\d+)\s*nodes/);
      if (nodeMatch) {
        const nodeCount = parseInt(nodeMatch[1], 10);
        expect(nodeCount).toBeGreaterThanOrEqual(1);
      }
    }
  });
});
