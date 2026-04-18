# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: performance\graph-3d-performance.spec.ts >> 3D Graph Performance @performance >> should not leak memory during node interactions
- Location: tests\performance\graph-3d-performance.spec.ts:86:3

# Error details

```
Error: page.evaluate: Target crashed 
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | import { createNote, createLink, getBackendUrl } from '../helpers/testData';
  3   | 
  4   | /**
  5   |  * Performance tests for 3D Graph
  6   |  * Measures FPS, loading time, and memory usage with large node counts
  7   |  * 
  8   |  * Thresholds:
  9   |  * - FPS: > 30fps (target: 60fps)
  10  |  * - Load time: < 5s for 100 nodes
  11  |  * - Memory: < 200MB increase
  12  |  */
  13  | 
  14  | test.describe('3D Graph Performance @performance', { tag: ['@performance', '@3d'] }, () => {
  15  |   
  16  |   test('should maintain 30+ FPS with 50 nodes', async ({ page, request }) => {
  17  |     // Create 50 connected notes
  18  |     const notes = [];
  19  |     for (let i = 0; i < 50; i++) {
  20  |       const note = await createNote(request, {
  21  |         title: `Performance Node ${i}`,
  22  |         content: `Content ${i}`,
  23  |         type: i % 5 === 0 ? 'star' : i % 5 === 1 ? 'planet' : 'comet'
  24  |       });
  25  |       notes.push(note.id);
  26  |     }
  27  |     
  28  |     // Create links between consecutive nodes
  29  |     for (let i = 0; i < notes.length - 1; i++) {
  30  |       await createLink(request, notes[i], notes[i + 1], 0.5, 'related');
  31  |     }
  32  |     
  33  |     // Navigate to 3D graph with first node as center
  34  |     await page.goto(`/graph/3d/${notes[0]}`);
  35  |     await page.waitForLoadState('networkidle');
  36  |     
  37  |     // Wait for graph to fully load
  38  |     await page.waitForTimeout(5000);
  39  |     
  40  |     // Measure FPS using Performance API
  41  |     const fps = await page.evaluate(async () => {
  42  |       let frames = 0;
  43  |       const startTime = performance.now();
  44  |       
  45  |       return new Promise<number>((resolve) => {
  46  |         const countFrames = () => {
  47  |           frames++;
  48  |           if (performance.now() - startTime < 1000) {
  49  |             requestAnimationFrame(countFrames);
  50  |           } else {
  51  |             resolve(frames);
  52  |           }
  53  |         };
  54  |         requestAnimationFrame(countFrames);
  55  |       });
  56  |     });
  57  |     
  58  |     console.log(`FPS with 50 nodes: ${fps}`);
  59  |     expect(fps).toBeGreaterThan(30);
  60  |   });
  61  | 
  62  |   test('should load 100 nodes within 5 seconds', async ({ page, request }) => {
  63  |     // Create 100 notes
  64  |     const notes = [];
  65  |     for (let i = 0; i < 100; i++) {
  66  |       const note = await createNote(request, {
  67  |         title: `Load Test Node ${i}`,
  68  |         content: `Content ${i}`,
  69  |         type: 'planet'
  70  |       });
  71  |       notes.push(note.id);
  72  |     }
  73  |     
  74  |     // Measure load time
  75  |     const startTime = Date.now();
  76  |     
  77  |     await page.goto(`/graph/3d/${notes[0]}`);
  78  |     await page.waitForSelector('[data-testid="graph-3d-container"]', { timeout: 10000 });
  79  |     
  80  |     const loadTime = Date.now() - startTime;
  81  |     console.log(`Load time for 100 nodes: ${loadTime}ms`);
  82  |     
  83  |     expect(loadTime).toBeLessThan(5000);
  84  |   });
  85  | 
  86  |   test('should not leak memory during node interactions', async ({ page, request }) => {
  87  |     const note = await createNote(request, {
  88  |       title: 'Memory Test',
  89  |       content: 'Testing memory',
  90  |       type: 'star'
  91  |     });
  92  |     
  93  |     await page.goto(`/graph/3d/${note.id}`);
  94  |     await page.waitForLoadState('networkidle');
  95  |     await page.waitForTimeout(3000);
  96  |     
  97  |     // Get initial memory
> 98  |     const initialMemory = await page.evaluate(() => {
      |                                      ^ Error: page.evaluate: Target crashed 
  99  |       return (performance as any).memory?.usedJSHeapSize || 0;
  100 |     });
  101 |     
  102 |     // Simulate interactions
  103 |     for (let i = 0; i < 10; i++) {
  104 |       await page.mouse.click(400, 300);
  105 |       await page.waitForTimeout(200);
  106 |       await page.mouse.move(400 + i * 10, 300 + i * 10);
  107 |       await page.waitForTimeout(100);
  108 |     }
  109 |     
  110 |     // Force GC if available
  111 |     await page.evaluate(() => {
  112 |       if ('gc' in window) (window as any).gc();
  113 |     });
  114 |     
  115 |     await page.waitForTimeout(1000);
  116 |     
  117 |     // Get final memory
  118 |     const finalMemory = await page.evaluate(() => {
  119 |       return (performance as any).memory?.usedJSHeapSize || 0;
  120 |     });
  121 |     
  122 |     const memoryIncrease = finalMemory - initialMemory;
  123 |     console.log(`Memory increase: ${memoryIncrease / 1024 / 1024}MB`);
  124 |     
  125 |     // Allow up to 50MB increase
  126 |     if (initialMemory > 0) {
  127 |       expect(memoryIncrease).toBeLessThan(50 * 1024 * 1024);
  128 |     }
  129 |   });
  130 | 
  131 |   test('should handle rapid filter changes without crashing', async ({ page, request }) => {
  132 |     // Create notes of different types
  133 |     const types = ['star', 'planet', 'comet', 'galaxy'];
  134 |     const notes = [];
  135 |     
  136 |     for (let i = 0; i < 40; i++) {
  137 |       const note = await createNote(request, {
  138 |         title: `Filter Test ${i}`,
  139 |         content: 'Test',
  140 |         type: types[i % 4]
  141 |       });
  142 |       notes.push(note.id);
  143 |     }
  144 |     
  145 |     await page.goto('/');
  146 |     await page.waitForLoadState('networkidle');
  147 |     
  148 |     // Rapidly switch filters
  149 |     for (const type of ['all', 'star', 'planet', 'comet', 'galaxy', 'all']) {
  150 |       const filterButton = page.locator(`[data-testid="filter-chip-${type}"]`);
  151 |       if (await filterButton.isVisible().catch(() => false)) {
  152 |         await filterButton.click();
  153 |         await page.waitForTimeout(100);
  154 |       }
  155 |     }
  156 |     
  157 |     // Verify graph still responsive
  158 |     const graph = page.locator('[data-testid="graph-2d-container"]');
  159 |     await expect(graph).toBeVisible();
  160 |   });
  161 | 
  162 |   test('Lighthouse performance audit', async ({ page }) => {
  163 |     // This test is a placeholder for Lighthouse CI integration
  164 |     // Actual Lighthouse run happens in CI workflow
  165 |     
  166 |     await page.goto('/');
  167 |     await page.waitForLoadState('networkidle');
  168 |     
  169 |     // Basic performance metrics
  170 |     const metrics = await page.evaluate(() => {
  171 |       const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
  172 |       return {
  173 |         domContentLoaded: nav.domContentLoadedEventEnd - nav.startTime,
  174 |         loadComplete: nav.loadEventEnd - nav.startTime,
  175 |         firstPaint: performance.getEntriesByName('first-paint')[0]?.startTime,
  176 |         firstContentfulPaint: performance.getEntriesByName('first-contentful-paint')[0]?.startTime
  177 |       };
  178 |     });
  179 |     
  180 |     console.log('Performance metrics:', metrics);
  181 |     
  182 |     expect(metrics.domContentLoaded).toBeLessThan(3000);
  183 |     expect(metrics.loadComplete).toBeLessThan(5000);
  184 |   });
  185 | });
  186 | 
```