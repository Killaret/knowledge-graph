# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: progressive-rendering.spec.ts >> Progressive Graph Rendering - Fog of War >> should display correct celestial body types
- Location: tests\progressive-rendering.spec.ts:347:3

# Error details

```
Error: expect(received).toBeGreaterThanOrEqual(expected)

Expected: >= 4
Received:    1
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e5] [cursor=pointer]:
    - checkbox "Показать все заметки (выключено)" [ref=e6]
    - generic [ref=e7]: Показать все заметки (выключено)
  - generic [ref=e8]:
    - generic [ref=e9]:
      - strong [ref=e10]: "1"
      - text: nodes
    - generic [ref=e11]:
      - strong [ref=e12]: "0"
      - text: links
    - generic [ref=e13]: (Local view)
```

# Test source

```ts
  288 |     const linked2 = await request.post('http://localhost:8080/notes', {
  289 |       data: { title: 'Linked to Second', content: 'Link' }
  290 |     });
  291 |     await request.post('http://localhost:8080/links', {
  292 |       data: { sourceNoteId: note2Id, targetNoteId: (await linked2.json()).id, weight: 0.7 }
  293 |     });
  294 |     
  295 |     // Navigate to first graph
  296 |     await page.goto(`http://localhost:5173/graph/3d/${note1Id}`);
  297 |     await page.waitForLoadState('networkidle');
  298 |     await page.waitForTimeout(2000);
  299 |     
  300 |     const graphContainer = page.locator('.graph-3d-container').first();
  301 |     await expect(graphContainer).toBeVisible();
  302 |     
  303 |     // Navigate to second graph
  304 |     await page.goto(`http://localhost:5173/graph/3d/${note2Id}`);
  305 |     await page.waitForLoadState('networkidle');
  306 |     await page.waitForTimeout(2000);
  307 |     
  308 |     // Verify second graph loaded
  309 |     await expect(graphContainer).toBeVisible();
  310 |     
  311 |     // Verify stats updated
  312 |     const statsBar = page.locator('.stats-bar').first();
  313 |     await expect(statsBar).toBeVisible();
  314 |   });
  315 | 
  316 |   test('should handle empty graph (no connections) gracefully', async ({ page, request }) => {
  317 |     // Create an isolated note with no connections
  318 |     const isolatedNote = await request.post('http://localhost:8080/notes', {
  319 |       data: { title: 'Isolated Node', content: 'No connections', type: 'asteroid' }
  320 |     });
  321 |     const isolatedId = (await isolatedNote.json()).id;
  322 |     
  323 |     // Navigate to 3D graph
  324 |     await page.goto(`http://localhost:5173/graph/3d/${isolatedId}`);
  325 |     await page.waitForLoadState('networkidle');
  326 |     await page.waitForTimeout(2000);
  327 |     
  328 |     // Verify either graph container or no-data message is shown
  329 |     const graphContainer = page.locator('.graph-3d-container').first();
  330 |     const noDataMessage = page.locator('.no-data-message, .empty-content').first();
  331 |     const singleNodeMessage = page.locator('text=/single|one|only/i').first();
  332 |     
  333 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  334 |     const hasNoData = await noDataMessage.isVisible().catch(() => false);
  335 |     
  336 |     // Should show either graph or appropriate message for single node
  337 |     expect(hasGraph || hasNoData).toBe(true);
  338 |     
  339 |     // Verify stats show 1 node, 0 links
  340 |     const statsBar = page.locator('.stats-bar').first();
  341 |     if (await statsBar.isVisible().catch(() => false)) {
  342 |       const statsText = await statsBar.textContent();
  343 |       expect(statsText).toMatch(/1.*node|1.*nodes/i);
  344 |     }
  345 |   });
  346 | 
  347 |   test('should display correct celestial body types', async ({ page, request }) => {
  348 |     // Create notes with different celestial body types
  349 |     const types = ['star', 'planet', 'comet', 'asteroid'];
  350 |     const createdIds = [];
  351 |     
  352 |     for (const type of types) {
  353 |       const note = await request.post('http://localhost:8080/notes', {
  354 |         data: { title: `${type} Node`, content: `Type: ${type}`, type }
  355 |       });
  356 |       createdIds.push((await note.json()).id);
  357 |     }
  358 |     
  359 |     // Link them in a chain
  360 |     for (let i = 0; i < createdIds.length - 1; i++) {
  361 |       await request.post('http://localhost:8080/links', {
  362 |         data: { 
  363 |           sourceNoteId: createdIds[i], 
  364 |           targetNoteId: createdIds[i + 1], 
  365 |           weight: 0.8 
  366 |         }
  367 |       });
  368 |     }
  369 |     
  370 |     // Navigate to first node's graph
  371 |     await page.goto(`http://localhost:5173/graph/3d/${createdIds[0]}`);
  372 |     await page.waitForLoadState('networkidle');
  373 |     await page.waitForTimeout(3000);
  374 |     
  375 |     // Verify graph loaded
  376 |     const graphContainer = page.locator('.graph-3d-container').first();
  377 |     await expect(graphContainer).toBeVisible();
  378 |     
  379 |     // Verify stats show multiple nodes
  380 |     const statsBar = page.locator('.stats-bar').first();
  381 |     await expect(statsBar).toBeVisible();
  382 |     
  383 |     const statsText = await statsBar.textContent();
  384 |     expect(statsText).not.toBeNull();
  385 |     const nodeMatch = statsText!.match(/(\d+)\s*nodes?/i);
  386 |     if (nodeMatch) {
  387 |       const nodeCount = parseInt(nodeMatch[1], 10);
> 388 |       expect(nodeCount).toBeGreaterThanOrEqual(types.length);
      |                         ^ Error: expect(received).toBeGreaterThanOrEqual(expected)
  389 |     }
  390 |   });
  391 | 
  392 | });
  393 | 
  394 | test.describe('Progressive Graph - Camera & Animation', () => {
  395 |   
  396 |   test.beforeAll(async ({ request }) => {
  397 |     try {
  398 |       const healthCheck = await request.get('http://localhost:8080/notes', { timeout: 5000 });
  399 |       backendAvailable = healthCheck.status() < 500;
  400 |     } catch {
  401 |       backendAvailable = false;
  402 |     }
  403 |     
  404 |     if (!backendAvailable) {
  405 |       console.log('⚠️  Backend not available - Camera tests will be skipped');
  406 |     }
  407 |   });
  408 |   
  409 |   test.beforeEach(async ({ page }) => {
  410 |     if (!backendAvailable) {
  411 |       test.skip();
  412 |     }
  413 |     
  414 |     await page.goto('http://localhost:5173/');
  415 |     await page.waitForLoadState('networkidle');
  416 |   });
  417 | 
  418 |   test('should verify WebGL canvas dimensions are correct', async ({ page, request }) => {
  419 |     // Create a note
  420 |     const note = await request.post('http://localhost:8080/notes', {
  421 |       data: { title: 'Canvas Dimension Test', content: 'Testing canvas size' }
  422 |     });
  423 |     const noteId = (await note.json()).id;
  424 |     
  425 |     // Navigate to 3D graph
  426 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  427 |     await page.waitForLoadState('networkidle');
  428 |     await page.waitForTimeout(2000);
  429 |     
  430 |     // Get canvas element
  431 |     const canvas = page.locator('canvas').first();
  432 |     const hasCanvas = await canvas.isVisible().catch(() => false);
  433 |     
  434 |     if (hasCanvas) {
  435 |       const box = await canvas.boundingBox();
  436 |       expect(box).not.toBeNull();
  437 |       expect(box!.width).toBeGreaterThan(100);
  438 |       expect(box!.height).toBeGreaterThan(100);
  439 |       
  440 |       // Verify canvas fills most of the viewport
  441 |       const viewport = page.viewportSize();
  442 |       if (viewport) {
  443 |         expect(box!.width).toBeGreaterThan(viewport.width * 0.8);
  444 |         expect(box!.height).toBeGreaterThan(viewport.height * 0.8);
  445 |       }
  446 |     }
  447 |   });
  448 | 
  449 |   test('should maintain graph container structure', async ({ page, request }) => {
  450 |     // Create a note with connections
  451 |     const note = await request.post('http://localhost:8080/notes', {
  452 |       data: { title: 'Structure Test', content: 'Testing DOM structure' }
  453 |     });
  454 |     const noteId = (await note.json()).id;
  455 |     
  456 |     const linked = await request.post('http://localhost:8080/notes', {
  457 |       data: { title: 'Structure Link', content: 'Link' }
  458 |     });
  459 |     await request.post('http://localhost:8080/links', {
  460 |       data: { sourceNoteId: noteId, targetNoteId: (await linked.json()).id, weight: 0.7 }
  461 |     });
  462 |     
  463 |     // Navigate to 3D graph
  464 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  465 |     await page.waitForLoadState('networkidle');
  466 |     await page.waitForTimeout(2000);
  467 |     
  468 |     // Verify DOM structure
  469 |     const graphContainer = page.locator('.graph-3d-container').first();
  470 |     await expect(graphContainer).toBeVisible();
  471 |     
  472 |     // Check that container has proper styling
  473 |     const containerStyles = await graphContainer.evaluate(el => {
  474 |       const styles = window.getComputedStyle(el);
  475 |       return {
  476 |         position: styles.position,
  477 |         width: styles.width,
  478 |         height: styles.height,
  479 |         overflow: styles.overflow
  480 |       };
  481 |     });
  482 |     
  483 |     expect(containerStyles.position).toBe('relative');
  484 |     expect(containerStyles.overflow).toBe('hidden');
  485 |   });
  486 | 
  487 | });
  488 | 
```