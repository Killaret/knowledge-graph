# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d-modules.spec.ts >> 3D Graph - Modular Architecture >> should open graph page and verify full graph toggle
- Location: tests\graph-3d-modules.spec.ts:250:3

# Error details

```
TimeoutError: locator.click: Timeout 15000ms exceeded.
Call log:
  - waiting for locator('.toggle input[type="checkbox"], [data-testid="full-graph-toggle"]').first()
    - locator resolved to <input type="checkbox" class="s-5mW6p3cFhHG9"/>
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
    28 × waiting for element to be visible, enabled and stable
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
    - generic [ref=e15]:
      - generic [ref=e16]: ⚠️
      - paragraph [ref=e17]: Failed to load 3D visualization
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
  182 |   test('should auto-zoom camera to fit graph', async ({ page, request }) => {
  183 |     // Create multiple notes to form a graph
  184 |     const notes = [];
  185 |     for (let i = 0; i < 5; i++) {
  186 |       const note = await request.post('http://localhost:8080/notes', {
  187 |         data: { title: `Note ${i}`, content: `Content ${i}` }
  188 |       });
  189 |       notes.push((await note.json()).id);
  190 |     }
  191 |     
  192 |     // Create links between notes
  193 |     for (let i = 0; i < notes.length - 1; i++) {
  194 |       await request.post('http://localhost:8080/links', {
  195 |         data: { sourceNoteId: notes[i], targetNoteId: notes[i + 1], weight: 0.5 }
  196 |       });
  197 |     }
  198 |     
  199 |     await page.goto(`http://localhost:5173/graph/3d/${notes[0]}`);
  200 |     await page.waitForLoadState('networkidle');
  201 |     await page.waitForTimeout(4000); // Wait for simulation to settle and auto-zoom
  202 |     
  203 |     const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
  204 |     await expect(container).toBeVisible();
  205 |     
  206 |     // After auto-zoom, loading should be complete (no loading overlay)
  207 |     const loading = page.locator('.loading-overlay');
  208 |     const isLoading = await loading.isVisible().catch(() => false);
  209 |     // Either not loading or error state is acceptable
  210 |     expect(typeof isLoading).toBe('boolean');
  211 |   });
  212 | 
  213 |   test('should handle full graph toggle', async ({ page, request }) => {
  214 |     const note = await request.post('http://localhost:8080/notes', {
  215 |       data: { title: 'Toggle Test', content: 'Testing full graph toggle' }
  216 |     });
  217 |     const noteId = (await note.json()).id;
  218 |     
  219 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  220 |     await page.waitForLoadState('networkidle');
  221 |     await page.waitForTimeout(2000);
  222 |     
  223 |     // Find and click the toggle
  224 |     const toggle = page.locator('.toggle input[type="checkbox"]').first();
  225 |     if (await toggle.isVisible().catch(() => false)) {
  226 |       await toggle.click();
  227 |       await page.waitForTimeout(2000);
  228 |       
  229 |       // Verify graph still renders after toggle
  230 |       const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
  231 |       await expect(container).toBeVisible();
  232 |     }
  233 |   });
  234 | 
  235 |   test('should display node labels via CSS2D', async ({ page, request }) => {
  236 |     const note = await request.post('http://localhost:8080/notes', {
  237 |       data: { title: 'Labeled Node', content: 'Testing CSS2D labels' }
  238 |     });
  239 |     const noteId = (await note.json()).id;
  240 |     
  241 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  242 |     await page.waitForLoadState('networkidle');
  243 |     await page.waitForTimeout(4000);
  244 |     
  245 |     // Check for labels in the DOM (CSS2D creates div elements)
  246 |     const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
  247 |     await expect(container).toBeVisible();
  248 |   });
  249 | 
  250 |   test('should open graph page and verify full graph toggle', async ({ page, request }) => {
  251 |     // Create test notes
  252 |     const note1 = await request.post('http://localhost:8080/notes', {
  253 |       data: { title: 'Main Note', content: 'Main content', type: 'star' }
  254 |     });
  255 |     const note1Id = (await note1.json()).id;
  256 |     
  257 |     const note2 = await request.post('http://localhost:8080/notes', {
  258 |       data: { title: 'Linked Note', content: 'Linked content', type: 'planet' }
  259 |     });
  260 |     const note2Id = (await note2.json()).id;
  261 |     
  262 |     // Create link between notes
  263 |     await request.post('http://localhost:8080/links', {
  264 |       data: { sourceNoteId: note1Id, targetNoteId: note2Id, weight: 0.8 }
  265 |     });
  266 |     
  267 |     // Test 1: Open 3D graph page for specific note (local constellation)
  268 |     await page.goto(`http://localhost:5173/graph/3d/${note1Id}`);
  269 |     await page.waitForLoadState('networkidle');
  270 |     await page.waitForTimeout(4000);
  271 |     
  272 |     // Verify graph container or error state is visible
  273 |     const container = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
  274 |     await expect(container).toBeVisible();
  275 |     
  276 |     // Test 2: Verify the "Show all notes" toggle exists and works
  277 |     const toggle = page.locator('.toggle input[type="checkbox"], [data-testid="full-graph-toggle"]').first();
  278 |     const hasToggle = await toggle.isVisible().catch(() => false);
  279 |     
  280 |     if (hasToggle) {
  281 |       // Click toggle to switch to full graph view
> 282 |       await toggle.click();
      |                    ^ TimeoutError: locator.click: Timeout 15000ms exceeded.
  283 |       await page.waitForTimeout(3000);
  284 |       
  285 |       // Verify graph still renders after toggle
  286 |       const containerAfterToggle = page.locator('.graph-3d-container, .lazy-error, .error-overlay, .center.error').first();
  287 |       await expect(containerAfterToggle).toBeVisible();
  288 |       
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
```