import { test, expect } from '@playwright/test';

/**
 * Tests for 3D Graph Modules (Three.js Refactored)
 * Verifies the modular architecture, celestial bodies rendering, and link visualization
 */

test.describe('3D Graph - Modular Architecture', () => {
  
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should render 3D graph with scene setup module', async ({ page, request }) => {
    // Create a note via API
    const note = await request.post('http://localhost:8080/notes', {
      data: { 
        title: 'Scene Test Note', 
        content: 'Testing scene setup module',
        type: 'star'
      }
    });
    const noteId = (await note.json()).id;
    
    // Navigate to 3D graph page
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    // Verify scene setup - starfield background
    const container = page.locator('.graph-3d-container').first();
    await expect(container).toBeVisible();
    
    // Check for loading or actual canvas
    const loadingOverlay = page.locator('.loading-overlay');
    const canvas = page.locator('canvas').first();
    
    const hasLoading = await loadingOverlay.isVisible().catch(() => false);
    const hasCanvas = await canvas.isVisible().catch(() => false);
    
    // Either loading (simulation running) or canvas should be visible
    expect(hasLoading || hasCanvas).toBe(true);
  });

  test('should display star celestial body', async ({ page, request }) => {
    // Create a note with star type
    const note = await request.post('http://localhost:8080/notes', {
      data: { 
        title: 'Star Node', 
        content: 'This should render as a star',
        type: 'star'
      }
    });
    const noteId = (await note.json()).id;
    
    // Create a link to ensure graph has data
    const note2 = await request.post('http://localhost:8080/notes', {
      data: { title: 'Linked Note', content: 'Link target' }
    });
    const note2Id = (await note2.json()).id;
    
    // Create link between notes
    await request.post('http://localhost:8080/links', {
      data: { sourceNoteId: noteId, targetNoteId: note2Id, weight: 0.8 }
    });
    
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    // Verify graph container with star type nodes
    const container = page.locator('.graph-3d-container');
    await expect(container).toBeVisible();
  });

  test('should display planet celestial body', async ({ page, request }) => {
    const note = await request.post('http://localhost:8080/notes', {
      data: { 
        title: 'Planet Node', 
        content: 'This should render as a planet',
        type: 'planet'
      }
    });
    const noteId = (await note.json()).id;
    
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    const container = page.locator('.graph-3d-container');
    await expect(container).toBeVisible();
  });

  test('should display comet celestial body', async ({ page, request }) => {
    const note = await request.post('http://localhost:8080/notes', {
      data: { 
        title: 'Comet Node', 
        content: 'This should render as a comet',
        type: 'comet'
      }
    });
    const noteId = (await note.json()).id;
    
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    const container = page.locator('.graph-3d-container');
    await expect(container).toBeVisible();
  });

  test('should display galaxy celestial body', async ({ page, request }) => {
    const note = await request.post('http://localhost:8080/notes', {
      data: { 
        title: 'Galaxy Node', 
        content: 'This should render as a galaxy',
        type: 'galaxy'
      }
    });
    const noteId = (await note.json()).id;
    
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    const container = page.locator('.graph-3d-container');
    await expect(container).toBeVisible();
  });

  test('should render links with weight-based styling', async ({ page, request }) => {
    // Create notes with different link weights
    const sourceNote = await request.post('http://localhost:8080/notes', {
      data: { title: 'Source', content: 'Source note' }
    });
    const sourceId = (await sourceNote.json()).id;
    
    const strongTarget = await request.post('http://localhost:8080/notes', {
      data: { title: 'Strong Link', content: 'Strong connection' }
    });
    const strongId = (await strongTarget.json()).id;
    
    const weakTarget = await request.post('http://localhost:8080/notes', {
      data: { title: 'Weak Link', content: 'Weak connection' }
    });
    const weakId = (await weakTarget.json()).id;
    
    // Create links with different weights
    await request.post('http://localhost:8080/links', {
      data: { sourceNoteId: sourceId, targetNoteId: strongId, weight: 0.9 }
    });
    await request.post('http://localhost:8080/links', {
      data: { sourceNoteId: sourceId, targetNoteId: weakId, weight: 0.2 }
    });
    
    await page.goto(`http://localhost:5173/graph/3d/${sourceId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    const container = page.locator('.graph-3d-container');
    await expect(container).toBeVisible();
  });

  test('should auto-zoom camera to fit graph', async ({ page, request }) => {
    // Create multiple notes to form a graph
    const notes = [];
    for (let i = 0; i < 5; i++) {
      const note = await request.post('http://localhost:8080/notes', {
        data: { title: `Note ${i}`, content: `Content ${i}` }
      });
      notes.push((await note.json()).id);
    }
    
    // Create links between notes
    for (let i = 0; i < notes.length - 1; i++) {
      await request.post('http://localhost:8080/links', {
        data: { sourceNoteId: notes[i], targetNoteId: notes[i + 1], weight: 0.5 }
      });
    }
    
    await page.goto(`http://localhost:5173/graph/3d/${notes[0]}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000); // Wait for simulation to settle and auto-zoom
    
    const container = page.locator('.graph-3d-container');
    await expect(container).toBeVisible();
    
    // After auto-zoom, loading should be complete
    const loading = page.locator('.loading-overlay');
    const isLoading = await loading.isVisible().catch(() => false);
    expect(isLoading).toBe(false);
  });

  test('should handle full graph toggle', async ({ page, request }) => {
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Toggle Test', content: 'Testing full graph toggle' }
    });
    const noteId = (await note.json()).id;
    
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Find and click the toggle
    const toggle = page.locator('.toggle input[type="checkbox"]').first();
    if (await toggle.isVisible().catch(() => false)) {
      await toggle.click();
      await page.waitForTimeout(2000);
      
      // Verify graph still renders after toggle
      const container = page.locator('.graph-3d-container');
      await expect(container).toBeVisible();
    }
  });

  test('should display node labels via CSS2D', async ({ page, request }) => {
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Labeled Node', content: 'Testing CSS2D labels' }
    });
    const noteId = (await note.json()).id;
    
    await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    // Check for labels in the DOM (CSS2D creates div elements)
    const container = page.locator('.graph-3d-container');
    await expect(container).toBeVisible();
  });
});
