# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: home-page.spec.ts >> Home Page - Graph First >> should search notes from home page
- Location: tests\home-page.spec.ts:157:3

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
      - textbox "Search notes..." [active] [ref=e44]: Searchable1776441631093
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
  79  |     const content = page.locator('[data-testid="graph-2d-container"], [data-testid="list-container"], .note-card, [data-testid="loading-overlay"], .error-overlay').first();
  80  |     await expect(content).toBeVisible({ timeout: 10000 });
  81  |   });
  82  | 
  83  |   test('should show note count in stats bar', async ({ page, request }) => {
  84  |     // Create test notes using helper
  85  |     const timestamp = Date.now();
  86  |     await createNote(request, {
  87  |       title: `Stats Test 1 ${timestamp}`,
  88  |       content: 'Content 1',
  89  |       type: 'star'
  90  |     });
  91  |     await createNote(request, {
  92  |       title: `Stats Test 2 ${timestamp}`,
  93  |       content: 'Content 2',
  94  |       type: 'planet'
  95  |     });
  96  |     
  97  |     // Reload page and wait for network
  98  |     await page.reload();
  99  |     await page.waitForLoadState('networkidle');
  100 |     await page.waitForTimeout(3000);
  101 |     
  102 |     // Verify stats bar shows note count (or wait for loading to finish)
  103 |     const statsBar = page.locator('.stats-bar, .stats-total').first();
  104 |     
  105 |     // Wait for loading to finish
  106 |     await page.waitForTimeout(2000);
  107 |     
  108 |     const hasStats = await statsBar.isVisible().catch(() => false);
  109 |     if (hasStats) {
  110 |       // Check that count is greater than 0
  111 |       const statsText = await statsBar.textContent();
  112 |       const countMatch = statsText?.match(/(\d+)\s+note/);
  113 |       if (countMatch) {
  114 |         const count = parseInt(countMatch[1], 10);
  115 |         expect(count).toBeGreaterThan(0);
  116 |       }
  117 |     }
  118 |     // If stats not visible, test passes if page loaded without errors
  119 |     expect(true).toBe(true);
  120 |   });
  121 | 
  122 |   test('should filter notes by type from home page', async ({ page, request }) => {
  123 |     // Create notes of different types using helper
  124 |     const timestamp = Date.now();
  125 |     await createNote(request, {
  126 |       title: `Star Note ${timestamp}`,
  127 |       content: 'Star content',
  128 |       type: 'star'
  129 |     });
  130 |     await createNote(request, {
  131 |       title: `Planet Note ${timestamp}`,
  132 |       content: 'Planet content',
  133 |       type: 'planet'
  134 |     });
  135 |     
  136 |     // Reload page
  137 |     await page.reload();
  138 |     await page.waitForTimeout(2000);
  139 |     
  140 |     // Click on "Stars" filter
  141 |     const starsFilter = page.locator('button:has-text("⭐"), button:has-text("Stars"), [data-filter="star"]').first();
  142 |     if (await starsFilter.isVisible().catch(() => false)) {
  143 |       await starsFilter.click();
  144 |       await page.waitForTimeout(500);
  145 |       
  146 |       // Verify filter is applied (stats should show filtered count)
  147 |       const statsFilter = page.locator('.stats-filter').first();
  148 |       const hasFilterText = await statsFilter.isVisible().catch(() => false);
  149 |       
  150 |       if (hasFilterText) {
  151 |         const filterText = await statsFilter.textContent();
  152 |         expect(filterText?.toLowerCase()).toContain('filter');
  153 |       }
  154 |     }
  155 |   });
  156 | 
  157 |   test('should search notes from home page', async ({ page, request }) => {
  158 |     // Create a searchable note using helper
  159 |     const timestamp = Date.now();
  160 |     const searchTerm = `Searchable${timestamp}`;
  161 |     await createNote(request, {
  162 |       title: `Test ${searchTerm} Note`,
  163 |       content: 'Test content',
  164 |       type: 'star'
  165 |     });
  166 |     
  167 |     // Reload page
  168 |     await page.reload();
  169 |     await page.waitForTimeout(2000);
  170 |     
  171 |     // Fill search input
  172 |     const searchInput = page.locator('.search-input, input[type="search"]').first();
  173 |     if (await searchInput.isVisible().catch(() => false)) {
  174 |       await searchInput.fill(searchTerm);
  175 |       await page.waitForTimeout(1000); // Wait for search to apply
  176 |       
  177 |       // Verify stats bar appears
  178 |       const statsBar = page.locator('.stats-bar').first();
> 179 |       await expect(statsBar).toBeVisible();
      |                              ^ Error: expect(locator).toBeVisible() failed
  180 |     }
  181 |   });
  182 | 
  183 |   test('should open side panel when clicking on graph node', async ({ page, request }) => {
  184 |     // Create a note using helper
  185 |     const timestamp = Date.now();
  186 |     await createNote(request, {
  187 |       title: `Side Panel Test ${timestamp}`,
  188 |       content: 'Test content for side panel',
  189 |       type: 'star'
  190 |     });
  191 |     
  192 |     // Reload page
  193 |     await page.reload();
  194 |     await page.waitForTimeout(2000);
  195 |     
  196 |     // Try to click on a note card (fallback if graph click doesn't work)
  197 |     const noteCard = page.locator('.note-card').first();
  198 |     if (await noteCard.isVisible().catch(() => false)) {
  199 |       await noteCard.click();
  200 |       
  201 |       // Verify side panel opens
  202 |       const sidePanel = page.locator('.side-panel, .note-side-panel').first();
  203 |       await expect(sidePanel).toBeVisible({ timeout: 5000 });
  204 |     }
  205 |   });
  206 | 
  207 |   test('should navigate to graph view for specific note', async ({ page, request }) => {
  208 |     // Create a note using helper
  209 |     const timestamp = Date.now();
  210 |     const note = await createNote(request, {
  211 |       title: `Graph View Test ${timestamp}`,
  212 |       content: 'Test content',
  213 |       type: 'star'
  214 |     });
  215 |     const noteId = note.id;
  216 | 
  217 |     // Navigate to specific graph page
  218 |     await page.goto(`/graph/${noteId}`);
  219 |     await page.waitForTimeout(2000);
  220 |     
  221 |     // Verify graph container is visible
  222 |     const graphContainer = page.locator('.graph-3d-container, .fullscreen-graph, canvas').first();
  223 |     await expect(graphContainer).toBeVisible({ timeout: 10000 });
  224 |   });
  225 | 
  226 |   test('should display general graph view at /graph', async ({ page }) => {
  227 |     // Navigate to general graph page
  228 |     await page.goto('/graph');
  229 |     await page.waitForLoadState('networkidle');
  230 |     await page.waitForTimeout(3000);
  231 |     
  232 |     // Verify no 404 error
  233 |     const error404 = page.locator('text=404, text=Not Found').first();
  234 |     const has404 = await error404.isVisible().catch(() => false);
  235 |     expect(has404).toBe(false);
  236 |     
  237 |     // Verify graph container, empty state, or error state is visible
  238 |     const graphContainer = page.locator('.fullscreen-graph, .graph-3d-container, canvas').first();
  239 |     const emptyState = page.locator('text=No notes found, text=No graph data').first();
  240 |     const errorState = page.locator('text=Failed to load graph data').first();
  241 |     
  242 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  243 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  244 |     const hasError = await errorState.isVisible().catch(() => false);
  245 |     
  246 |     expect(hasGraph || hasEmpty || hasError).toBe(true);
  247 |   });
  248 | 
  249 |   test('should handle empty state when no notes exist', async ({ page, request }) => {
  250 |     // Check current notes count
  251 |     const notesResponse = await request.get('http://localhost:8080/notes');
  252 |     const notesData = await notesResponse.json();
  253 |     const hasNotes = notesData.total > 0 || (notesData.notes?.length > 0);
  254 |     
  255 |     // Reload page
  256 |     await page.reload();
  257 |     await page.waitForTimeout(2000);
  258 |     
  259 |     if (!hasNotes) {
  260 |       // If no notes, verify some content is visible (empty state, graph container, or error)
  261 |       const content = page.locator('.fullscreen-graph, .list-container, .empty-state, .loading-overlay, .error-overlay, text=/No notes|empty|Loading/i').first();
  262 |       await expect(content).toBeVisible({ timeout: 10000 });
  263 |     }
  264 |     // If notes exist, test passes - we just verify the page loads
  265 |     expect(true).toBe(true);
  266 |   });
  267 | 
  268 |   test('should toggle full graph mode on home page', async ({ page, request }) => {
  269 |     // Create test notes if needed
  270 |     const notesResponse = await request.get('http://localhost:8080/notes');
  271 |     const notesData = await notesResponse.json();
  272 |     
  273 |     if (notesData.total < 2) {
  274 |       // Create at least 2 notes for meaningful test
  275 |       await request.post('http://localhost:8080/notes', {
  276 |         data: { title: 'Note 1', content: 'Content 1', type: 'star' }
  277 |       });
  278 |       await request.post('http://localhost:8080/notes', {
  279 |         data: { title: 'Note 2', content: 'Content 2', type: 'planet' }
```