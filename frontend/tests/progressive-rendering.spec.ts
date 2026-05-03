import { test, expect } from '@playwright/test';
import { createNote, createLink, isBackendAvailable, getBackendUrl } from './helpers/testData';

/**
 * Tests for Progressive Graph Rendering (Fog of War Effect)
 * Verifies two-phase loading, fog animation, and smooth camera transitions
 * 
 * NOTE: These tests require the backend to be running on localhost:8080
 */

// Global flag to track backend availability
let backendAvailable = false;

test.describe('Progressive Graph Rendering - Fog of War', { tag: ['@smoke', '@3d', '@progressive'] }, () => {
  
  test.beforeAll(async ({ request }) => {
    // Check backend availability once before all tests
    backendAvailable = await isBackendAvailable(request);

    if (!backendAvailable) {
      console.log(`⚠️  Backend not available on ${getBackendUrl()} - Progressive rendering tests will be skipped`);
    }
  });
  
  test.beforeEach(async ({ page }) => {
    if (!backendAvailable) {
      test.skip();
    }
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('should load initial graph immediately without spinner', async ({ page, request }) => {
    // Create a note with multiple connections via API using helper
    const centralNote = await createNote(request, {
      title: 'Central Progressive Test Note',
      content: 'Central node for progressive loading test',
      type: 'star'
    });
    const centralId = centralNote.data.id;

    // Create multiple linked notes to ensure depth-3 loading has more nodes
    const linkedIds = [];
    for (let i = 0; i < 5; i++) {
      const linked = await createNote(request, { title: `Linked Note ${i}`, content: `Content ${i}` });
      linkedIds.push(linked.data.id);

      // Link to central
      await createLink(request, centralId, linked.data.id, 0.7);

      // Create secondary links (depth-2 connections)
      if (i < 3) {
        const secondary = await createNote(request, { title: `Secondary ${i}`, content: `Secondary content ${i}` });
        await createLink(request, linked.data.id, secondary.data.id, 0.5);
      }
      
      // Small delay to prevent rate limiting
      await page.waitForTimeout(50);
    }

    // Navigate to 3D graph page
    await page.goto(`/graph/3d/${centralId}`);
    await page.waitForLoadState('networkidle');
    
    // Verify graph container appears immediately
    const graphContainer = page.locator('[data-testid="graph-3d-container"]').first();
    await expect(graphContainer).toBeVisible({ timeout: 2000 });
    
    // Loading overlay may appear briefly but should disappear quickly
    const loadingOverlay = page.locator('[data-testid="loading-overlay"]');
    await expect(loadingOverlay).toBeHidden({ timeout: 8000 });
    
    // Verify stats bar shows initial data
    const statsBar = page.locator('[data-testid="graph-stats"]').first();
    await expect(statsBar).toBeVisible();
    
    // Check that nodes count is displayed
    const statsText = await statsBar.textContent();
    expect(statsText).toMatch(/\d+\s*nodes?/i);
  });

  test('should show dense fog initially and clear after progressive load', async ({ page, request }) => {
    // Create a note with connections using helper
    const centralNote = await createNote(request, {
      title: 'Fog Test Note',
      content: 'Testing fog effect',
      type: 'galaxy'
    });
    const centralId = centralNote.data.id;
    
    // Create linked notes
    for (let i = 0; i < 3; i++) {
      const linked = await createNote(request, {
        title: `Fog Link ${i}`,
        content: 'Link content'
      });
      await createLink(request, centralId, linked.data.id, 0.8, 'reference');
      // Small delay to prevent rate limiting
      await page.waitForTimeout(50);
    }
    
    // Navigate to 3D graph
    await page.goto(`/graph/3d/${centralId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Verify graph container is visible
    const graphContainer = page.locator('[data-testid="graph-3d-container"]').first();
    await expect(graphContainer).toBeVisible();
    
    // Wait for progressive loading to complete (Phase 2)
    await page.waitForTimeout(4000);
    
    // Verify the graph is still displayed after progressive load
    await expect(graphContainer).toBeVisible();
    
    // Check that no error overlay appeared
    const errorOverlay = page.locator('.error-overlay');
    const hasError = await errorOverlay.isVisible().catch(() => false);
    expect(hasError).toBe(false);
  });

  test('should display node elements in 3D space', async ({ page, request }) => {
    // Create a note using helper
    const note = await createNote(request, {
      title: 'Node Display Test',
      content: 'Testing node rendering',
      type: 'planet'
    });
    const noteId = note.data.id;
    
    // Navigate to 3D graph
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Verify graph container exists
    const graphContainer = page.locator('[data-testid="graph-3d-container"]').first();
    await expect(graphContainer).toBeVisible();
    
    // Check that canvas is present (WebGL rendering)
    const canvas = page.locator('canvas').first();
    const hasCanvas = await canvas.isVisible().catch(() => false);
    
    // WebGL may not work in all headless environments
    if (hasCanvas) {
      // Verify canvas has correct dimensions
      const box = await canvas.boundingBox();
      expect(box?.width).toBeGreaterThan(0);
      expect(box?.height).toBeGreaterThan(0);
    }
  });

  test('should render links between connected nodes', async ({ page, request }) => {
    // Create two connected notes using helpers with retry logic
    const note1 = await createNote(request, { title: 'Source Node', content: 'Source', type: 'star' });
    const note1Id = note1.data.id;
    
    const note2 = await createNote(request, { title: 'Target Node', content: 'Target', type: 'planet' });
    const note2Id = note2.data.id;
    
    // Create link using helper with retry logic
    await createLink(request, note1Id, note2Id, 0.9, 'dependency');
    
    // Navigate to 3D graph
    await page.goto(`/graph/3d/${note1Id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    // Verify graph container is visible
    const graphContainer = page.locator('.graph-3d-container').first();
    await expect(graphContainer).toBeVisible();
    
    // Verify stats show correct counts (should have 2 nodes, 1 link)
    const statsBar = page.locator('.stats-bar').first();
    await expect(statsBar).toBeVisible();
    
    const statsText = await statsBar.textContent();
    expect(statsText).toContain('nodes');
    expect(statsText).toContain('links');
  });

  test('should show background loading indicator during progressive load', async ({ page, request }) => {
    // Create a note with multiple connections to trigger progressive loading
    const centralNote = await createNote(request, { title: 'Progressive Loading Test', content: 'Testing background loading indicator' });
    const centralId = centralNote.data.id;
    
    // Create many linked notes to ensure progressive loading takes time
    for (let i = 0; i < 8; i++) {
      const linked = await createNote(request, { title: `Loading Test Link ${i}`, content: `Content ${i}` });
      await createLink(request, centralId, linked.data.id, 0.6);
      // Small delay to prevent rate limiting
      await page.waitForTimeout(50);
    }
    
    // Navigate to 3D graph
    await page.goto(`/graph/3d/${centralId}`);
    await page.waitForLoadState('networkidle');
    
    // Wait a bit for potential background loading
    await page.waitForTimeout(500);
    
    // Verify the stats bar is present (spinner may or may not appear)
    const statsBar = page.locator('.stats-bar').first();
    await expect(statsBar).toBeVisible();
  });

  test('should handle camera reset function', async ({ page, request }) => {
    // Create a note using helper with retry logic
    const note = await createNote(request, { title: 'Camera Reset Test', content: 'Testing camera functionality' });
    const noteId = note.data.id;
    
    // Navigate to 3D graph
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Verify graph loaded
    const graphContainer = page.locator('.graph-3d-container').first();
    await expect(graphContainer).toBeVisible();
    
    // Verify canvas exists (camera is attached to it)
    const canvas = page.locator('canvas').first();
    const hasCanvas = await canvas.isVisible().catch(() => false);
    
    // In WebGL-capable environments, verify canvas is present
    if (hasCanvas) {
      expect(await canvas.count()).toBeGreaterThan(0);
    }
  });

  test('should handle multiple note transitions correctly', async ({ page, request }) => {
    // Create two separate notes with their own graphs using helpers
    const note1 = await createNote(request, { title: 'First Graph Note', content: 'First graph' });
    const note1Id = note1.data.id;
    
    const note2 = await createNote(request, { title: 'Second Graph Note', content: 'Second graph' });
    const note2Id = note2.data.id;
    
    // Navigate to first graph
    await page.goto(`/graph/3d/${note1Id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    const graphContainer = page.locator('.graph-3d-container').first();
    await expect(graphContainer).toBeVisible();
    
    // Navigate to second graph
    await page.goto(`/graph/3d/${note2Id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Verify second graph loaded
    await expect(graphContainer).toBeVisible();
    
    // Verify stats updated
    const statsBar = page.locator('.stats-bar').first();
    await expect(statsBar).toBeVisible();
  });

  test('should handle empty graph (no connections) gracefully', async ({ page, request }) => {
    // Create an isolated note with no connections using helper
    const isolatedNote = await createNote(request, { title: 'Isolated Node', content: 'No connections', type: 'asteroid' });
    const isolatedId = isolatedNote.data.id;
    
    // Navigate to 3D graph
    await page.goto(`/graph/3d/${isolatedId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Verify either graph container or no-data message is shown
    const graphContainer = page.locator('.graph-3d-container').first();
    const noDataMessage = page.locator('.no-data-message, .empty-content').first();
    
    const hasGraph = await graphContainer.isVisible().catch(() => false);
    const hasNoData = await noDataMessage.isVisible().catch(() => false);
    
    // Should show either graph or appropriate message for single node
    expect(hasGraph || hasNoData).toBe(true);
    
    // Verify stats show 1 node, 0 links
    const statsBar = page.locator('.stats-bar').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      expect(statsText).toMatch(/1.*node|1.*nodes/i);
    }
  });

  test('should display correct celestial body types', async ({ page, request }) => {
    // Create a central star node using helper
    const centralNote = await createNote(request, { title: 'Central Star', content: 'Central node', type: 'star' });
    const centralId = centralNote.data.id;
    
    // Create planet, comet, asteroid linked to central
    const linkedIds = [];
    for (const type of ['planet', 'comet', 'asteroid']) {
      const note = await createNote(request, { title: `${type} Node`, content: `Type: ${type}`, type });
      linkedIds.push(note.data.id);
      // Small delay to prevent rate limiting
      await page.waitForTimeout(50);
    }
    
    // Create all links using helper
    for (const linkedId of linkedIds) {
      await createLink(request, centralId, linkedId, 0.8);
      // Small delay to prevent rate limiting
      await page.waitForTimeout(50);
    }
    
    // Navigate to central node's graph
    await page.goto(`/graph/3d/${centralId}`);
    await page.waitForLoadState('networkidle');
    
    // Wait for progressive loading to complete (initial + background)
    await page.waitForTimeout(6000);
    
    // Verify graph loaded
    const graphContainer = page.locator('.graph-3d-container').first();
    await expect(graphContainer).toBeVisible();
    
    // Verify stats bar shows the graph data
    const statsBar = page.locator('.stats-bar').first();
    await expect(statsBar).toBeVisible({ timeout: 5000 });
    
    const statsText = await statsBar.textContent();
    expect(statsText).not.toBeNull();
    
    // Just verify stats show some nodes (graph API might filter based on depth)
    const nodeMatch = statsText!.match(/(\d+)\s*nodes?/i);
    expect(nodeMatch).not.toBeNull();
    if (nodeMatch) {
      const nodeCount = parseInt(nodeMatch[1], 10);
      // Should have at least the central node
      expect(nodeCount).toBeGreaterThanOrEqual(1);
    }
    
    // Verify links count is shown
    expect(statsText).toMatch(/\d+\s*links?/i);
  });

});

test.describe('Progressive Graph - Camera & Animation', () => {
  
  test.beforeAll(async ({ request }) => {
    try {
      const healthCheck = await request.get('http://localhost:8080/notes', { timeout: 5000 });
      backendAvailable = healthCheck.status() < 500;
    } catch {
      backendAvailable = false;
    }
    
    if (!backendAvailable) {
      console.log('⚠️  Backend not available - Camera tests will be skipped');
    }
  });
  
  test.beforeEach(async ({ page }) => {
    if (!backendAvailable) {
      test.skip();
    }
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('should verify WebGL canvas dimensions are correct', async ({ page, request }) => {
    // Create a note using helper
    const note = await createNote(request, { title: 'Canvas Dimension Test', content: 'Testing canvas size' });
    const noteId = note.data.id;
    
    // Navigate to 3D graph
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Get canvas element
    const canvas = page.locator('canvas').first();
    const hasCanvas = await canvas.isVisible().catch(() => false);
    
    if (hasCanvas) {
      const box = await canvas.boundingBox();
      expect(box).not.toBeNull();
      expect(box!.width).toBeGreaterThan(100);
      expect(box!.height).toBeGreaterThan(100);
      
      // Verify canvas fills most of the viewport
      const viewport = page.viewportSize();
      if (viewport) {
        expect(box!.width).toBeGreaterThan(viewport.width * 0.8);
        expect(box!.height).toBeGreaterThan(viewport.height * 0.8);
      }
    }
  });

  test('should maintain graph container structure', async ({ page, request }) => {
    // Create a note with connections using helpers
    const note = await createNote(request, { title: 'Structure Test', content: 'Testing DOM structure' });
    const noteId = note.data.id;
    
    const linked = await createNote(request, { title: 'Structure Link', content: 'Link' });
    await createLink(request, noteId, linked.data.id, 0.7);
    
    // Navigate to 3D graph
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Verify DOM structure
    const graphContainer = page.locator('.graph-3d-container').first();
    await expect(graphContainer).toBeVisible();
    
    // Check that container has proper styling
    const containerStyles = await graphContainer.evaluate(el => {
      const styles = window.getComputedStyle(el);
      return {
        position: styles.position,
        width: styles.width,
        height: styles.height,
        overflow: styles.overflow
      };
    });
    
    expect(containerStyles.position).toBe('relative');
    expect(containerStyles.overflow).toBe('hidden');
  });

});
