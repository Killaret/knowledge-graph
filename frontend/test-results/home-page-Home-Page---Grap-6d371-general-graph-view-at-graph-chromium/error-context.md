# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: home-page.spec.ts >> Home Page - Graph First >> should display general graph view at /graph
- Location: tests\home-page.spec.ts:224:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - button "Go back" [ref=e4] [cursor=pointer]: « Back
  - heading "Knowledge Graph" [level=1] [ref=e5]
  - generic [ref=e7] [cursor=pointer]:
    - checkbox "Показать все заметки (включено)" [checked] [ref=e8]
    - generic [ref=e9]: Показать все заметки (включено)
  - generic [ref=e10]:
    - paragraph [ref=e11]: Failed to load graph data
    - button "Go Home" [ref=e12] [cursor=pointer]
```

# Test source

```ts
  141 |       if (hasFilterText) {
  142 |         const filterText = await statsFilter.textContent();
  143 |         expect(filterText?.toLowerCase()).toContain('filter');
  144 |       }
  145 |     }
  146 |   });
  147 | 
  148 |   test('should search notes from home page', async ({ page, request }) => {
  149 |     // Create a searchable note
  150 |     const timestamp = Date.now();
  151 |     const searchTerm = `Searchable${timestamp}`;
  152 |     await request.post('http://localhost:8080/notes', {
  153 |       data: { 
  154 |         title: `Test ${searchTerm} Note`, 
  155 |         content: 'Test content',
  156 |         type: 'star'
  157 |       }
  158 |     });
  159 |     
  160 |     // Reload page
  161 |     await page.reload();
  162 |     await page.waitForTimeout(2000);
  163 |     
  164 |     // Fill search input
  165 |     const searchInput = page.locator('.search-input, input[type="search"]').first();
  166 |     if (await searchInput.isVisible().catch(() => false)) {
  167 |       await searchInput.fill(searchTerm);
  168 |       await page.waitForTimeout(1000); // Wait for search to apply
  169 |       
  170 |       // Verify stats bar appears
  171 |       const statsBar = page.locator('.stats-bar').first();
  172 |       await expect(statsBar).toBeVisible();
  173 |     }
  174 |   });
  175 | 
  176 |   test('should open side panel when clicking on graph node', async ({ page, request }) => {
  177 |     // Create a note
  178 |     const timestamp = Date.now();
  179 |     const note = await request.post('http://localhost:8080/notes', {
  180 |       data: { 
  181 |         title: `Side Panel Test ${timestamp}`, 
  182 |         content: 'Test content for side panel',
  183 |         type: 'star'
  184 |       }
  185 |     });
  186 |     await note.json();
  187 |     
  188 |     // Reload page
  189 |     await page.reload();
  190 |     await page.waitForTimeout(2000);
  191 |     
  192 |     // Try to click on a note card (fallback if graph click doesn't work)
  193 |     const noteCard = page.locator('.note-card').first();
  194 |     if (await noteCard.isVisible().catch(() => false)) {
  195 |       await noteCard.click();
  196 |       
  197 |       // Verify side panel opens
  198 |       const sidePanel = page.locator('.side-panel, .note-side-panel').first();
  199 |       await expect(sidePanel).toBeVisible({ timeout: 5000 });
  200 |     }
  201 |   });
  202 | 
  203 |   test('should navigate to graph view for specific note', async ({ page, request }) => {
  204 |     // Create a note
  205 |     const timestamp = Date.now();
  206 |     const note = await request.post('http://localhost:8080/notes', {
  207 |       data: { 
  208 |         title: `Graph View Test ${timestamp}`, 
  209 |         content: 'Test content',
  210 |         type: 'star'
  211 |       }
  212 |     });
  213 |     const noteId = (await note.json()).id;
  214 |     
  215 |     // Navigate to specific graph page
  216 |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  217 |     await page.waitForTimeout(2000);
  218 |     
  219 |     // Verify graph container is visible
  220 |     const graphContainer = page.locator('.graph-3d-container, .graph-container, canvas').first();
  221 |     await expect(graphContainer).toBeVisible({ timeout: 10000 });
  222 |   });
  223 | 
  224 |   test('should display general graph view at /graph', async ({ page }) => {
  225 |     // Navigate to general graph page
  226 |     await page.goto('http://localhost:5173/graph');
  227 |     await page.waitForTimeout(2000);
  228 |     
  229 |     // Verify no 404 error
  230 |     const error404 = page.locator('text=404, text=Not Found').first();
  231 |     const has404 = await error404.isVisible().catch(() => false);
  232 |     expect(has404).toBe(false);
  233 |     
  234 |     // Verify graph container or empty state is visible
  235 |     const graphContainer = page.locator('.graph-container, .graph-3d-container, canvas').first();
  236 |     const emptyState = page.locator('text=No notes found, text=No graph data').first();
  237 |     
  238 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  239 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  240 |     
> 241 |     expect(hasGraph || hasEmpty).toBe(true);
      |                                  ^ Error: expect(received).toBe(expected) // Object.is equality
  242 |   });
  243 | 
  244 |   test('should handle empty state when no notes exist', async ({ page, request }) => {
  245 |     // Clear all notes via API (if possible) or just check current state
  246 |     // This test verifies the empty state message
  247 |     
  248 |     const notesResponse = await request.get('http://localhost:8080/notes');
  249 |     const notesData = await notesResponse.json();
  250 |     
  251 |     if (notesData.total === 0 || notesData.notes?.length === 0) {
  252 |       // If truly empty, verify empty state
  253 |       await page.reload();
  254 |       await page.waitForTimeout(1000);
  255 |       
  256 |       const emptyState = page.locator('.empty-state, text=No notes').first();
  257 |       await expect(emptyState).toBeVisible();
  258 |     } else {
  259 |       // Skip this test if notes exist
  260 |       test.skip();
  261 |     }
  262 |   });
  263 | });
  264 | 
```