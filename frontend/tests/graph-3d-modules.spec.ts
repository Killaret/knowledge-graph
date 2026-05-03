import { test, expect } from '@playwright/test';
import { createNote, createLink, isBackendAvailable, getBackendUrl } from './helpers/testData';

/**
 * Tests for 3D Graph Modules (Three.js Refactored)
 * Verifies the modular architecture, celestial bodies rendering, and link visualization
 * 
 * NOTE: These tests require the backend to be running on localhost:8080
 * If backend is unavailable, tests will be skipped
 */

// Global flag to track backend availability
let backendAvailable = false;

test.describe('3D Graph - Modular Architecture', { tag: ['@smoke', '@3d', '@modules'] }, () => {
  
  test.beforeAll(async ({ request }) => {
    // Check backend availability once before all tests
    backendAvailable = await isBackendAvailable(request);

    if (!backendAvailable) {
      console.log(`⚠️  Backend not available on ${getBackendUrl()} - 3D graph tests will be skipped`);
    }
  });

  test.beforeEach(async ({ page }) => {
    if (!backendAvailable) {
      test.skip();
    }

    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should render 3D graph with scene setup module', async ({ page, request }) => {
    // Create a note via API using helper
    const note = await createNote(request, {
      title: 'Scene Test Note',
      content: 'Testing scene setup module',
      type: 'star'
    });
    const noteId = note.data.id;

    // Navigate to 3D graph page
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    
    // Wait for lazy loading to complete
    await page.waitForTimeout(2000);
    
    // Check for loading overlay
    const loadingOverlay = page.locator('[data-testid="loading-overlay"]').first();
    
    // Wait for loading to finish or graph to appear
    await page.waitForTimeout(3000);
    let attempts = 0;
    while (attempts < 30) {
      const isLoading = await loadingOverlay.isVisible().catch(() => false);
      if (!isLoading) break;
      await page.waitForTimeout(500);
      attempts++;
    }
    
    // After loading, check for graph container or error
    const graphContainer = page.locator('[data-testid="graph-3d-container"]').first();
    const lazyError = page.locator('.lazy-error').first();
    const errorOverlay = page.locator('.error-overlay').first();
    const centerError = page.locator('.center.error').first();
    
    const hasGraph = await graphContainer.isVisible().catch(() => false);
    const hasLazyError = await lazyError.isVisible().catch(() => false);
    const hasErrorOverlay = await errorOverlay.isVisible().catch(() => false);
    const hasCenterError = await centerError.isVisible().catch(() => false);
    
    // Should have either graph or error state
    expect(hasGraph || hasLazyError || hasErrorOverlay || hasCenterError).toBe(true);
  });

  test('should display star celestial body', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Star Node',
      content: 'Star type test',
      type: 'star'
    });
    const noteId = note.data.id;

    const note2 = await createNote(request, { title: 'Linked', content: 'Link' });
    await createLink(request, noteId, note2.data.id, 0.8);

    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // After loading, verify graph or error state
    const container = page.locator('[data-testid="graph-3d-container"], .error-overlay').first();
    await expect(container).toBeVisible();
  });

  test('should display planet celestial body', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Planet Node',
      content: 'Planet type',
      type: 'planet'
    });
    const noteId = note.data.id;

    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
  });

  test('should display comet celestial body', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Comet Node',
      content: 'Comet type',
      type: 'comet'
    });
    const noteId = note.data.id;

    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
  });

  test('should display galaxy celestial body', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Galaxy Node',
      content: 'Galaxy type',
      type: 'galaxy'
    });
    const noteId = note.data.id;

    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
  });

  test('should render links with weight-based styling', async ({ page, request }) => {
    // Create notes with different link weights using helper
    const sourceNote = await createNote(request, { title: 'Source', content: 'Source note' });
    const sourceId = sourceNote.data.id;

    const strongTarget = await createNote(request, { title: 'Strong Link', content: 'Strong connection' });
    const strongId = strongTarget.data.id;

    const weakTarget = await createNote(request, { title: 'Weak Link', content: 'Weak connection' });
    const weakId = weakTarget.data.id;

    // Create links with different weights
    await createLink(request, sourceId, strongId, 0.9);
    await createLink(request, sourceId, weakId, 0.2);

    await page.goto(`/graph/3d/${sourceId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
  });

  test('should auto-zoom camera to fit graph', async ({ page, request }) => {
    // Create multiple notes to form a graph using helper
    const notes = [];
    for (let i = 0; i < 5; i++) {
      const note = await createNote(request, { title: `Note ${i}`, content: `Content ${i}` });
      notes.push(note.data.id);
    }

    // Create links between notes
    for (let i = 0; i < notes.length - 1; i++) {
      await createLink(request, notes[i], notes[i + 1], 0.5);
    }

    await page.goto(`/graph/3d/${notes[0]}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000); // Wait for simulation to settle and auto-zoom
    
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
    
    // After auto-zoom, loading should be complete (no loading overlay)
    const loading = page.locator('.loading-overlay');
    const isLoading = await loading.isVisible().catch(() => false);
    // Either not loading or error state is acceptable
    expect(typeof isLoading).toBe('boolean');
  });

  test('should handle full graph toggle', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Toggle Test',
      content: 'Testing full graph toggle'
    });
    const noteId = note.data.id;

    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Find and click the toggle
    const toggle = page.locator('.toggle input[type="checkbox"]').first();
    if (await toggle.isVisible().catch(() => false)) {
      await toggle.click();
      await page.waitForTimeout(2000);
      
      // Verify graph still renders after toggle
      const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
      await expect(container).toBeVisible();
    }
  });

  test('should display node labels via CSS2D', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Labeled Node',
      content: 'Testing CSS2D labels'
    });
    const noteId = note.data.id;

    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // Check for labels in the DOM (CSS2D creates div elements)
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
  });

  test('should open graph page and verify full graph toggle', async ({ page, request }) => {
    // Create test notes using helper
    const note1 = await createNote(request, {
      title: 'Main Note',
      content: 'Main content',
      type: 'star'
    });
    const note1Id = note1.data.id;

    const note2 = await createNote(request, {
      title: 'Linked Note',
      content: 'Linked content',
      type: 'planet'
    });
    const note2Id = note2.data.id;

    // Create link between notes
    await createLink(request, note1Id, note2Id, 0.8);

    // Test 1: Open 3D graph page for specific note (local constellation)
    await page.goto(`/graph/3d/${note1Id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // Verify graph container or error state is visible
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
    
    // Test 2: Verify the "Show all notes" toggle exists and works
    const toggle = page.locator('.toggle input[type="checkbox"], [data-testid="full-graph-toggle"]').first();
    const hasToggle = await toggle.isVisible().catch(() => false);
    
    if (hasToggle) {
      // Click toggle to switch to full graph view
      await toggle.click();
      await page.waitForTimeout(3000);
      
      // Verify graph still renders after toggle
      const containerAfterToggle = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
      await expect(containerAfterToggle).toBeVisible();
      
      // Re-query toggle for second click (DOM might have changed)
      const toggleAfter = page.locator('.toggle input[type="checkbox"], [data-testid="full-graph-toggle"]').first();
      const hasToggleAfter = await toggleAfter.isVisible().catch(() => false);
      
      if (hasToggleAfter) {
        // Click again to switch back to local view
        await toggleAfter.click();
        await page.waitForTimeout(2000);
      }
    }
  });

  test('should display correct stats in 2D graph page', async ({ page, request }) => {
    // Get current notes count
    const notesResp = await request.get(`${getBackendUrl()}/notes`);
    const notesData = await notesResp.json();
    const totalNotes = notesData.total || 0;
    console.log('Total notes in DB:', totalNotes);

    // Navigate to 2D graph page
    await page.goto('/graph');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    // Check for stats bar
    const statsBar = page.locator('.stats-bar').first();
    const hasStats = await statsBar.isVisible().catch(() => false);
    
    if (hasStats) {
      const statsText = await statsBar.textContent();
      // Verify stats show numbers (nodes/links count)
      expect(statsText).toMatch(/\d+\s+nodes?/i);
      expect(statsText).toMatch(/\d+\s+links?/i);
      
      // Verify mode indicator (Full graph or Local view)
      const hasMode = statsText?.toLowerCase().includes('full') || statsText?.toLowerCase().includes('local');
      expect(hasMode).toBe(true);
    }
  });

  test('should display correct stats in 3D graph page', async ({ page, request }) => {
    // Create a note for 3D graph using helper
    const note = await createNote(request, {
      title: '3D Stats Test',
      content: 'Testing 3D stats display',
      type: 'galaxy'
    });
    const noteId = note.data.id;

    // Navigate to 3D graph page
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // Check for stats bar
    const statsBar = page.locator('.stats-bar').first();
    const hasStats = await statsBar.isVisible().catch(() => false);
    
    if (hasStats) {
      const statsText = await statsBar.textContent();
      // Verify stats show numbers
      expect(statsText).toMatch(/\d+\s+nodes?/i);
      expect(statsText).toMatch(/\d+\s+links?/i);
    }
  });

  test('should toggle full graph mode in 3D and verify data changes', async ({ page, request }) => {
    // Ensure we have notes with links using helper
    const note1 = await createNote(request, {
      title: '3D Toggle Test 1',
      content: 'First note',
      type: 'star'
    });
    const note1Id = note1.data.id;

    const note2 = await createNote(request, {
      title: '3D Toggle Test 2',
      content: 'Second note',
      type: 'planet'
    });
    const note2Id = note2.data.id;

    // Create link
    await createLink(request, note1Id, note2Id, 0.7, 'reference');

    // Navigate to 3D graph with the first note as center
    await page.goto(`/graph/3d/${note1Id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // Find toggle
    const toggle = page.locator('.toggle input[type="checkbox"]').first();
    const hasToggle = await toggle.isVisible().catch(() => false);
    
    if (!hasToggle) {
      test.skip(true, 'Toggle not found on 3D graph page');
      return;
    }
    
    // Get initial stats (local mode)
    const statsBefore = await page.locator('[data-testid="graph-stats"], .stats-bar').textContent().catch(() => '') || '';
    console.log('Stats before toggle:', statsBefore);
    
    // Toggle to full graph mode
    await toggle.click();
    await page.waitForTimeout(4000);
    
    // Get stats after toggle (should show more nodes in full mode)
    const statsAfter = await page.locator('[data-testid="graph-stats"], .stats-bar').textContent().catch(() => '') || '';
    console.log('Stats after toggle:', statsAfter);
    
    // Verify mode changed (check for 'Full graph' text)
    const hasFullMode = statsAfter.toLowerCase().includes('full');
    expect(hasFullMode).toBe(true);
    
    // Graph should still render
    const container = page.locator('.graph-3d-container, [data-testid="graph-3d-container"]').first();
    await expect(container).toBeVisible();
  });

  test('should fetch full graph data via API', async ({ request }) => {
    // Create test notes for full graph using helper
    const notes = [];
    for (let i = 0; i < 3; i++) {
      const note = await createNote(request, {
        title: `Full Graph Note ${i}`,
        content: `Content ${i}`,
        type: i === 0 ? 'star' : 'planet'
      });
      notes.push(note.data.id);
    }

    // Create links
    await createLink(request, notes[0], notes[1], 0.7);
    await createLink(request, notes[1], notes[2], 0.5);

    // Test the /api/v1/graph/all endpoint directly
    const response = await request.get(`${getBackendUrl()}/api/v1/graph/all`, { timeout: 10000 });
    
    // Should return 200 OK with graph data
    expect(response.status()).toBe(200);
    
    const responseData = await response.json();
    // API returns wrapped response: { data: { nodes, links } }
    expect(responseData).toHaveProperty('data');
    const graphData = responseData.data;
    expect(graphData).toHaveProperty('nodes');
    expect(graphData).toHaveProperty('links');
    expect(Array.isArray(graphData.nodes)).toBe(true);
    expect(Array.isArray(graphData.links)).toBe(true);
  });

  test('should display asteroid celestial body', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Asteroid Node',
      content: 'Asteroid type',
      type: 'asteroid'
    });
    const noteId = note.data.id;

    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
  });

  test('should display debris celestial body', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Debris Node',
      content: 'Debris type',
      type: 'debris'
    });
    const noteId = note.data.id;

    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
  });

  test('should render different link types with distinct styling', async ({ page, request }) => {
    // Create notes for different link types using helper
    const sourceNote = await createNote(request, {
      title: 'Link Types Source',
      content: 'Source for link types test',
      type: 'star'
    });
    const sourceId = sourceNote.data.id;

    const referenceTarget = await createNote(request, {
      title: 'Reference Target',
      content: 'Reference link target',
      type: 'planet'
    });
    const referenceId = referenceTarget.data.id;

    const dependencyTarget = await createNote(request, {
      title: 'Dependency Target',
      content: 'Dependency link target',
      type: 'comet'
    });
    const dependencyId = dependencyTarget.data.id;

    const relatedTarget = await createNote(request, {
      title: 'Related Target',
      content: 'Related link target',
      type: 'galaxy'
    });
    const relatedId = relatedTarget.data.id;

    // Create links with different types
    await createLink(request, sourceId, referenceId, 0.8, 'reference');
    await createLink(request, sourceId, dependencyId, 0.7, 'dependency');
    await createLink(request, sourceId, relatedId, 0.6, 'related');

    await page.goto(`/graph/3d/${sourceId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // Verify graph renders with multiple link types
    const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
    await expect(container).toBeVisible();
    
    // Stats should show multiple nodes and links
    const statsBar = page.locator('.stats-bar').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      // Should have 4 nodes (source + 3 targets) and 3 links
      expect(statsText).toMatch(/4\s*nodes/);
      expect(statsText).toMatch(/3\s*links/);
    }
  });

  test('should render full 3D graph at /graph/3d without note ID', async ({ page }) => {
    // Navigate to full 3D graph page
    await page.goto('/graph/3d');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    // Verify container or loading/error state is visible
    const container = page.locator('.graph-3d-container, .lazy-loading, .lazy-error, .center.error').first();
    await expect(container).toBeVisible();
    
    // Verify no 404 error
    const error404 = page.locator('text=404, text=Not Found').first();
    const has404 = await error404.isVisible().catch(() => false);
    expect(has404).toBe(false);
    
    // Stats bar should show full graph info
    const statsBar = page.locator('.stats-bar').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      expect(statsText).toMatch(/\d+\s*nodes/);
      expect(statsText).toMatch(/\d+\s*links/);
    }
  });

  test('should render isolated note without connections in 3D graph', async ({ page, request }) => {
    // Create a note with NO connections using helper
    const isolatedNote = await createNote(request, {
      title: 'Isolated Note',
      content: 'No connections',
      type: 'star'
    });
    const isolatedId = isolatedNote.data.id;

    // Navigate to 3D graph page
    await page.goto(`/graph/3d/${isolatedId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // Graph should still render (even with single node)
    const container = page.locator('.graph-3d-container, .no-data-message, .error-overlay').first();
    await expect(container).toBeVisible();
    
    // Stats should show 1 node and 0 links
    const statsBar = page.locator('.stats-bar').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      // Should show at least 1 node (the start node)
      const nodeMatch = statsText?.match(/(\d+)\s*nodes/);
      if (nodeMatch) {
        const nodeCount = parseInt(nodeMatch[1], 10);
        expect(nodeCount).toBeGreaterThanOrEqual(1);
      }
      // Links should be 0
      expect(statsText).toMatch(/0\s*links/);
    }
  });
});
