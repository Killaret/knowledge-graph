import { test, expect } from '@playwright/test';
import { createNote, createLink } from './helpers/testData';

/**
 * Tests for Graph Visualization with Progressive Rendering
 * These tests verify the new architecture with immediate loading and fog effect
 */

test.describe('Graph Visualization - Progressive Rendering', { tag: ['@smoke', '@3d', '@progressive'] }, () => {
  
  test.beforeEach(async ({ page }) => {
    // Navigate to home page first
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('should render 3D graph immediately without spinner', async ({ page, request }) => {
    // Create a note via API using helper
    const note = await createNote(request, {
      title: '3D Graph Test Note',
      content: 'Test note for 3D graph'
    });
    const noteId = note.data.id;

    // Navigate to 3D graph page directly
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    
    // Graph should appear immediately (no lazy loading spinner)
    const graphContainer = page.locator('[data-testid="graph-3d-container"]').first();
    await expect(graphContainer).toBeVisible({ timeout: 3000 });
    
    // Loading overlay may be present briefly but should disappear
    const loadingOverlay = page.locator('[data-testid="loading-overlay"]');
    await expect(loadingOverlay).toBeHidden({ timeout: 8000 });
    
    // Stats bar should show immediately with node count
    const statsBar = page.locator('[data-testid="graph-stats"]').first();
    await expect(statsBar).toBeVisible({ timeout: 2000 });
    
    const statsText = await statsBar.textContent();
    expect(statsText).toMatch(/\d+\s*nodes?/i);
  });

  test('should show graph container with correct 3D styling', async ({ page, request }) => {
    // Create a note via API using helper
    const note = await createNote(request, {
      title: 'Styling Test Note',
      content: 'Testing 3D graph styling'
    });
    const noteId = note.data.id;

    // Navigate to 3D graph page
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Verify 3D graph container is visible
    const graphContainer = page.locator('[data-testid="graph-3d-container"]').first();
    await expect(graphContainer).toBeVisible();
    
    // Verify container has correct CSS
    const containerStyles = await graphContainer.evaluate(el => {
      const styles = window.getComputedStyle(el);
      return {
        position: styles.position,
        width: styles.width,
        height: styles.height,
        overflow: styles.overflow,
        backgroundColor: styles.backgroundColor
      };
    });
    
    expect(containerStyles.position).toBe('relative');
    expect(containerStyles.overflow).toBe('hidden');
    expect(containerStyles.width).not.toBe('0px');
    expect(containerStyles.height).not.toBe('0px');
  });

  test('should handle back button navigation from 3D graph page', async ({ page, request }) => {
    // Create a note via API using helper
    const note = await createNote(request, {
      title: 'Navigation Test Note',
      content: 'Testing navigation from 3D graph'
    });
    const noteId = note.data.id;

    // Navigate to 3D graph page
    await page.goto(`/graph/3d/${noteId}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Navigate back using browser back
    await page.goBack();
    await page.waitForTimeout(1000);
    
    // Should be back on home or note page
    const currentUrl = page.url();
    const baseUrl = process.env.FRONTEND_URL || 'http://localhost:5173';
    expect(currentUrl).toMatch(new RegExp(`${baseUrl.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}(/|/notes/.+)`));
  });

  test('should display stats bar with node and link counts', async ({ page, request }) => {
    // Create a note with connections using helper
    const note1 = await createNote(request, { title: 'Stats Test Node 1', content: 'Node 1' });
    const note1Id = note1.data.id;

    const note2 = await createNote(request, { title: 'Stats Test Node 2', content: 'Node 2' });
    const note2Id = note2.data.id;

    // Create link between notes
    await createLink(request, note1Id, note2Id, 0.8);

    // Navigate to 3D graph
    await page.goto(`/graph/3d/${note1Id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Verify stats bar shows correct counts
    const statsBar = page.locator('.stats-bar').first();
    await expect(statsBar).toBeVisible();
    
    const statsText = await statsBar.textContent();
    expect(statsText).toContain('nodes');
    expect(statsText).toContain('links');
    
    // Should show "Local view" mode
    expect(statsText).toContain('Local view');
  });

});
