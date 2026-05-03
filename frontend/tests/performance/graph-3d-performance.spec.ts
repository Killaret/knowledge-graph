import { test, expect } from '@playwright/test';
import { createNote, createLink } from '../helpers/testData';

/**
 * Performance tests for 3D Graph
 * Measures FPS, loading time, and memory usage with large node counts
 * 
 * Thresholds:
 * - FPS: > 30fps (target: 60fps)
 * - Load time: < 5s for 100 nodes
 * - Memory: < 200MB increase
 */

test.describe('3D Graph Performance @performance', { tag: ['@performance', '@3d'] }, () => {
  
  test('should maintain 30+ FPS with 50 nodes', async ({ page, request }) => {
    // Create 50 connected notes
    const notes = [];
    for (let i = 0; i < 50; i++) {
      const note = await createNote(request, {
        title: `Performance Node ${i}`,
        content: `Content ${i}`,
        type: i % 5 === 0 ? 'star' : i % 5 === 1 ? 'planet' : 'comet'
      });
      notes.push(note.data.id);
    }
    
    // Create links between consecutive nodes
    for (let i = 0; i < notes.length - 1; i++) {
      await createLink(request, notes[i], notes[i + 1], 0.5, 'related');
    }
    
    // Navigate to 3D graph with first node as center
    await page.goto(`/graph/3d/${notes[0]}`);
    await page.waitForLoadState('networkidle');
    
    // Wait for graph to fully load
    await page.waitForTimeout(5000);
    
    // Measure FPS using Performance API
    const fps = await page.evaluate(async () => {
      let frames = 0;
      const startTime = performance.now();
      
      return new Promise<number>((resolve) => {
        const countFrames = () => {
          frames++;
          if (performance.now() - startTime < 1000) {
            requestAnimationFrame(countFrames);
          } else {
            resolve(frames);
          }
        };
        requestAnimationFrame(countFrames);
      });
    });
    
    console.log(`FPS with 50 nodes: ${fps}`);
    expect(fps).toBeGreaterThan(30);
  });

  test('should load 100 nodes within 5 seconds', async ({ page, request }) => {
    // Create 100 notes
    const notes = [];
    for (let i = 0; i < 100; i++) {
      const note = await createNote(request, {
        title: `Load Test Node ${i}`,
        content: `Content ${i}`,
        type: 'planet'
      });
      notes.push(note.data.id);
    }
    
    // Measure load time
    const startTime = Date.now();
    
    await page.goto(`/graph/3d/${notes[0]}`);
    await page.waitForSelector('[data-testid="graph-3d-container"]', { timeout: 10000 });
    
    const loadTime = Date.now() - startTime;
    console.log(`Load time for 100 nodes: ${loadTime}ms`);
    
    expect(loadTime).toBeLessThan(5000);
  });

  test('should not leak memory during node interactions', async ({ page, request }) => {
    const note = await createNote(request, {
      title: 'Memory Test',
      content: 'Testing memory',
      type: 'star'
    });
    
    await page.goto(`/graph/3d/${note.data.id}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);
    
    // Get initial memory
    const initialMemory = await page.evaluate(() => {
      return (performance as any).memory?.usedJSHeapSize || 0;
    });
    
    // Simulate interactions
    for (let i = 0; i < 10; i++) {
      await page.mouse.click(400, 300);
      await page.waitForTimeout(200);
      await page.mouse.move(400 + i * 10, 300 + i * 10);
      await page.waitForTimeout(100);
    }
    
    // Force GC if available
    await page.evaluate(() => {
      if ('gc' in window) (window as any).gc();
    });
    
    await page.waitForTimeout(1000);
    
    // Get final memory
    const finalMemory = await page.evaluate(() => {
      return (performance as any).memory?.usedJSHeapSize || 0;
    });
    
    const memoryIncrease = finalMemory - initialMemory;
    console.log(`Memory increase: ${memoryIncrease / 1024 / 1024}MB`);
    
    // Allow up to 50MB increase
    if (initialMemory > 0) {
      expect(memoryIncrease).toBeLessThan(50 * 1024 * 1024);
    }
  });

  test('should handle rapid filter changes without crashing', async ({ page, request }) => {
    // Create notes of different types
    const types = ['star', 'planet', 'comet', 'galaxy'];
    const notes = [];
    
    for (let i = 0; i < 40; i++) {
      const note = await createNote(request, {
        title: `Filter Test ${i}`,
        content: 'Test',
        type: types[i % 4]
      });
      notes.push(note.data.id);
    }
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Rapidly switch filters
    for (const type of ['all', 'star', 'planet', 'comet', 'galaxy', 'all']) {
      const filterButton = page.locator(`[data-testid="filter-chip-${type}"]`);
      if (await filterButton.isVisible().catch(() => false)) {
        await filterButton.click();
        await page.waitForTimeout(100);
      }
    }
    
    // Verify graph still responsive
    const graph = page.locator('[data-testid="graph-2d-container"]');
    await expect(graph).toBeVisible();
  });

  test('Lighthouse performance audit', async ({ page }) => {
    // This test is a placeholder for Lighthouse CI integration
    // Actual Lighthouse run happens in CI workflow
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Basic performance metrics
    const metrics = await page.evaluate(() => {
      const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return {
        domContentLoaded: nav.domContentLoadedEventEnd - nav.startTime,
        loadComplete: nav.loadEventEnd - nav.startTime,
        firstPaint: performance.getEntriesByName('first-paint')[0]?.startTime,
        firstContentfulPaint: performance.getEntriesByName('first-contentful-paint')[0]?.startTime
      };
    });
    
    console.log('Performance metrics:', metrics);
    
    expect(metrics.domContentLoaded).toBeLessThan(3000);
    expect(metrics.loadComplete).toBeLessThan(5000);
  });
});
