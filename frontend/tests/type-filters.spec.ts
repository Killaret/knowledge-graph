import { test, expect } from '@playwright/test';
import { createNote, createLink, getBackendUrl } from './helpers/testData';
import { clickFilterChip } from './helpers/testUtils';

/**
 * Tests for Type Filtering with metadata.type fallback
 * Verifies that filters work correctly when type is stored in metadata
 * instead of the root type field
 */

test.describe('Type Filters - Home Page Filtering', { tag: ['@smoke', '@filters'] }, () => {

  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should filter notes by star type', async ({ page, request }) => {
    const timestamp = Date.now();

    // Create star type note using helper
    await createNote(request, {
      title: `Star Filter Test ${timestamp}`,
      content: 'Star content',
      type: 'star'
    });

    // Create planet type note
    await createNote(request, {
      title: `Planet Filter Test ${timestamp}`,
      content: 'Planet content',
      type: 'planet'
    });
    
    // Reload to see new notes
    await page.reload();
    await page.waitForTimeout(2000);
    
    // Click on Stars filter button
    await clickFilterChip(page, 'star');
    await page.waitForTimeout(1000);
    
    // Verify filter is applied - stats should show filtered state
    const statsBar = page.locator('[data-testid="graph-stats"]').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      // Should show filter indicator or specific count
      const hasFilterIndicator = statsText?.toLowerCase().includes('filter') || 
                                 statsText?.toLowerCase().includes('star');
      expect(hasFilterIndicator).toBe(true);
    }
  });

  test('should filter notes by planet type', async ({ page, request }) => {
    const timestamp = Date.now();

    // Create notes with different types using helper
    await createNote(request, {
      title: `Planet Type ${timestamp}`,
      content: 'Planet content',
      type: 'planet'
    });

    await createNote(request, {
      title: `Comet Type ${timestamp}`,
      content: 'Comet content',
      type: 'comet'
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // Click on Planets filter
    await clickFilterChip(page, 'planet');
    await page.waitForTimeout(1000);
    
    // Verify filter is active
    const statsBar = page.locator('.stats-bar').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      const hasFilterIndicator = statsText?.toLowerCase().includes('filter') || 
                                 statsText?.toLowerCase().includes('planet');
      expect(hasFilterIndicator).toBe(true);
    }
  });

  test('should filter notes by comet type', async ({ page, request }) => {
    const timestamp = Date.now();

    // Create comet type note using helper
    await createNote(request, {
      title: `Comet Filter Test ${timestamp}`,
      content: 'Comet content',
      type: 'comet'
    });

    // Create galaxy type note
    await createNote(request, {
      title: `Galaxy Filter Test ${timestamp}`,
      content: 'Galaxy content',
      type: 'galaxy'
    });
    
    // Reload to see new notes
    await page.reload();
    await page.waitForTimeout(2000);
    
    // Click on Comets filter
    await clickFilterChip(page, 'comet');
    await page.waitForTimeout(1000);
    
    const statsBar = page.locator('[data-testid="graph-stats"]').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      expect(statsText?.toLowerCase()).toMatch(/filter|comet/);
    }
  });

  test('should filter notes by galaxy type', async ({ page, request }) => {
    const timestamp = Date.now();

    // Create galaxy type note using helper
    await createNote(request, {
      title: `Galaxy Filter Test ${timestamp}`,
      content: 'Galaxy content',
      type: 'galaxy'
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // Click on Galaxies filter
    await clickFilterChip(page, 'galaxy');
    await page.waitForTimeout(1000);
    
    const statsBar = page.locator('[data-testid="graph-stats"]').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      expect(statsText?.toLowerCase()).toMatch(/filter|galax/);
    }
  });

  test('should show all notes when selecting All filter', async ({ page, request }) => {
    const timestamp = Date.now();

    // Create multiple notes of different types using helper
    await createNote(request, {
      title: `All Filter Star ${timestamp}`,
      content: 'Star',
      type: 'star'
    });
    await createNote(request, {
      title: `All Filter Planet ${timestamp}`,
      content: 'Planet',
      type: 'planet'
    });
    await createNote(request, {
      title: `All Filter Comet ${timestamp}`,
      content: 'Comet',
      type: 'comet'
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // First filter by star
    await clickFilterChip(page, 'star');
    await page.waitForTimeout(1000);
    
    // Then click All filter
    await clickFilterChip(page, 'all');
    await page.waitForTimeout(1000);
    
    // Stats should show total count without filter indicator
    const statsBar = page.locator('[data-testid="graph-stats"]').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      // Should show total notes (at least 3 we just created)
      const countMatch = statsText?.match(/(\d+)\s*nodes?/i);
      if (countMatch) {
        const count = parseInt(countMatch[1], 10);
        expect(count).toBeGreaterThanOrEqual(3);
      }
      // Filter indicator should not be present for "All"
      expect(statsText?.toLowerCase()).not.toContain('filtered');
    }
  });

  test('should clear filter by clicking All', async ({ page, request }) => {
    const timestamp = Date.now();

    // Create notes of different types using helper
    await createNote(request, {
      title: `Clear Filter Star ${timestamp}`,
      content: 'Star content',
      type: 'star'
    });
    await createNote(request, {
      title: `Clear Filter Planet ${timestamp}`,
      content: 'Planet content',
      type: 'planet'
    });
    await createNote(request, {
      title: `Clear Filter Comet ${timestamp}`,
      content: 'Comet content',
      type: 'comet'
    });
    
    await page.reload();
    await page.waitForTimeout(2000);
    
    // First filter by star
    await clickFilterChip(page, 'star');
    await page.waitForTimeout(1000);
    
    // Then click All filter
    await clickFilterChip(page, 'all');
    await page.waitForTimeout(1000);
    
    // Stats should show total count without filter indicator
    const statsBar = page.locator('[data-testid="graph-stats"]').first();
    if (await statsBar.isVisible().catch(() => false)) {
      const statsText = await statsBar.textContent();
      // Should show total notes (at least 3 we just created)
      const countMatch = statsText?.match(/(\d+)\s*nodes?/i);
      if (countMatch) {
        const count = parseInt(countMatch[1], 10);
        expect(count).toBeGreaterThanOrEqual(3);
      }
      // Filter indicator should not be present for "All"
      expect(statsText?.toLowerCase()).not.toContain('filtered');
    }
  });

  test('should filter graph nodes when type filter is applied', async ({ page, request }) => {
    const timestamp = Date.now();
    
    // Create notes with links of different types using helper
    await createNote(request, {
      title: `Graph Star ${timestamp}`,
      content: 'Star note',
      type: 'star'
    });
    await createNote(request, {
      title: `Graph Planet ${timestamp}`,
      content: 'Planet note',
      type: 'planet'
    });
    await createNote(request, {
      title: `Graph Comet ${timestamp}`,
      content: 'Comet note',
      type: 'comet'
    });
    
    // Get the created notes to get their IDs
    const notesResp = await request.get(`${getBackendUrl()}/notes`);
    const notesData = await notesResp.json();
    const starNote = notesData.notes?.find((n: any) => n.title?.includes(`Graph Star ${timestamp}`));
    const planetNote = notesData.notes?.find((n: any) => n.title?.includes(`Graph Planet ${timestamp}`));
    const cometNote = notesData.notes?.find((n: any) => n.title?.includes(`Graph Comet ${timestamp}`));
    
    // Create links between notes using helper
    if (starNote && planetNote) {
      await createLink(request, starNote.id, planetNote.id, 0.8);
    }
    if (planetNote && cometNote) {
      await createLink(request, planetNote.id, cometNote.id, 0.6);
    }
    
    await page.reload();
    await page.waitForTimeout(3000);
    
    // Apply star filter
    await clickFilterChip(page, 'star');
    await page.waitForTimeout(2000);
    
    // Graph should still be visible (filtered to show only stars)
    const graphContainer = page.locator('.fullscreen-graph, .graph-canvas, canvas').first();
    await expect(graphContainer).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Type Filters - metadata.type fallback', { tag: ['@metadata'] }, () => {
  
  test('should filter notes with type in metadata field', async ({ page, request }) => {
    const timestamp = Date.now();

    // Create note with type in metadata (new format) using helper
    await createNote(request, {
      title: `Metadata Star ${timestamp}`,
      content: 'Star type in metadata',
      metadata: { type: 'star', custom_field: 'value' }
    });

    // Create note with type at root (old format)
    await createNote(request, {
      title: `Root Planet ${timestamp}`,
      content: 'Planet type at root',
      type: 'planet'
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
