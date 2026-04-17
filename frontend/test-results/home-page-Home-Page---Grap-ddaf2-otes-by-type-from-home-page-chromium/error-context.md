# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: home-page.spec.ts >> Home Page - Graph First >> should filter notes by type from home page
- Location: tests\home-page.spec.ts:126:3

# Error details

```
Error: expect(received).toContain(expected) // indexOf

Expected substring: "filter"
Received string:    "20 nodes 0 links stars"
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
      - button "⭐ Stars" [active] [ref=e31] [cursor=pointer]:
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
      - textbox "Search notes..." [ref=e44]
      - button "Search" [ref=e45] [cursor=pointer]:
        - img [ref=e46]
    - button "Menu" [ref=e50] [cursor=pointer]:
      - img [ref=e51]
    - button "Create new note" [ref=e52] [cursor=pointer]:
      - img [ref=e53]
  - generic [ref=e56]:
    - generic [ref=e57]:
      - strong [ref=e58]: "20"
      - text: nodes
    - generic [ref=e59]:
      - strong [ref=e60]: "0"
      - text: links
    - generic [ref=e61]: Stars
```

# Test source

```ts
  56  |       const graphCanvas = page.locator('[data-testid="graph-2d-container"] canvas, canvas').first();
  57  |       await expect(graphCanvas).toBeVisible();
  58  |     }
  59  |   });
  60  | 
  61  |   test('should display list view when toggled from graph view', async ({ page }) => {
  62  |     // Create a note first
  63  |     await page.click('.create-btn, button:has-text("+")');
  64  |     await page.fill('input[name="title"]', 'List View Test Note');
  65  |     await page.fill('textarea[name="content"]', 'Content for list view test');
  66  |     await page.click('button[type="submit"]');
  67  |     await page.waitForTimeout(1500);
  68  |     
  69  |     // Reload to ensure we're on fresh home page
  70  |     await page.goto('/');
  71  |     await page.waitForTimeout(3000); // Wait for page to load
  72  |     
  73  |     // Toggle to list view (using FloatingControls or view toggle)
  74  |     const viewToggle = page.locator('button:has-text("List"), button:has-text("View"), .view-toggle').first();
  75  |     const hasToggle = await viewToggle.isVisible().catch(() => false);
  76  |     
  77  |     if (hasToggle) {
  78  |       await viewToggle.click();
  79  |       await page.waitForTimeout(1500);
  80  |     }
  81  |     
  82  |     // Verify any content is visible (graph, list, loading, or error states)
  83  |     const content = page.locator('[data-testid="graph-2d-container"], [data-testid="list-container"], .note-card, [data-testid="loading-overlay"], .error-overlay').first();
  84  |     await expect(content).toBeVisible({ timeout: 10000 });
  85  |   });
  86  | 
  87  |   test('should show note count in stats bar', async ({ page, request }) => {
  88  |     // Create test notes using helper
  89  |     const timestamp = Date.now();
  90  |     await createNote(request, {
  91  |       title: `Stats Test 1 ${timestamp}`,
  92  |       content: 'Content 1',
  93  |       type: 'star'
  94  |     });
  95  |     await createNote(request, {
  96  |       title: `Stats Test 2 ${timestamp}`,
  97  |       content: 'Content 2',
  98  |       type: 'planet'
  99  |     });
  100 |     
  101 |     // Reload page and wait for network
  102 |     await page.reload();
  103 |     await page.waitForLoadState('networkidle');
  104 |     await page.waitForTimeout(3000);
  105 |     
  106 |     // Verify stats bar shows note count (or wait for loading to finish)
  107 |     const statsBar = page.locator('[data-testid="graph-stats"]').first();
  108 |     
  109 |     // Wait for loading to finish
  110 |     await page.waitForTimeout(2000);
  111 |     
  112 |     const hasStats = await statsBar.isVisible().catch(() => false);
  113 |     if (hasStats) {
  114 |       // Check that count is greater than 0
  115 |       const statsText = await statsBar.textContent();
  116 |       const countMatch = statsText?.match(/(\d+)\s+note/);
  117 |       if (countMatch) {
  118 |         const count = parseInt(countMatch[1], 10);
  119 |         expect(count).toBeGreaterThan(0);
  120 |       }
  121 |     }
  122 |     // If stats not visible, test passes if page loaded without errors
  123 |     expect(true).toBe(true);
  124 |   });
  125 | 
  126 |   test('should filter notes by type from home page', async ({ page, request }) => {
  127 |     // Create notes of different types using helper
  128 |     const timestamp = Date.now();
  129 |     await createNote(request, {
  130 |       title: `Star Note ${timestamp}`,
  131 |       content: 'Star content',
  132 |       type: 'star'
  133 |     });
  134 |     await createNote(request, {
  135 |       title: `Planet Note ${timestamp}`,
  136 |       content: 'Planet content',
  137 |       type: 'planet'
  138 |     });
  139 |     
  140 |     // Reload page
  141 |     await page.reload();
  142 |     await page.waitForTimeout(2000);
  143 |     
  144 |     // Click on "Stars" filter
  145 |     const starsFilter = page.locator('button:has-text("⭐"), button:has-text("Stars"), [data-filter="star"]').first();
  146 |     if (await starsFilter.isVisible().catch(() => false)) {
  147 |       await starsFilter.click();
  148 |       await page.waitForTimeout(500);
  149 |       
  150 |       // Verify filter is applied (stats should show filtered count)
  151 |       const statsFilter = page.locator('[data-testid="graph-stats"]').first();
  152 |       const hasFilterText = await statsFilter.isVisible().catch(() => false);
  153 |       
  154 |       if (hasFilterText) {
  155 |         const filterText = await statsFilter.textContent();
> 156 |         expect(filterText?.toLowerCase()).toContain('filter');
      |                                           ^ Error: expect(received).toContain(expected) // indexOf
  157 |       }
  158 |     }
  159 |   });
  160 | 
  161 |   test('should search notes from home page', async ({ page, request }) => {
  162 |     // Create a searchable note using helper
  163 |     const timestamp = Date.now();
  164 |     const searchTerm = `Searchable${timestamp}`;
  165 |     await createNote(request, {
  166 |       title: `Test ${searchTerm} Note`,
  167 |       content: 'Test content',
  168 |       type: 'star'
  169 |     });
  170 |     
  171 |     // Reload page
  172 |     await page.reload();
  173 |     await page.waitForTimeout(2000);
  174 |     
  175 |     // Fill search input
  176 |     const searchInput = page.locator('.search-input, input[type="search"]').first();
  177 |     if (await searchInput.isVisible().catch(() => false)) {
  178 |       await searchInput.fill(searchTerm);
  179 |       await page.waitForTimeout(1000); // Wait for search to apply
  180 |       
  181 |       // Verify stats bar appears
  182 |       const statsBar = page.locator('[data-testid="graph-stats"]').first();
  183 |       await expect(statsBar).toBeVisible();
  184 |     }
  185 |   });
  186 | 
  187 |   test('should open side panel when clicking on graph node', async ({ page, request }) => {
  188 |     // Create a note using helper
  189 |     const timestamp = Date.now();
  190 |     await createNote(request, {
  191 |       title: `Side Panel Test ${timestamp}`,
  192 |       content: 'Test content for side panel',
  193 |       type: 'star'
  194 |     });
  195 |     
  196 |     // Reload page
  197 |     await page.reload();
  198 |     await page.waitForTimeout(2000);
  199 |     
  200 |     // Try to click on a note card (fallback if graph click doesn't work)
  201 |     const noteCard = page.locator('.note-card').first();
  202 |     if (await noteCard.isVisible().catch(() => false)) {
  203 |       await noteCard.click();
  204 |       
  205 |       // Verify side panel opens
  206 |       const sidePanel = page.locator('.side-panel, .note-side-panel').first();
  207 |       await expect(sidePanel).toBeVisible({ timeout: 5000 });
  208 |     }
  209 |   });
  210 | 
  211 |   test('should navigate to graph view for specific note', async ({ page, request }) => {
  212 |     // Create a note using helper
  213 |     const timestamp = Date.now();
  214 |     const note = await createNote(request, {
  215 |       title: `Graph View Test ${timestamp}`,
  216 |       content: 'Test content',
  217 |       type: 'star'
  218 |     });
  219 |     const noteId = note.id;
  220 | 
  221 |     // Navigate to specific graph page
  222 |     await page.goto(`/graph/${noteId}`);
  223 |     await page.waitForTimeout(2000);
  224 |     
  225 |     // Verify graph container is visible
  226 |     const graphContainer = page.locator('.graph-3d-container, .fullscreen-graph, canvas').first();
  227 |     await expect(graphContainer).toBeVisible({ timeout: 10000 });
  228 |   });
  229 | 
  230 |   test('should display general graph view at /graph', async ({ page }) => {
  231 |     // Navigate to general graph page
  232 |     await page.goto('/graph');
  233 |     await page.waitForLoadState('networkidle');
  234 |     await page.waitForTimeout(3000);
  235 |     
  236 |     // Verify no 404 error
  237 |     const error404 = page.locator('text=404, text=Not Found').first();
  238 |     const has404 = await error404.isVisible().catch(() => false);
  239 |     expect(has404).toBe(false);
  240 |     
  241 |     // Verify graph container, empty state, or error state is visible
  242 |     const graphContainer = page.locator('.fullscreen-graph, .graph-3d-container, canvas').first();
  243 |     const emptyState = page.locator('text=No notes found, text=No graph data').first();
  244 |     const errorState = page.locator('text=Failed to load graph data').first();
  245 |     
  246 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  247 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  248 |     const hasError = await errorState.isVisible().catch(() => false);
  249 |     
  250 |     expect(hasGraph || hasEmpty || hasError).toBe(true);
  251 |   });
  252 | 
  253 |   test('should handle empty state when no notes exist', async ({ page, request }) => {
  254 |     // Check current notes count
  255 |     const notesResponse = await request.get('http://localhost:8080/notes');
  256 |     const notesData = await notesResponse.json();
```