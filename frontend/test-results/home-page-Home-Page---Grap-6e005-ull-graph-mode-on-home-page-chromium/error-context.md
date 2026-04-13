# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: home-page.spec.ts >> Home Page - Graph First >> should toggle full graph mode on home page
- Location: tests\home-page.spec.ts:275:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: false
Received: true
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
      - textbox "Search notes..." [ref=e28]
      - button "Search" [ref=e29] [cursor=pointer]:
        - img [ref=e30]
    - button "Menu" [ref=e34] [cursor=pointer]:
      - img [ref=e35]
    - button "Create new note" [ref=e36] [cursor=pointer]:
      - img [ref=e37]
  - generic [ref=e38]:
    - generic [ref=e39]:
      - generic [ref=e40]:
        - generic [ref=e41]: "Filter:"
        - generic [ref=e42]:
          - button "🌌 All" [ref=e43] [cursor=pointer]:
            - generic [ref=e44]: 🌌
            - generic [ref=e45]: All
          - button "⭐ Stars" [ref=e46] [cursor=pointer]:
            - generic [ref=e47]: ⭐
            - generic [ref=e48]: Stars
          - button "🪐 Planets" [ref=e49] [cursor=pointer]:
            - generic [ref=e50]: 🪐
            - generic [ref=e51]: Planets
          - button "☄️ Comets" [ref=e52] [cursor=pointer]:
            - generic [ref=e53]: ☄️
            - generic [ref=e54]: Comets
          - button "🌀 Galaxies" [ref=e55] [cursor=pointer]:
            - generic [ref=e56]: 🌀
            - generic [ref=e57]: Galaxies
      - generic [ref=e58]:
        - generic [ref=e59]: "Sort:"
        - combobox "Sort:" [ref=e60] [cursor=pointer]:
          - option "Newest first" [selected]
          - option "Oldest first"
          - option "Alphabetical (A-Z)"
          - option "Alphabetical (Z-A)"
    - generic [ref=e62]: 20 notes
    - generic [ref=e64] [cursor=pointer]:
      - checkbox "Показать все заметки (включено)" [checked] [active] [ref=e65]
      - generic [ref=e66]: Показать все заметки (включено)
```

# Test source

```ts
  210 | 
  211 |   test('should navigate to graph view for specific note', async ({ page, request }) => {
  212 |     // Create a note
  213 |     const timestamp = Date.now();
  214 |     const note = await request.post('http://localhost:8080/notes', {
  215 |       data: { 
  216 |         title: `Graph View Test ${timestamp}`, 
  217 |         content: 'Test content',
  218 |         type: 'star'
  219 |       }
  220 |     });
  221 |     const noteId = (await note.json()).id;
  222 |     
  223 |     // Navigate to specific graph page
  224 |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  225 |     await page.waitForTimeout(2000);
  226 |     
  227 |     // Verify graph container is visible
  228 |     const graphContainer = page.locator('.graph-3d-container, .graph-container, canvas').first();
  229 |     await expect(graphContainer).toBeVisible({ timeout: 10000 });
  230 |   });
  231 | 
  232 |   test('should display general graph view at /graph', async ({ page }) => {
  233 |     // Navigate to general graph page
  234 |     await page.goto('http://localhost:5173/graph');
  235 |     await page.waitForLoadState('networkidle');
  236 |     await page.waitForTimeout(3000);
  237 |     
  238 |     // Verify no 404 error
  239 |     const error404 = page.locator('text=404, text=Not Found').first();
  240 |     const has404 = await error404.isVisible().catch(() => false);
  241 |     expect(has404).toBe(false);
  242 |     
  243 |     // Verify graph container, empty state, or error state is visible
  244 |     const graphContainer = page.locator('.graph-container, .graph-3d-container, canvas').first();
  245 |     const emptyState = page.locator('text=No notes found, text=No graph data').first();
  246 |     const errorState = page.locator('text=Failed to load graph data').first();
  247 |     
  248 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  249 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  250 |     const hasError = await errorState.isVisible().catch(() => false);
  251 |     
  252 |     expect(hasGraph || hasEmpty || hasError).toBe(true);
  253 |   });
  254 | 
  255 |   test('should handle empty state when no notes exist', async ({ page, request }) => {
  256 |     // Clear all notes via API (if possible) or just check current state
  257 |     // This test verifies the empty state message
  258 |     
  259 |     const notesResponse = await request.get('http://localhost:8080/notes');
  260 |     const notesData = await notesResponse.json();
  261 |     
  262 |     if (notesData.total === 0 || notesData.notes?.length === 0) {
  263 |       // If truly empty, verify empty state
  264 |       await page.reload();
  265 |       await page.waitForTimeout(1000);
  266 |       
  267 |       const emptyState = page.locator('.empty-state, text=No notes').first();
  268 |       await expect(emptyState).toBeVisible();
  269 |     } else {
  270 |       // Skip this test if notes exist
  271 |       test.skip();
  272 |     }
  273 |   });
  274 | 
  275 |   test('should toggle full graph mode on home page', async ({ page, request }) => {
  276 |     // Create test notes if needed
  277 |     const notesResponse = await request.get('http://localhost:8080/notes');
  278 |     const notesData = await notesResponse.json();
  279 |     
  280 |     if (notesData.total < 2) {
  281 |       // Create at least 2 notes for meaningful test
  282 |       await request.post('http://localhost:8080/notes', {
  283 |         data: { title: 'Note 1', content: 'Content 1', type: 'star' }
  284 |       });
  285 |       await request.post('http://localhost:8080/notes', {
  286 |         data: { title: 'Note 2', content: 'Content 2', type: 'planet' }
  287 |       });
  288 |       await page.reload();
  289 |       await page.waitForTimeout(2000);
  290 |     }
  291 |     
  292 |     // Find and click the full graph toggle
  293 |     const toggle = page.locator('.graph-mode-toggle input[type="checkbox"]').first();
  294 |     const hasToggle = await toggle.isVisible().catch(() => false);
  295 |     
  296 |     if (!hasToggle) {
  297 |       test.skip(true, 'Full graph toggle not found');
  298 |       return;
  299 |     }
  300 |     
  301 |     // Get initial state
  302 |     const isInitiallyChecked = await toggle.isChecked();
  303 |     
  304 |     // Click to toggle
  305 |     await toggle.click();
  306 |     await page.waitForTimeout(2000);
  307 |     
  308 |     // Verify toggle changed state
  309 |     const isNowChecked = await toggle.isChecked();
> 310 |     expect(isNowChecked).toBe(!isInitiallyChecked);
      |                          ^ Error: expect(received).toBe(expected) // Object.is equality
  311 |     
  312 |     // Verify graph still renders
  313 |     const container = page.locator('.graph-container, .lazy-error, .error-overlay').first();
  314 |     await expect(container).toBeVisible();
  315 |   });
  316 | 
  317 |   test('should display correct note count in stats', async ({ page, request }) => {
  318 |     // Get actual note count from API
  319 |     const notesResponse = await request.get('http://localhost:8080/notes');
  320 |     const notesData = await notesResponse.json();
  321 |     const totalNotes = notesData.total || notesData.notes?.length || 0;
  322 |     
  323 |     // Reload page to ensure fresh data
  324 |     await page.reload();
  325 |     await page.waitForLoadState('networkidle');
  326 |     await page.waitForTimeout(2000);
  327 |     
  328 |     // Check stats bar shows correct count
  329 |     const statsBar = page.locator('.stats-bar').first();
  330 |     const hasStats = await statsBar.isVisible().catch(() => false);
  331 |     
  332 |     if (hasStats) {
  333 |       const statsText = await statsBar.textContent();
  334 |       // Stats should show at least one number
  335 |       expect(statsText).toMatch(/\d+/);
  336 |     }
  337 |     
  338 |     // Toggle to list view and check count
  339 |     const viewToggle = page.locator('button:has-text("List")').first();
  340 |     if (await viewToggle.isVisible().catch(() => false)) {
  341 |       await viewToggle.click();
  342 |       await page.waitForTimeout(1000);
  343 |       
  344 |       // Verify notes grid or count matches
  345 |       const notesGrid = page.locator('.notes-grid').first();
  346 |       const noteCards = page.locator('.note-card');
  347 |       
  348 |       const hasGrid = await notesGrid.isVisible().catch(() => false);
  349 |       const cardCount = hasGrid ? await noteCards.count() : 0;
  350 |       
  351 |       // Card count should be reasonable (not more than total)
  352 |       expect(cardCount).toBeLessThanOrEqual(totalNotes + 5); // +5 tolerance for newly created notes
  353 |     }
  354 |   });
  355 | });
  356 | 
```