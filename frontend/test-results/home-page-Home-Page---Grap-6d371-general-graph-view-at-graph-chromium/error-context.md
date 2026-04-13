# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: home-page.spec.ts >> Home Page - Graph First >> should display general graph view at /graph
- Location: tests\home-page.spec.ts:225:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/graph
Call log:
  - navigating to "http://localhost:5173/graph", waiting until "load"

```

# Test source

```ts
  127 |     // Reload page
  128 |     await page.reload();
  129 |     await page.waitForTimeout(2000);
  130 |     
  131 |     // Click on "Stars" filter
  132 |     const starsFilter = page.locator('button:has-text("⭐"), button:has-text("Stars"), [data-filter="star"]').first();
  133 |     if (await starsFilter.isVisible().catch(() => false)) {
  134 |       await starsFilter.click();
  135 |       await page.waitForTimeout(500);
  136 |       
  137 |       // Verify filter is applied (stats should show filtered count)
  138 |       const statsFilter = page.locator('.stats-filter').first();
  139 |       const hasFilterText = await statsFilter.isVisible().catch(() => false);
  140 |       
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
  170 |       // Verify search filter indicator appears
  171 |       const searchIndicator = page.locator('text=search:, .stats-filter').first();
  172 |       const statsBar = page.locator('.stats-bar').first();
  173 |       await expect(statsBar).toBeVisible();
  174 |     }
  175 |   });
  176 | 
  177 |   test('should open side panel when clicking on graph node', async ({ page, request }) => {
  178 |     // Create a note
  179 |     const timestamp = Date.now();
  180 |     const note = await request.post('http://localhost:8080/notes', {
  181 |       data: { 
  182 |         title: `Side Panel Test ${timestamp}`, 
  183 |         content: 'Test content for side panel',
  184 |         type: 'star'
  185 |       }
  186 |     });
  187 |     const noteId = (await note.json()).id;
  188 |     
  189 |     // Reload page
  190 |     await page.reload();
  191 |     await page.waitForTimeout(2000);
  192 |     
  193 |     // Try to click on a note card (fallback if graph click doesn't work)
  194 |     const noteCard = page.locator('.note-card').first();
  195 |     if (await noteCard.isVisible().catch(() => false)) {
  196 |       await noteCard.click();
  197 |       
  198 |       // Verify side panel opens
  199 |       const sidePanel = page.locator('.side-panel, .note-side-panel').first();
  200 |       await expect(sidePanel).toBeVisible({ timeout: 5000 });
  201 |     }
  202 |   });
  203 | 
  204 |   test('should navigate to graph view for specific note', async ({ page, request }) => {
  205 |     // Create a note
  206 |     const timestamp = Date.now();
  207 |     const note = await request.post('http://localhost:8080/notes', {
  208 |       data: { 
  209 |         title: `Graph View Test ${timestamp}`, 
  210 |         content: 'Test content',
  211 |         type: 'star'
  212 |       }
  213 |     });
  214 |     const noteId = (await note.json()).id;
  215 |     
  216 |     // Navigate to specific graph page
  217 |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  218 |     await page.waitForTimeout(2000);
  219 |     
  220 |     // Verify graph container is visible
  221 |     const graphContainer = page.locator('.graph-3d-container, .graph-container, canvas').first();
  222 |     await expect(graphContainer).toBeVisible({ timeout: 10000 });
  223 |   });
  224 | 
  225 |   test('should display general graph view at /graph', async ({ page }) => {
  226 |     // Navigate to general graph page
> 227 |     await page.goto('http://localhost:5173/graph');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/graph
  228 |     await page.waitForTimeout(2000);
  229 |     
  230 |     // Verify no 404 error
  231 |     const error404 = page.locator('text=404, text=Not Found').first();
  232 |     const has404 = await error404.isVisible().catch(() => false);
  233 |     expect(has404).toBe(false);
  234 |     
  235 |     // Verify graph container or empty state is visible
  236 |     const graphContainer = page.locator('.graph-container, .graph-3d-container, canvas').first();
  237 |     const emptyState = page.locator('text=No notes found, text=No graph data').first();
  238 |     
  239 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  240 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  241 |     
  242 |     expect(hasGraph || hasEmpty).toBe(true);
  243 |   });
  244 | 
  245 |   test('should handle empty state when no notes exist', async ({ page, request }) => {
  246 |     // Clear all notes via API (if possible) or just check current state
  247 |     // This test verifies the empty state message
  248 |     
  249 |     const notesResponse = await request.get('http://localhost:8080/notes');
  250 |     const notesData = await notesResponse.json();
  251 |     
  252 |     if (notesData.total === 0 || notesData.notes?.length === 0) {
  253 |       // If truly empty, verify empty state
  254 |       await page.reload();
  255 |       await page.waitForTimeout(1000);
  256 |       
  257 |       const emptyState = page.locator('.empty-state, text=No notes').first();
  258 |       await expect(emptyState).toBeVisible();
  259 |     } else {
  260 |       // Skip this test if notes exist
  261 |       test.skip();
  262 |     }
  263 |   });
  264 | });
  265 | 
```