import { test } from '@playwright/test';
import { createNote, createLink } from '../helpers/testData';

/**
 * Visual Regression Tests with Argos
 * Captures screenshots of key UI states for comparison
 * 
 * Requires: ARGOS_TOKEN environment variable
 * Screenshots saved to: argos-screenshots/
 */

test.describe('Visual Regression @visual', { tag: ['@visual'] }, () => {
  
  test('Home page - default view', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Capture full page
    await page.screenshot({ 
      path: 'argos-screenshots/home-default.png',
      fullPage: true 
    });
  });

  test('Home page - list view', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Click list toggle using JavaScript to bypass viewport issues
    const listToggle = page.locator('[data-testid="view-toggle-list"]');
    if (await listToggle.isVisible().catch(() => false)) {
      await listToggle.evaluate(el => (el as HTMLElement).click());
      await page.waitForTimeout(1000);
    }
    
    await page.screenshot({ 
      path: 'argos-screenshots/home-list-view.png',
      fullPage: true 
    });
  });

  test('Home page - with star filter', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Click star filter using JavaScript to bypass viewport issues
    const starFilter = page.locator('[data-testid="filter-chip-star"]');
    if (await starFilter.isVisible().catch(() => false)) {
      await starFilter.evaluate(el => (el as HTMLElement).click());
      await page.waitForTimeout(1000);
    }
    
    await page.screenshot({ 
      path: 'argos-screenshots/home-filtered-stars.png',
      fullPage: true 
    });
  });

  test('3D Graph - loading state', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Visual 3D Test',
      content: 'Testing 3D visualization',
      type: 'star'
    });
    
    await page.goto(`/graph/3d/${note.id}`);
    
    // Capture loading state
    await page.waitForSelector('[data-testid="loading-overlay"]');
    await page.screenshot({ 
      path: 'argos-screenshots/3d-loading-state.png',
      fullPage: true 
    });
  });

  test('3D Graph - with star', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Star Visual Test',
      content: 'Star type visualization',
      type: 'star'
    });
    
    await page.goto(`/graph/3d/${note.id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000); // Wait for graph to render
    
    await page.screenshot({ 
      path: 'argos-screenshots/3d-star-node.png',
      fullPage: true 
    });
  });

  test('3D Graph - with planet', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Planet Visual Test',
      content: 'Planet type visualization',
      type: 'planet'
    });
    
    await page.goto(`/graph/3d/${note.id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/3d-planet-node.png',
      fullPage: true 
    });
  });

  test('3D Graph - with connections', async ({ page, request }) => {
    const center = await createNote(request, {
      title: 'Center Node',
      content: 'Center',
      type: 'star'
    });
    
    const satellite = await createNote(request, {
      title: 'Satellite Node',
      content: 'Satellite',
      type: 'planet'
    });
    
    await createLink(request, center.id, satellite.id, 0.8, 'related');
    
    await page.goto(`/graph/3d/${center.id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/3d-with-link.png',
      fullPage: true 
    });
  });

  test('ConfirmModal - delete confirmation', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Trigger delete modal (if accessible)
    // For now, just render modal directly via URL or state
    // This is a placeholder - actual implementation depends on how modal is triggered
    
    await page.screenshot({ 
      path: 'argos-screenshots/modal-confirm.png',
      fullPage: false 
    });
  });

  test('Search - with results', async ({ page, request }) => {
    // Create searchable note
    await createNote(request, {
      title: 'Searchable Test Note',
      content: 'Unique content for search',
      type: 'comet'
    });
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Type in search - use JavaScript to ensure input is interactable
    const searchInput = page.locator('[data-testid="search-input"]');
    if (await searchInput.isVisible().catch(() => false)) {
      await searchInput.evaluate(el => (el as HTMLElement).focus());
      await searchInput.fill('Searchable');
      await page.waitForTimeout(1000);
    }
    
    await page.screenshot({ 
      path: 'argos-screenshots/search-with-results.png',
      fullPage: true 
    });
  });

  test('Empty state - no notes', async ({ page }) => {
    // Navigate to list view when no notes exist
    await page.goto('/?view=list');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/empty-state.png',
      fullPage: true 
    });
  });

  test('Full 3D Graph view', async ({ page }) => {
    await page.goto('/graph/3d');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(4000);
    
    await page.screenshot({ 
      path: 'argos-screenshots/3d-full-graph.png',
      fullPage: true 
    });
  });
});
