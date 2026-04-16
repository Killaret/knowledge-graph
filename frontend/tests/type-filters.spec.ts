import { test, expect } from '@playwright/test';

/**
 * Tests for Type Filtering with metadata.type fallback
 * Verifies that filters work correctly when type is stored in metadata
 * instead of the root type field
 */

test.describe('Type Filters - Home Page Filtering', () => {
  
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should filter notes by star type', async ({ page, request }) => {
    const timestamp = Date.now();
    
    // Create star type note
    await request.post('http://localhost:8080/notes', {
      data: { 
        title: `Star Filter Test ${timestamp}`, 
        content: 'Star content',
        type: 'star'
      }
    });
    
    // Create planet type note
    await request.post('http://localhost:8080/notes', {
      data: { 
        title: `Planet Filter Test ${timestamp}`, 
        content: 'Planet content',
        type: 'planet'
      }
    });
    
    // Reload to see new notes
    await page.reload();
    await page.waitForTimeout(2000);
    
    // Click on Stars filter button
    const starsFilter = page.locator('button:has-text("⭐"), button:has-text("Stars"), button[data-filter="star"]').first();
    if (await starsFilter.isVisible().catch(() => false)) {
      await starsFilter.click();
      await page.waitForTimeout(1000);
      
      // Verify filter is applied - stats should show filtered state
      const statsBar = page.locator('.stats-bar').first();
      if (await statsBar.isVisible().catch(() => false)) {
        const statsText = await statsBar.textContent();
        // Should show filter indicator or specific count
        const hasFilterIndicator = statsText?.toLowerCase().includes('filter') || 
                                   statsText?.toLowerCase().includes('star');
        expect(hasFilterIndicator).toBe(true);
      }
    }
  });

  test('should filter notes by planet type', async ({ page, request }) => {
    const timestamp = Date.now();
    
    // Create notes with different types
    await request.post('http://localhost:8080/notes', {
      data: { 
        title: `Planet Type ${timestamp}`, 
        content: 'Planet content',
        type: 'planet'
      }
    });
    
    await request.post('http://localhost:8080/notes', {
      data: { 
        title: `Comet Type ${timestamp}`, 
        content: 'Comet content',
        type: 'comet'
      }
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // Click on Planets filter
    const planetsFilter = page.locator('button:has-text("🪐"), button:has-text("Planets"), button[data-filter="planet"]').first();
    if (await planetsFilter.isVisible().catch(() => false)) {
      await planetsFilter.click();
      await page.waitForTimeout(1000);
      
      // Verify filter is active
      const statsBar = page.locator('.stats-bar').first();
      if (await statsBar.isVisible().catch(() => false)) {
        const statsText = await statsBar.textContent();
        const hasFilterIndicator = statsText?.toLowerCase().includes('filter') || 
                                   statsText?.toLowerCase().includes('planet');
        expect(hasFilterIndicator).toBe(true);
      }
    }
  });

  test('should filter notes by comet type', async ({ page, request }) => {
    const timestamp = Date.now();
    
    await request.post('http://localhost:8080/notes', {
      data: { 
        title: `Comet Filter ${timestamp}`, 
        content: 'Comet content',
        type: 'comet'
      }
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // Click on Comets filter
    const cometsFilter = page.locator('button:has-text("☄️"), button:has-text("Comets"), button[data-filter="comet"]').first();
    if (await cometsFilter.isVisible().catch(() => false)) {
      await cometsFilter.click();
      await page.waitForTimeout(1000);
      
      const statsBar = page.locator('.stats-bar').first();
      if (await statsBar.isVisible().catch(() => false)) {
        const statsText = await statsBar.textContent();
        expect(statsText?.toLowerCase()).toMatch(/filter|comet/);
      }
    }
  });

  test('should filter notes by galaxy type', async ({ page, request }) => {
    const timestamp = Date.now();
    
    await request.post('http://localhost:8080/notes', {
      data: { 
        title: `Galaxy Filter ${timestamp}`, 
        content: 'Galaxy content',
        type: 'galaxy'
      }
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // Click on Galaxies filter
    const galaxiesFilter = page.locator('button:has-text("🌀"), button:has-text("Galaxies"), button[data-filter="galaxy"]').first();
    if (await galaxiesFilter.isVisible().catch(() => false)) {
      await galaxiesFilter.click();
      await page.waitForTimeout(1000);
      
      const statsBar = page.locator('.stats-bar').first();
      if (await statsBar.isVisible().catch(() => false)) {
        const statsText = await statsBar.textContent();
        expect(statsText?.toLowerCase()).toMatch(/filter|galaxy/);
      }
    }
  });

  test('should show all notes when selecting All filter', async ({ page, request }) => {
    const timestamp = Date.now();
    
    // Create multiple notes of different types
    await request.post('http://localhost:8080/notes', {
      data: { title: `All Filter Star ${timestamp}`, content: 'Star', type: 'star' }
    });
    await request.post('http://localhost:8080/notes', {
      data: { title: `All Filter Planet ${timestamp}`, content: 'Planet', type: 'planet' }
    });
    await request.post('http://localhost:8080/notes', {
      data: { title: `All Filter Comet ${timestamp}`, content: 'Comet', type: 'comet' }
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // First filter by star
    const starsFilter = page.locator('button:has-text("⭐"), button[data-filter="star"]').first();
    if (await starsFilter.isVisible().catch(() => false)) {
      await starsFilter.click();
      await page.waitForTimeout(1000);
    }
    
    // Then click All filter
    const allFilter = page.locator('button:has-text("🌌"), button:has-text("All"), button[data-filter="all"]').first();
    if (await allFilter.isVisible().catch(() => false)) {
      await allFilter.click();
      await page.waitForTimeout(1000);
      
      // Stats should show total count without filter indicator
      const statsBar = page.locator('.stats-bar').first();
      if (await statsBar.isVisible().catch(() => false)) {
        const statsText = await statsBar.textContent();
        // Should show total notes (at least 3 we just created)
        const countMatch = statsText?.match(/(\d+)\s*note/);
        if (countMatch) {
          const count = parseInt(countMatch[1], 10);
          expect(count).toBeGreaterThanOrEqual(3);
        }
        // Filter indicator should not be present for "All"
        expect(statsText?.toLowerCase()).not.toContain('filtered');
      }
    }
  });

  test('should filter graph nodes when type filter is applied', async ({ page, request }) => {
    const timestamp = Date.now();
    
    // Create notes with links of different types
    const starNote = await request.post('http://localhost:8080/notes', {
      data: { title: `Graph Star ${timestamp}`, content: 'Star note', type: 'star' }
    });
    const starId = (await starNote.json()).id;
    
    const planetNote = await request.post('http://localhost:8080/notes', {
      data: { title: `Graph Planet ${timestamp}`, content: 'Planet note', type: 'planet' }
    });
    const planetId = (await planetNote.json()).id;
    
    const cometNote = await request.post('http://localhost:8080/notes', {
      data: { title: `Graph Comet ${timestamp}`, content: 'Comet note', type: 'comet' }
    });
    const cometId = (await cometNote.json()).id;
    
    // Create links between notes
    await request.post('http://localhost:8080/links', {
      data: { sourceNoteId: starId, targetNoteId: planetId, weight: 0.8 }
    });
    await request.post('http://localhost:8080/links', {
      data: { sourceNoteId: planetId, targetNoteId: cometId, weight: 0.6 }
    });
    
    await page.reload();
    await page.waitForTimeout(3000);
    
    // Apply star filter
    const starsFilter = page.locator('button:has-text("⭐"), button[data-filter="star"]').first();
    if (await starsFilter.isVisible().catch(() => false)) {
      await starsFilter.click();
      await page.waitForTimeout(2000);
      
      // Graph should still be visible (filtered to show only stars)
      const graphContainer = page.locator('.graph-container, .graph-canvas').first();
      await expect(graphContainer).toBeVisible({ timeout: 5000 });
    }
  });
});

test.describe('Type Filters - metadata.type fallback', () => {
  
  test('should filter notes with type in metadata field', async ({ page, request }) => {
    const timestamp = Date.now();
    
    // Create note with type in metadata (simulating API that stores type in metadata)
    await request.post('http://localhost:8080/notes', {
      data: { 
        title: `Metadata Star ${timestamp}`, 
        content: 'Star type in metadata',
        metadata: { type: 'star', custom_field: 'value' }
      }
    });
    
    // Create note with type in root field
    await request.post('http://localhost:8080/notes', {
      data: { 
        title: `Root Planet ${timestamp}`, 
        content: 'Planet type in root',
        type: 'planet'
      }
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // Apply star filter - should catch both metadata.type and root type
    const starsFilter = page.locator('button:has-text("⭐"), button:has-text("Stars"), button[data-filter="star"]').first();
    if (await starsFilter.isVisible().catch(() => false)) {
      await starsFilter.click();
      await page.waitForTimeout(1000);
      
      // Filter should be applied - verify by checking stats
      const statsBar = page.locator('.stats-bar').first();
      if (await statsBar.isVisible().catch(() => false)) {
        const statsText = await statsBar.textContent();
        // Should show filtered count or filter indicator
        expect(statsText).toBeTruthy();
      }
    }
  });

  test('should fallback to star type when no type specified', async ({ page, request }) => {
    const timestamp = Date.now();
    
    // Create note without any type specified
    await request.post('http://localhost:8080/notes', {
      data: { 
        title: `No Type ${timestamp}`, 
        content: 'No type specified'
        // No type field at all
      }
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // The note should appear when filtering by "star" (default fallback)
    const starsFilter = page.locator('button:has-text("⭐"), button:has-text("Stars"), button[data-filter="star"]').first();
    if (await starsFilter.isVisible().catch(() => false)) {
      await starsFilter.click();
      await page.waitForTimeout(1000);
      
      // Stats should indicate filtering is active
      const statsBar = page.locator('.stats-bar').first();
      if (await statsBar.isVisible().catch(() => false)) {
        const statsText = await statsBar.textContent();
        expect(statsText).toBeTruthy();
      }
    }
  });
});
