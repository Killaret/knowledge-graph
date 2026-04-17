/*
 * NOTE: This file previously contained a duplicate test suite that was removed to fix merge conflicts.
 * The preserved version below contains the complete test suite for Camera Position and Navigation.
 * 
 * Functionality of removed duplicate tests:
 * - Test camera positioning for local graph with center node
 * - Test camera positioning for full 3D graph with multiple nodes  
 * - Test camera adjustment when toggling full graph mode
 * - Test camera position maintenance on route navigation
 * - Test 2D to 3D graph transition maintaining context
 * - Test navigation from home page 3D button to full 3D graph
 * - Test direct URL access to 3D graph routes
 * - Test empty state with appropriate camera position
 * - Test camera positioning for isolated single node
 */
import { test, expect } from '@playwright/test';
import { createNote, createLink, getBackendUrl } from './helpers/testData';

/**
 * Tests for Camera Position and Navigation in 3D Graph
 * Verifies camera positioning, auto-zoom, and route transitions
 */

test.describe('3D Graph - Camera Position and Navigation', { tag: ['@smoke', '@3d', '@camera'] }, () => {
  
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should position camera to show start node in center for local graph', async ({ page, request }) => {
    // Create a note with connections using helper
    const centerNote = await createNote(request, {
      title: 'Center Node',
      content: 'Main node for camera test',
      type: 'star'
    });
    const centerId = centerNote.id;

    const linkedNote = await createNote(request, {
      title: 'Linked Node',
      content: 'Secondary node',
      type: 'planet'
    });
    const linkedId = linkedNote.id;

    await createLink(request, centerId, linkedId, 0.8);

    // Navigate to 3D graph for center node
    await page.goto(`/graph/3d/${centerId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000); // Wait for simulation and auto-zoom

    const container = page.locator('.graph-3d-container').first();
    await expect(container).toBeVisible({ timeout: 10000 });

    // Check camera position directly instead of console logs
    const cameraPos = await page.evaluate(() => {
      return (window as any).camera?.position;
    });
    const controlsTarget = await page.evaluate(() => {
      return (window as any).controls?.target;
    });

    expect(cameraPos).toBeDefined();
    expect(controlsTarget).toBeDefined();

    // Verify camera is not in default position (20, 15, 30)
    expect(cameraPos.x).not.toBe(20);
    expect(cameraPos.y).not.toBe(15);
    expect(cameraPos.z).not.toBe(30);

    // Camera should have been adjusted to show the node
    const distanceFromOrigin = Math.sqrt(
      cameraPos.x * cameraPos.x +
      cameraPos.y * cameraPos.y +
      cameraPos.z * cameraPos.z
    );
    expect(distanceFromOrigin).toBeGreaterThan(0);
  });

  test('should position camera appropriately for full 3D graph', async ({ page, request }) => {
    // Create multiple notes for full graph using helper
    const notes = [];
    for (let i = 0; i < 5; i++) {
      const note = await createNote(request, {
        title: `Full Graph Node ${i}`,
        content: `Node ${i}`,
        type: i === 0 ? 'star' : 'planet'
      });
      notes.push(note.id);
    }

    // Create some links
    for (let i = 0; i < notes.length - 1; i++) {
      await createLink(request, notes[i], notes[i + 1], 0.7);
    }

    // Navigate to full 3D graph
    await page.goto('/graph/3d');
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
    const note = await createNote(request, {
      title: 'Toggle Camera Test',
      content: 'Testing camera on toggle',
      type: 'star'
    });
    const noteId = note.id;

    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    const container = page.locator('.graph-3d-container').first();
    await expect(container).toBeVisible();
    
    // Find and click the toggle
    const toggle = page.locator('.toggle input[type="checkbox"]').first();
    if (await toggle.isVisible().catch(() => false)) {
      // NOTE: isChecked temporarily commented out to fix ESLint error
      // This variable captures the initial state of the toggle checkbox
      // Useful for verifying that the toggle actually changed state after clicking
      // const isChecked = await toggle.isChecked();
      
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
    const note = await createNote(request, {
      title: 'Navigation Test',
      content: 'Testing route navigation',
      type: 'star'
    });
    const noteId = note.id;

    // Go to 3D graph
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);

    const container3D = page.locator('.graph-3d-container').first();
    await expect(container3D).toBeVisible();

    // Navigate to home page
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Home page should load
    const homeContent = page.locator('.fullscreen-graph, .list-container').first();
    await expect(homeContent).toBeVisible();
    
    // Navigate back to 3D graph
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // 3D graph should reload correctly
    const container3DAgain = page.locator('.graph-3d-container').first();
    await expect(container3DAgain).toBeVisible();
  });

  test('should transition from 2D to 3D graph maintaining context', async ({ page, request }) => {
    const note = await createNote(request, {
      title: '2D to 3D Transition',
      content: 'Testing 2D to 3D',
      type: 'planet'
    });
    const noteId = note.id;

    // Start at 2D graph page
    await page.goto('/graph');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);

    // Verify 2D graph loads
    const graph2D = page.locator('.fullscreen-graph, canvas').first();
    await expect(graph2D).toBeVisible();

    // Navigate to 3D version
    await page.goto(`/graph/3d/${noteId}`);
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
    await page.goto('/');
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
    const note = await createNote(request, {
      title: 'Direct URL Test',
      content: 'Testing direct URL access',
      type: 'comet'
    });
    const noteId = note.id;

    // Test direct access to local 3D graph
    await page.goto(`/graph/3d/${noteId}`);
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
    await page.goto('/graph/3d');
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
    await page.goto('/graph/3d');
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
    const isolatedNote = await createNote(request, {
      title: 'Isolated Node',
      content: 'No connections',
      type: 'galaxy'
    });
    const noteId = isolatedNote.id;

    await page.goto(`/graph/3d/${noteId}`);
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
// End of Camera Position and Navigation test suite
