# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: home-page.spec.ts >> Home Page - Graph First >> should search notes from home page
- Location: tests\home-page.spec.ts:153:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.stats-bar').first()
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.stats-bar').first()

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e4]:
    - generic [ref=e5]:
      - button "2D" [ref=e6] [cursor=pointer]:
        - img [ref=e7]
        - generic [ref=e17]: 2D
      - button "3D" [ref=e18] [cursor=pointer]:
        - img [ref=e19]
        - generic [ref=e23]: 3D
      - button "List" [ref=e24] [cursor=pointer]:
        - img [ref=e25]
        - generic [ref=e26]: List
    - generic [ref=e27]:
      - button "🌌 All" [ref=e28] [cursor=pointer]:
        - generic [ref=e29]: 🌌
        - generic [ref=e30]: All
      - button "⭐ Stars" [ref=e31] [cursor=pointer]:
        - generic [ref=e32]: ⭐
        - generic [ref=e33]: Stars
      - button "🪐 Planets" [ref=e34] [cursor=pointer]:
        - generic [ref=e35]: 🪐
        - generic [ref=e36]: Planets
      - button "☄️ Comets" [ref=e37] [cursor=pointer]:
        - generic [ref=e38]: ☄️
        - generic [ref=e39]: Comets
      - button "🌀 Galaxies" [ref=e40] [cursor=pointer]:
        - generic [ref=e41]: 🌀
        - generic [ref=e42]: Galaxies
    - generic [ref=e43]:
      - textbox "Search notes..." [active] [ref=e44]: Searchable1776420727804
      - button "Search" [ref=e45] [cursor=pointer]:
        - img [ref=e46]
    - button "Menu" [ref=e50] [cursor=pointer]:
      - img [ref=e51]
    - button "Create new note" [ref=e52] [cursor=pointer]:
      - img [ref=e53]
  - generic [ref=e56]:
    - generic [ref=e57]:
      - strong [ref=e58]: "100"
      - text: nodes
    - generic [ref=e59]:
      - strong [ref=e60]: "2444"
      - text: links
```

# Test source

```ts
  75  |       await page.waitForTimeout(1500);
  76  |     }
  77  |     
  78  |     // Verify any content is visible (graph, list, loading, or error states)
  79  |     const content = page.locator('.fullscreen-graph, .list-container, .note-card, .loading-overlay, .error-overlay').first();
  80  |     await expect(content).toBeVisible({ timeout: 10000 });
  81  |   });
  82  | 
  83  |   test('should show note count in stats bar', async ({ page, request }) => {
  84  |     // Create test notes
  85  |     const timestamp = Date.now();
  86  |     await request.post('http://localhost:8080/notes', {
  87  |       data: { title: `Stats Test 1 ${timestamp}`, content: 'Content 1', type: 'star' }
  88  |     });
  89  |     await request.post('http://localhost:8080/notes', {
  90  |       data: { title: `Stats Test 2 ${timestamp}`, content: 'Content 2', type: 'planet' }
  91  |     });
  92  |     
  93  |     // Reload page and wait for network
  94  |     await page.reload();
  95  |     await page.waitForLoadState('networkidle');
  96  |     await page.waitForTimeout(3000);
  97  |     
  98  |     // Verify stats bar shows note count (or wait for loading to finish)
  99  |     const statsBar = page.locator('.stats-bar, .stats-total').first();
  100 |     
  101 |     // Wait for loading to finish
  102 |     await page.waitForTimeout(2000);
  103 |     
  104 |     const hasStats = await statsBar.isVisible().catch(() => false);
  105 |     if (hasStats) {
  106 |       // Check that count is greater than 0
  107 |       const statsText = await statsBar.textContent();
  108 |       const countMatch = statsText?.match(/(\d+)\s+note/);
  109 |       if (countMatch) {
  110 |         const count = parseInt(countMatch[1], 10);
  111 |         expect(count).toBeGreaterThan(0);
  112 |       }
  113 |     }
  114 |     // If stats not visible, test passes if page loaded without errors
  115 |     expect(true).toBe(true);
  116 |   });
  117 | 
  118 |   test('should filter notes by type from home page', async ({ page, request }) => {
  119 |     // Create notes of different types using helper
  120 |     const timestamp = Date.now();
  121 |     await createNote(request, {
  122 |       title: `Star Note ${timestamp}`,
  123 |       content: 'Star content',
  124 |       type: 'star'
  125 |     });
  126 |     await createNote(request, {
  127 |       title: `Planet Note ${timestamp}`,
  128 |       content: 'Planet content',
  129 |       type: 'planet'
  130 |     });
  131 |     
  132 |     // Reload page
  133 |     await page.reload();
  134 |     await page.waitForTimeout(2000);
  135 |     
  136 |     // Click on "Stars" filter
  137 |     const starsFilter = page.locator('button:has-text("⭐"), button:has-text("Stars"), [data-filter="star"]').first();
  138 |     if (await starsFilter.isVisible().catch(() => false)) {
  139 |       await starsFilter.click();
  140 |       await page.waitForTimeout(500);
  141 |       
  142 |       // Verify filter is applied (stats should show filtered count)
  143 |       const statsFilter = page.locator('.stats-filter').first();
  144 |       const hasFilterText = await statsFilter.isVisible().catch(() => false);
  145 |       
  146 |       if (hasFilterText) {
  147 |         const filterText = await statsFilter.textContent();
  148 |         expect(filterText?.toLowerCase()).toContain('filter');
  149 |       }
  150 |     }
  151 |   });
  152 | 
  153 |   test('should search notes from home page', async ({ page, request }) => {
  154 |     // Create a searchable note using helper
  155 |     const timestamp = Date.now();
  156 |     const searchTerm = `Searchable${timestamp}`;
  157 |     await createNote(request, {
  158 |       title: `Test ${searchTerm} Note`,
  159 |       content: 'Test content',
  160 |       type: 'star'
  161 |     });
  162 |     
  163 |     // Reload page
  164 |     await page.reload();
  165 |     await page.waitForTimeout(2000);
  166 |     
  167 |     // Fill search input
  168 |     const searchInput = page.locator('.search-input, input[type="search"]').first();
  169 |     if (await searchInput.isVisible().catch(() => false)) {
  170 |       await searchInput.fill(searchTerm);
  171 |       await page.waitForTimeout(1000); // Wait for search to apply
  172 |       
  173 |       // Verify stats bar appears
  174 |       const statsBar = page.locator('.stats-bar').first();
> 175 |       await expect(statsBar).toBeVisible();
      |                              ^ Error: expect(locator).toBeVisible() failed
  176 |     }
  177 |   });
  178 | 
  179 |   test('should open side panel when clicking on graph node', async ({ page, request }) => {
  180 |     // Create a note using helper
  181 |     const timestamp = Date.now();
  182 |     await createNote(request, {
  183 |       title: `Side Panel Test ${timestamp}`,
  184 |       content: 'Test content for side panel',
  185 |       type: 'star'
  186 |     });
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
  204 |     // Create a note using helper
  205 |     const timestamp = Date.now();
  206 |     const note = await createNote(request, {
  207 |       title: `Graph View Test ${timestamp}`,
  208 |       content: 'Test content',
  209 |       type: 'star'
  210 |     });
  211 |     const noteId = note.id;
  212 | 
  213 |     // Navigate to specific graph page
  214 |     await page.goto(`/graph/${noteId}`);
  215 |     await page.waitForTimeout(2000);
  216 |     
  217 |     // Verify graph container is visible
  218 |     const graphContainer = page.locator('.graph-3d-container, .fullscreen-graph, canvas').first();
  219 |     await expect(graphContainer).toBeVisible({ timeout: 10000 });
  220 |   });
  221 | 
  222 |   test('should display general graph view at /graph', async ({ page }) => {
  223 |     // Navigate to general graph page
  224 |     await page.goto('/graph');
  225 |     await page.waitForLoadState('networkidle');
  226 |     await page.waitForTimeout(3000);
  227 |     
  228 |     // Verify no 404 error
  229 |     const error404 = page.locator('text=404, text=Not Found').first();
  230 |     const has404 = await error404.isVisible().catch(() => false);
  231 |     expect(has404).toBe(false);
  232 |     
  233 |     // Verify graph container, empty state, or error state is visible
  234 |     const graphContainer = page.locator('.fullscreen-graph, .graph-3d-container, canvas').first();
  235 |     const emptyState = page.locator('text=No notes found, text=No graph data').first();
  236 |     const errorState = page.locator('text=Failed to load graph data').first();
  237 |     
  238 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  239 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  240 |     const hasError = await errorState.isVisible().catch(() => false);
  241 |     
  242 |     expect(hasGraph || hasEmpty || hasError).toBe(true);
  243 |   });
  244 | 
  245 |   test('should handle empty state when no notes exist', async ({ page, request }) => {
  246 |     // Check current notes count
  247 |     const notesResponse = await request.get('http://localhost:8080/notes');
  248 |     const notesData = await notesResponse.json();
  249 |     const hasNotes = notesData.total > 0 || (notesData.notes?.length > 0);
  250 |     
  251 |     // Reload page
  252 |     await page.reload();
  253 |     await page.waitForTimeout(2000);
  254 |     
  255 |     if (!hasNotes) {
  256 |       // If no notes, verify some content is visible (empty state, graph container, or error)
  257 |       const content = page.locator('.fullscreen-graph, .list-container, .empty-state, .loading-overlay, .error-overlay, text=/No notes|empty|Loading/i').first();
  258 |       await expect(content).toBeVisible({ timeout: 10000 });
  259 |     }
  260 |     // If notes exist, test passes - we just verify the page loads
  261 |     expect(true).toBe(true);
  262 |   });
  263 | 
  264 |   test('should toggle full graph mode on home page', async ({ page, request }) => {
  265 |     // Create test notes if needed
  266 |     const notesResponse = await request.get('http://localhost:8080/notes');
  267 |     const notesData = await notesResponse.json();
  268 |     
  269 |     if (notesData.total < 2) {
  270 |       // Create at least 2 notes for meaningful test
  271 |       await request.post('http://localhost:8080/notes', {
  272 |         data: { title: 'Note 1', content: 'Content 1', type: 'star' }
  273 |       });
  274 |       await request.post('http://localhost:8080/notes', {
  275 |         data: { title: 'Note 2', content: 'Content 2', type: 'planet' }
```