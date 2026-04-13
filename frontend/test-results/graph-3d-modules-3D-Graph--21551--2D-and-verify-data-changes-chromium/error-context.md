# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d-modules.spec.ts >> 3D Graph - Modular Architecture >> should toggle full graph mode in 2D and verify data changes
- Location: tests\graph-3d-modules.spec.ts:353:3

# Error details

```
TimeoutError: locator.click: Timeout 15000ms exceeded.
Call log:
  - waiting for locator('.toggle input[type="checkbox"]').first()
    - locator resolved to <input type="checkbox" class="s-pF_U0RPRS0F9"/>
  - attempting click action
    2 × waiting for element to be visible, enabled and stable
      - element is visible, enabled and stable
      - scrolling into view if needed
      - done scrolling
      - <vite-error-overlay></vite-error-overlay> intercepts pointer events
    - retrying click action
    - waiting 20ms
    2 × waiting for element to be visible, enabled and stable
      - element is visible, enabled and stable
      - scrolling into view if needed
      - done scrolling
      - <vite-error-overlay></vite-error-overlay> intercepts pointer events
    - retrying click action
      - waiting 100ms
    27 × waiting for element to be visible, enabled and stable
       - element is visible, enabled and stable
       - scrolling into view if needed
       - done scrolling
       - <vite-error-overlay></vite-error-overlay> intercepts pointer events
     - retrying click action
       - waiting 500ms

```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - generic [ref=e3]:
    - button "Go back" [ref=e4] [cursor=pointer]: « Back
    - heading "Knowledge Graph" [level=1] [ref=e5]
    - generic [ref=e7] [cursor=pointer]:
      - checkbox "Показать все заметки (выключено)" [ref=e8]
      - generic [ref=e9]: Показать все заметки (выключено)
    - generic [ref=e10]:
      - generic [ref=e11]:
        - strong [ref=e12]: "1"
        - text: nodes
      - generic [ref=e13]:
        - strong [ref=e14]: "0"
        - text: links
      - generic [ref=e15]: (Local view)
  - generic [ref=e21]:
    - generic [ref=e22]: "[plugin:vite-plugin-svelte] D:/knowledge-graph/frontend/src/lib/components/Graph3D.svelte:167:2 'import' and 'export' may only appear at the top level https://svelte.dev/e/js_parse_error"
    - generic [ref=e23]: D:/knowledge-graph/frontend/src/lib/components/Graph3D.svelte:167:2
    - generic [ref=e24]: "165 | 166 | // Public method to reset camera 167 | export function resetCamera() { ^ 168 | if (simulation && camera && controls) { 169 | autoZoomToFit(simulation.nodes(), camera, controls);"
    - generic [ref=e25]:
      - text: Click outside, press Esc key, or fix the code to dismiss.
      - text: You can also disable this overlay by setting
      - code [ref=e26]: server.hmr.overlay
      - text: to
      - code [ref=e27]: "false"
      - text: in
      - code [ref=e28]: vite.config.ts
      - text: .
```

# Test source

```ts
  289 |       // Re-query toggle for second click (DOM might have changed)
  290 |       const toggleAfter = page.locator('.toggle input[type="checkbox"], [data-testid="full-graph-toggle"]').first();
  291 |       const hasToggleAfter = await toggleAfter.isVisible().catch(() => false);
  292 |       
  293 |       if (hasToggleAfter) {
  294 |         // Click again to switch back to local view
  295 |         await toggleAfter.click();
  296 |         await page.waitForTimeout(2000);
  297 |       }
  298 |     }
  299 |   });
  300 | 
  301 |   test('should display correct stats in 2D graph page', async ({ page, request }) => {
  302 |     // Get current notes count
  303 |     const notesResp = await request.get('http://localhost:8080/notes');
  304 |     const notesData = await notesResp.json();
  305 |     const totalNotes = notesData.total || 0;
  306 |     console.log('Total notes in DB:', totalNotes);
  307 |     
  308 |     // Navigate to 2D graph page
  309 |     await page.goto('http://localhost:5173/graph');
  310 |     await page.waitForLoadState('networkidle');
  311 |     await page.waitForTimeout(3000);
  312 |     
  313 |     // Check for stats bar
  314 |     const statsBar = page.locator('.stats-bar').first();
  315 |     const hasStats = await statsBar.isVisible().catch(() => false);
  316 |     
  317 |     if (hasStats) {
  318 |       const statsText = await statsBar.textContent();
  319 |       // Verify stats show numbers (nodes/links count)
  320 |       expect(statsText).toMatch(/\d+\s+nodes?/i);
  321 |       expect(statsText).toMatch(/\d+\s+links?/i);
  322 |       
  323 |       // Verify mode indicator (Full graph or Local view)
  324 |       const hasMode = statsText?.toLowerCase().includes('full') || statsText?.toLowerCase().includes('local');
  325 |       expect(hasMode).toBe(true);
  326 |     }
  327 |   });
  328 | 
  329 |   test('should display correct stats in 3D graph page', async ({ page, request }) => {
  330 |     // Create a note for 3D graph
  331 |     const note = await request.post('http://localhost:8080/notes', {
  332 |       data: { title: '3D Stats Test', content: 'Testing 3D stats display', type: 'galaxy' }
  333 |     });
  334 |     const noteId = (await note.json()).id;
  335 |     
  336 |     // Navigate to 3D graph page
  337 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  338 |     await page.waitForLoadState('networkidle');
  339 |     await page.waitForTimeout(4000);
  340 |     
  341 |     // Check for stats bar
  342 |     const statsBar = page.locator('.stats-bar').first();
  343 |     const hasStats = await statsBar.isVisible().catch(() => false);
  344 |     
  345 |     if (hasStats) {
  346 |       const statsText = await statsBar.textContent();
  347 |       // Verify stats show numbers
  348 |       expect(statsText).toMatch(/\d+\s+nodes?/i);
  349 |       expect(statsText).toMatch(/\d+\s+links?/i);
  350 |     }
  351 |   });
  352 | 
  353 |   test('should toggle full graph mode in 2D and verify data changes', async ({ page, request }) => {
  354 |     // Ensure we have notes with links
  355 |     const note1 = await request.post('http://localhost:8080/notes', {
  356 |       data: { title: '2D Toggle Test 1', content: 'First note', type: 'star' }
  357 |     });
  358 |     const note1Id = (await note1.json()).id;
  359 |     
  360 |     const note2 = await request.post('http://localhost:8080/notes', {
  361 |       data: { title: '2D Toggle Test 2', content: 'Second note', type: 'planet' }
  362 |     });
  363 |     const note2Id = (await note2.json()).id;
  364 |     
  365 |     // Create link
  366 |     await request.post('http://localhost:8080/links', {
  367 |       data: { source_note_id: note1Id, target_note_id: note2Id, link_type: 'reference', weight: 0.7 }
  368 |     });
  369 |     
  370 |     // Navigate to 2D graph
  371 |     await page.goto('http://localhost:5173/graph');
  372 |     await page.waitForLoadState('networkidle');
  373 |     await page.waitForTimeout(3000);
  374 |     
  375 |     // Find toggle
  376 |     const toggle = page.locator('.toggle input[type="checkbox"]').first();
  377 |     const hasToggle = await toggle.isVisible().catch(() => false);
  378 |     
  379 |     if (!hasToggle) {
  380 |       test.skip(true, 'Toggle not found on 2D graph page');
  381 |       return;
  382 |     }
  383 |     
  384 |     // Get initial stats
  385 |     const statsBefore = await page.locator('.stats-bar').textContent().catch(() => '') || '';
  386 |     console.log('Stats before toggle:', statsBefore);
  387 |     
  388 |     // Toggle to local view
> 389 |     await toggle.click();
      |                  ^ TimeoutError: locator.click: Timeout 15000ms exceeded.
  390 |     await page.waitForTimeout(3000);
  391 |     
  392 |     // Get stats after toggle
  393 |     const statsAfter = await page.locator('.stats-bar').textContent().catch(() => '') || '';
  394 |     console.log('Stats after toggle:', statsAfter);
  395 |     
  396 |     // Graph should still render
  397 |     const container = page.locator('.graph-container, .error-overlay').first();
  398 |     await expect(container).toBeVisible();
  399 |   });
  400 | 
  401 |   test('should fetch full graph data via API', async ({ request }) => {
  402 |     // Create test notes for full graph
  403 |     const notes = [];
  404 |     for (let i = 0; i < 3; i++) {
  405 |       const note = await request.post('http://localhost:8080/notes', {
  406 |         data: { title: `Full Graph Note ${i}`, content: `Content ${i}`, type: i === 0 ? 'star' : 'planet' }
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
  432 | });
  433 | 
```