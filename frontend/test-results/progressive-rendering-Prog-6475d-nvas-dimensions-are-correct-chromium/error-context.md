# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: progressive-rendering.spec.ts >> Progressive Graph - Camera & Animation >> should verify WebGL canvas dimensions are correct
- Location: tests\progressive-rendering.spec.ts:396:3

# Error details

```
Error: page.waitForLoadState: Target page, context or browser has been closed
```

# Test source

```ts
  293 |     
  294 |     // Navigate to 3D graph
  295 |     await page.goto(`/graph/3d/${isolatedId}`);
  296 |     await page.waitForLoadState('networkidle');
  297 |     await page.waitForTimeout(2000);
  298 |     
  299 |     // Verify either graph container or no-data message is shown
  300 |     const graphContainer = page.locator('.graph-3d-container').first();
  301 |     const noDataMessage = page.locator('.no-data-message, .empty-content').first();
  302 |     
  303 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  304 |     const hasNoData = await noDataMessage.isVisible().catch(() => false);
  305 |     
  306 |     // Should show either graph or appropriate message for single node
  307 |     expect(hasGraph || hasNoData).toBe(true);
  308 |     
  309 |     // Verify stats show 1 node, 0 links
  310 |     const statsBar = page.locator('.stats-bar').first();
  311 |     if (await statsBar.isVisible().catch(() => false)) {
  312 |       const statsText = await statsBar.textContent();
  313 |       expect(statsText).toMatch(/1.*node|1.*nodes/i);
  314 |     }
  315 |   });
  316 | 
  317 |   test('should display correct celestial body types', async ({ page, request }) => {
  318 |     // Create a central star node
  319 |     const centralNote = await request.post('http://localhost:8080/notes', {
  320 |       data: { title: 'Central Star', content: 'Central node', type: 'star' }
  321 |     });
  322 |     const centralId = (await centralNote.json()).id;
  323 |     
  324 |     // Create planet, comet, asteroid linked to central
  325 |     const linkedIds = [];
  326 |     for (const type of ['planet', 'comet', 'asteroid']) {
  327 |       const note = await request.post('http://localhost:8080/notes', {
  328 |         data: { title: `${type} Node`, content: `Type: ${type}`, type }
  329 |       });
  330 |       const noteId = (await note.json()).id;
  331 |       linkedIds.push(noteId);
  332 |     }
  333 |     
  334 |     // Create all links
  335 |     for (const linkedId of linkedIds) {
  336 |       await createLink(request, centralId, linkedId, 0.8);
  337 |     }
  338 |     
  339 |     // Navigate to central node's graph
  340 |     await page.goto(`/graph/3d/${centralId}`);
  341 |     await page.waitForLoadState('networkidle');
  342 |     
  343 |     // Wait for progressive loading to complete (initial + background)
  344 |     await page.waitForTimeout(6000);
  345 |     
  346 |     // Verify graph loaded
  347 |     const graphContainer = page.locator('.graph-3d-container').first();
  348 |     await expect(graphContainer).toBeVisible();
  349 |     
  350 |     // Verify stats bar shows the graph data
  351 |     const statsBar = page.locator('.stats-bar').first();
  352 |     await expect(statsBar).toBeVisible({ timeout: 5000 });
  353 |     
  354 |     const statsText = await statsBar.textContent();
  355 |     expect(statsText).not.toBeNull();
  356 |     
  357 |     // Just verify stats show some nodes (graph API might filter based on depth)
  358 |     const nodeMatch = statsText!.match(/(\d+)\s*nodes?/i);
  359 |     expect(nodeMatch).not.toBeNull();
  360 |     if (nodeMatch) {
  361 |       const nodeCount = parseInt(nodeMatch[1], 10);
  362 |       // Should have at least the central node
  363 |       expect(nodeCount).toBeGreaterThanOrEqual(1);
  364 |     }
  365 |     
  366 |     // Verify links count is shown
  367 |     expect(statsText).toMatch(/\d+\s*links?/i);
  368 |   });
  369 | 
  370 | });
  371 | 
  372 | test.describe('Progressive Graph - Camera & Animation', () => {
  373 |   
  374 |   test.beforeAll(async ({ request }) => {
  375 |     try {
  376 |       const healthCheck = await request.get('http://localhost:8080/notes', { timeout: 5000 });
  377 |       backendAvailable = healthCheck.status() < 500;
  378 |     } catch {
  379 |       backendAvailable = false;
  380 |     }
  381 |     
  382 |     if (!backendAvailable) {
  383 |       console.log('⚠️  Backend not available - Camera tests will be skipped');
  384 |     }
  385 |   });
  386 |   
  387 |   test.beforeEach(async ({ page }) => {
  388 |     if (!backendAvailable) {
  389 |       test.skip();
  390 |     }
  391 |     
  392 |     await page.goto('/');
> 393 |     await page.waitForLoadState('networkidle');
      |                ^ Error: page.waitForLoadState: Target page, context or browser has been closed
  394 |   });
  395 | 
  396 |   test('should verify WebGL canvas dimensions are correct', async ({ page, request }) => {
  397 |     // Create a note
  398 |     const note = await request.post('http://localhost:8080/notes', {
  399 |       data: { title: 'Canvas Dimension Test', content: 'Testing canvas size' }
  400 |     });
  401 |     const noteId = (await note.json()).id;
  402 |     
  403 |     // Navigate to 3D graph
  404 |     await page.goto(`/graph/3d/${noteId}`);
  405 |     await page.waitForLoadState('networkidle');
  406 |     await page.waitForTimeout(2000);
  407 |     
  408 |     // Get canvas element
  409 |     const canvas = page.locator('canvas').first();
  410 |     const hasCanvas = await canvas.isVisible().catch(() => false);
  411 |     
  412 |     if (hasCanvas) {
  413 |       const box = await canvas.boundingBox();
  414 |       expect(box).not.toBeNull();
  415 |       expect(box!.width).toBeGreaterThan(100);
  416 |       expect(box!.height).toBeGreaterThan(100);
  417 |       
  418 |       // Verify canvas fills most of the viewport
  419 |       const viewport = page.viewportSize();
  420 |       if (viewport) {
  421 |         expect(box!.width).toBeGreaterThan(viewport.width * 0.8);
  422 |         expect(box!.height).toBeGreaterThan(viewport.height * 0.8);
  423 |       }
  424 |     }
  425 |   });
  426 | 
  427 |   test('should maintain graph container structure', async ({ page, request }) => {
  428 |     // Create a note with connections
  429 |     const note = await request.post('http://localhost:8080/notes', {
  430 |       data: { title: 'Structure Test', content: 'Testing DOM structure' }
  431 |     });
  432 |     const noteId = (await note.json()).id;
  433 |     
  434 |     const linked = await request.post('http://localhost:8080/notes', {
  435 |       data: { title: 'Structure Link', content: 'Link' }
  436 |     });
  437 |     await request.post('http://localhost:8080/links', {
  438 |       data: { sourceNoteId: noteId, targetNoteId: (await linked.json()).id, weight: 0.7 }
  439 |     });
  440 |     
  441 |     // Navigate to 3D graph
  442 |     await page.goto(`/graph/3d/${noteId}`);
  443 |     await page.waitForLoadState('networkidle');
  444 |     await page.waitForTimeout(2000);
  445 |     
  446 |     // Verify DOM structure
  447 |     const graphContainer = page.locator('.graph-3d-container').first();
  448 |     await expect(graphContainer).toBeVisible();
  449 |     
  450 |     // Check that container has proper styling
  451 |     const containerStyles = await graphContainer.evaluate(el => {
  452 |       const styles = window.getComputedStyle(el);
  453 |       return {
  454 |         position: styles.position,
  455 |         width: styles.width,
  456 |         height: styles.height,
  457 |         overflow: styles.overflow
  458 |       };
  459 |     });
  460 |     
  461 |     expect(containerStyles.position).toBe('relative');
  462 |     expect(containerStyles.overflow).toBe('hidden');
  463 |   });
  464 | 
  465 | });
  466 | 
```