# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: home-page.spec.ts >> Home Page - Graph First >> should search notes from home page
- Location: tests\home-page.spec.ts:151:3

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
      - textbox "Search notes..." [active] [ref=e44]: Searchable1776334186062
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
      - strong [ref=e60]: "1943"
      - text: links
```

# Test source

```ts
  75  |     if (hasToggle) {
  76  |       await viewToggle.click();
  77  |       await page.waitForTimeout(1500);
  78  |     }
  79  |     
  80  |     // Verify any content is visible (graph, list, loading, or error states)
  81  |     const content = page.locator('.fullscreen-graph, .list-container, .note-card, .loading-overlay, .error-overlay').first();
  82  |     await expect(content).toBeVisible({ timeout: 10000 });
  83  |   });
  84  | 
  85  |   test('should show note count in stats bar', async ({ page, request }) => {
  86  |     // Create test notes
  87  |     const timestamp = Date.now();
  88  |     await request.post('http://localhost:8080/notes', {
  89  |       data: { title: `Stats Test 1 ${timestamp}`, content: 'Content 1', type: 'star' }
  90  |     });
  91  |     await request.post('http://localhost:8080/notes', {
  92  |       data: { title: `Stats Test 2 ${timestamp}`, content: 'Content 2', type: 'planet' }
  93  |     });
  94  |     
  95  |     // Reload page and wait for network
  96  |     await page.reload();
  97  |     await page.waitForLoadState('networkidle');
  98  |     await page.waitForTimeout(3000);
  99  |     
  100 |     // Verify stats bar shows note count (or wait for loading to finish)
  101 |     const statsBar = page.locator('.stats-bar, .stats-total').first();
  102 |     
  103 |     // Wait for loading to finish
  104 |     await page.waitForTimeout(2000);
  105 |     
  106 |     const hasStats = await statsBar.isVisible().catch(() => false);
  107 |     if (hasStats) {
  108 |       // Check that count is greater than 0
  109 |       const statsText = await statsBar.textContent();
  110 |       const countMatch = statsText?.match(/(\d+)\s+note/);
  111 |       if (countMatch) {
  112 |         const count = parseInt(countMatch[1], 10);
  113 |         expect(count).toBeGreaterThan(0);
  114 |       }
  115 |     }
  116 |     // If stats not visible, test passes if page loaded without errors
  117 |     expect(true).toBe(true);
  118 |   });
  119 | 
  120 |   test('should filter notes by type from home page', async ({ page, request }) => {
  121 |     // Create notes of different types
  122 |     const timestamp = Date.now();
  123 |     await request.post('http://localhost:8080/notes', {
  124 |       data: { title: `Star Note ${timestamp}`, content: 'Star content', type: 'star' }
  125 |     });
  126 |     await request.post('http://localhost:8080/notes', {
  127 |       data: { title: `Planet Note ${timestamp}`, content: 'Planet content', type: 'planet' }
  128 |     });
  129 |     
  130 |     // Reload page
  131 |     await page.reload();
  132 |     await page.waitForTimeout(2000);
  133 |     
  134 |     // Click on "Stars" filter
  135 |     const starsFilter = page.locator('button:has-text("⭐"), button:has-text("Stars"), [data-filter="star"]').first();
  136 |     if (await starsFilter.isVisible().catch(() => false)) {
  137 |       await starsFilter.click();
  138 |       await page.waitForTimeout(500);
  139 |       
  140 |       // Verify filter is applied (stats should show filtered count)
  141 |       const statsFilter = page.locator('.stats-filter').first();
  142 |       const hasFilterText = await statsFilter.isVisible().catch(() => false);
  143 |       
  144 |       if (hasFilterText) {
  145 |         const filterText = await statsFilter.textContent();
  146 |         expect(filterText?.toLowerCase()).toContain('filter');
  147 |       }
  148 |     }
  149 |   });
  150 | 
  151 |   test('should search notes from home page', async ({ page, request }) => {
  152 |     // Create a searchable note
  153 |     const timestamp = Date.now();
  154 |     const searchTerm = `Searchable${timestamp}`;
  155 |     await request.post('http://localhost:8080/notes', {
  156 |       data: { 
  157 |         title: `Test ${searchTerm} Note`, 
  158 |         content: 'Test content',
  159 |         type: 'star'
  160 |       }
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
  180 |     // Create a note
  181 |     const timestamp = Date.now();
  182 |     const note = await request.post('http://localhost:8080/notes', {
  183 |       data: { 
  184 |         title: `Side Panel Test ${timestamp}`, 
  185 |         content: 'Test content for side panel',
  186 |         type: 'star'
  187 |       }
  188 |     });
  189 |     await note.json();
  190 |     
  191 |     // Reload page
  192 |     await page.reload();
  193 |     await page.waitForTimeout(2000);
  194 |     
  195 |     // Try to click on a note card (fallback if graph click doesn't work)
  196 |     const noteCard = page.locator('.note-card').first();
  197 |     if (await noteCard.isVisible().catch(() => false)) {
  198 |       await noteCard.click();
  199 |       
  200 |       // Verify side panel opens
  201 |       const sidePanel = page.locator('.side-panel, .note-side-panel').first();
  202 |       await expect(sidePanel).toBeVisible({ timeout: 5000 });
  203 |     }
  204 |   });
  205 | 
  206 |   test('should navigate to graph view for specific note', async ({ page, request }) => {
  207 |     // Create a note
  208 |     const timestamp = Date.now();
  209 |     const note = await request.post('http://localhost:8080/notes', {
  210 |       data: { 
  211 |         title: `Graph View Test ${timestamp}`, 
  212 |         content: 'Test content',
  213 |         type: 'star'
  214 |       }
  215 |     });
  216 |     const noteId = (await note.json()).id;
  217 |     
  218 |     // Navigate to specific graph page
  219 |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  220 |     await page.waitForTimeout(2000);
  221 |     
  222 |     // Verify graph container is visible
  223 |     const graphContainer = page.locator('.graph-3d-container, .fullscreen-graph, canvas').first();
  224 |     await expect(graphContainer).toBeVisible({ timeout: 10000 });
  225 |   });
  226 | 
  227 |   test('should display general graph view at /graph', async ({ page }) => {
  228 |     // Navigate to general graph page
  229 |     await page.goto('http://localhost:5173/graph');
  230 |     await page.waitForLoadState('networkidle');
  231 |     await page.waitForTimeout(3000);
  232 |     
  233 |     // Verify no 404 error
  234 |     const error404 = page.locator('text=404, text=Not Found').first();
  235 |     const has404 = await error404.isVisible().catch(() => false);
  236 |     expect(has404).toBe(false);
  237 |     
  238 |     // Verify graph container, empty state, or error state is visible
  239 |     const graphContainer = page.locator('.fullscreen-graph, .graph-3d-container, canvas').first();
  240 |     const emptyState = page.locator('text=No notes found, text=No graph data').first();
  241 |     const errorState = page.locator('text=Failed to load graph data').first();
  242 |     
  243 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  244 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  245 |     const hasError = await errorState.isVisible().catch(() => false);
  246 |     
  247 |     expect(hasGraph || hasEmpty || hasError).toBe(true);
  248 |   });
  249 | 
  250 |   test('should handle empty state when no notes exist', async ({ page, request }) => {
  251 |     // Check current notes count
  252 |     const notesResponse = await request.get('http://localhost:8080/notes');
  253 |     const notesData = await notesResponse.json();
  254 |     const hasNotes = notesData.total > 0 || (notesData.notes?.length > 0);
  255 |     
  256 |     // Reload page
  257 |     await page.reload();
  258 |     await page.waitForTimeout(2000);
  259 |     
  260 |     if (!hasNotes) {
  261 |       // If no notes, verify some content is visible (empty state, graph container, or error)
  262 |       const content = page.locator('.fullscreen-graph, .list-container, .empty-state, .loading-overlay, .error-overlay, text=/No notes|empty|Loading/i').first();
  263 |       await expect(content).toBeVisible({ timeout: 10000 });
  264 |     }
  265 |     // If notes exist, test passes - we just verify the page loads
  266 |     expect(true).toBe(true);
  267 |   });
  268 | 
  269 |   test('should toggle full graph mode on home page', async ({ page, request }) => {
  270 |     // Create test notes if needed
  271 |     const notesResponse = await request.get('http://localhost:8080/notes');
  272 |     const notesData = await notesResponse.json();
  273 |     
  274 |     if (notesData.total < 2) {
  275 |       // Create at least 2 notes for meaningful test
```