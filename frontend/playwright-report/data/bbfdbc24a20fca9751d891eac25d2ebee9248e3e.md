# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d-modules.spec.ts >> 3D Graph - Modular Architecture >> should render different link types with distinct styling
- Location: tests\graph-3d-modules.spec.ts:461:3

# Error details

```
Error: expect(received).toMatch(expected)

Expected pattern: /4\s*nodes/
Received string:  "1 nodes 0 links (Local view) "
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
  - generic [ref=e17] [cursor=pointer]: Link Types Source
```

# Test source

```ts
  407 |       });
  408 |       notes.push((await note.json()).id);
  409 |     }
  410 |     
  411 |     // Create links
  412 |     await request.post('http://localhost:8080/links', {
  413 |       data: { sourceNoteId: notes[0], targetNoteId: notes[1], weight: 0.7 }
  414 |     });
  415 |     await request.post('http://localhost:8080/links', {
  416 |       data: { sourceNoteId: notes[1], targetNoteId: notes[2], weight: 0.5 }
  417 |     });
  418 |     
  419 |     // Test the /api/graph/all endpoint via proxy
  420 |     const response = await request.get('http://localhost:5173/api/graph/all', { timeout: 10000 });
  421 |     
  422 |     // Should return 200 OK with graph data
  423 |     expect(response.status()).toBe(200);
  424 |     
  425 |     const data = await response.json();
  426 |     // Verify response structure
  427 |     expect(data).toHaveProperty('nodes');
  428 |     expect(data).toHaveProperty('links');
  429 |     expect(Array.isArray(data.nodes)).toBe(true);
  430 |     expect(Array.isArray(data.links)).toBe(true);
  431 |   });
  432 | 
  433 |   test('should display asteroid celestial body', async ({ page, request }) => {
  434 |     const note = await request.post('http://localhost:8080/notes', {
  435 |       data: { title: 'Asteroid Node', content: 'Asteroid type', type: 'asteroid' }
  436 |     });
  437 |     const noteId = (await note.json()).id;
  438 |     
  439 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  440 |     await page.waitForLoadState('networkidle');
  441 |     await page.waitForTimeout(4000);
  442 |     
  443 |     const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
  444 |     await expect(container).toBeVisible();
  445 |   });
  446 | 
  447 |   test('should display debris celestial body', async ({ page, request }) => {
  448 |     const note = await request.post('http://localhost:8080/notes', {
  449 |       data: { title: 'Debris Node', content: 'Debris type', type: 'debris' }
  450 |     });
  451 |     const noteId = (await note.json()).id;
  452 |     
  453 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  454 |     await page.waitForLoadState('networkidle');
  455 |     await page.waitForTimeout(4000);
  456 |     
  457 |     const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
  458 |     await expect(container).toBeVisible();
  459 |   });
  460 | 
  461 |   test('should render different link types with distinct styling', async ({ page, request }) => {
  462 |     // Create notes for different link types
  463 |     const sourceNote = await request.post('http://localhost:8080/notes', {
  464 |       data: { title: 'Link Types Source', content: 'Source for link types test', type: 'star' }
  465 |     });
  466 |     const sourceId = (await sourceNote.json()).id;
  467 |     
  468 |     const referenceTarget = await request.post('http://localhost:8080/notes', {
  469 |       data: { title: 'Reference Target', content: 'Reference link target', type: 'planet' }
  470 |     });
  471 |     const referenceId = (await referenceTarget.json()).id;
  472 |     
  473 |     const dependencyTarget = await request.post('http://localhost:8080/notes', {
  474 |       data: { title: 'Dependency Target', content: 'Dependency link target', type: 'comet' }
  475 |     });
  476 |     const dependencyId = (await dependencyTarget.json()).id;
  477 |     
  478 |     const relatedTarget = await request.post('http://localhost:8080/notes', {
  479 |       data: { title: 'Related Target', content: 'Related link target', type: 'galaxy' }
  480 |     });
  481 |     const relatedId = (await relatedTarget.json()).id;
  482 |     
  483 |     // Create links with different types
  484 |     await request.post('http://localhost:8080/links', {
  485 |       data: { sourceNoteId: sourceId, targetNoteId: referenceId, weight: 0.8, link_type: 'reference' }
  486 |     });
  487 |     await request.post('http://localhost:8080/links', {
  488 |       data: { sourceNoteId: sourceId, targetNoteId: dependencyId, weight: 0.7, link_type: 'dependency' }
  489 |     });
  490 |     await request.post('http://localhost:8080/links', {
  491 |       data: { sourceNoteId: sourceId, targetNoteId: relatedId, weight: 0.6, link_type: 'related' }
  492 |     });
  493 |     
  494 |     await page.goto(`http://localhost:5173/graph/3d/${sourceId}`);
  495 |     await page.waitForLoadState('networkidle');
  496 |     await page.waitForTimeout(4000);
  497 |     
  498 |     // Verify graph renders with multiple link types
  499 |     const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
  500 |     await expect(container).toBeVisible();
  501 |     
  502 |     // Stats should show multiple nodes and links
  503 |     const statsBar = page.locator('.stats-bar').first();
  504 |     if (await statsBar.isVisible().catch(() => false)) {
  505 |       const statsText = await statsBar.textContent();
  506 |       // Should have 4 nodes (source + 3 targets) and 3 links
> 507 |       expect(statsText).toMatch(/4\s*nodes/);
      |                         ^ Error: expect(received).toMatch(expected)
  508 |       expect(statsText).toMatch(/3\s*links/);
  509 |     }
  510 |   });
  511 | 
  512 |   test('should render full 3D graph at /graph/3d without note ID', async ({ page }) => {
  513 |     // Navigate to full 3D graph page
  514 |     await page.goto('http://localhost:5173/graph/3d');
  515 |     await page.waitForLoadState('networkidle');
  516 |     await page.waitForTimeout(4000);
  517 |     
  518 |     // Verify container or loading/error state is visible
  519 |     const container = page.locator('.graph-3d-container, .lazy-loading, .lazy-error, .center.error').first();
  520 |     await expect(container).toBeVisible();
  521 |     
  522 |     // Verify no 404 error
  523 |     const error404 = page.locator('text=404, text=Not Found').first();
  524 |     const has404 = await error404.isVisible().catch(() => false);
  525 |     expect(has404).toBe(false);
  526 |     
  527 |     // Stats bar should show full graph info
  528 |     const statsBar = page.locator('.stats-bar').first();
  529 |     if (await statsBar.isVisible().catch(() => false)) {
  530 |       const statsText = await statsBar.textContent();
  531 |       expect(statsText).toMatch(/\d+\s*nodes/);
  532 |       expect(statsText).toMatch(/\d+\s*links/);
  533 |     }
  534 |   });
  535 | 
  536 |   test('should render isolated note without connections in 3D graph', async ({ page, request }) => {
  537 |     // Create a note with NO connections
  538 |     const isolatedNote = await request.post('http://localhost:8080/notes', {
  539 |       data: { 
  540 |         title: 'Isolated Note No Links', 
  541 |         content: 'This note has no connections to other notes',
  542 |         type: 'star'
  543 |       }
  544 |     });
  545 |     const noteId = (await isolatedNote.json()).id;
  546 |     
  547 |     // Navigate to 3D graph for this isolated note
  548 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  549 |     await page.waitForLoadState('networkidle');
  550 |     await page.waitForTimeout(4000);
  551 |     
  552 |     // Graph should still render (even with single node)
  553 |     const container = page.locator('.graph-3d-container, .no-data-message, .error-overlay').first();
  554 |     await expect(container).toBeVisible();
  555 |     
  556 |     // Stats should show 1 node and 0 links
  557 |     const statsBar = page.locator('.stats-bar').first();
  558 |     if (await statsBar.isVisible().catch(() => false)) {
  559 |       const statsText = await statsBar.textContent();
  560 |       // Should show at least 1 node (the start node)
  561 |       const nodeMatch = statsText?.match(/(\d+)\s*nodes/);
  562 |       if (nodeMatch) {
  563 |         const nodeCount = parseInt(nodeMatch[1], 10);
  564 |         expect(nodeCount).toBeGreaterThanOrEqual(1);
  565 |       }
  566 |       // Links should be 0
  567 |       expect(statsText).toMatch(/0\s*links/);
  568 |     }
  569 |   });
  570 | });
  571 | 
```