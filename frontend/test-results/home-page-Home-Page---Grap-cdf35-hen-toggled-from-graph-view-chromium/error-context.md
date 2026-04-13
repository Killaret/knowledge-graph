# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: home-page.spec.ts >> Home Page - Graph First >> should display list view when toggled from graph view
- Location: tests\home-page.spec.ts:59:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e4]:
    - generic [ref=e5]:
      - button "2D" [ref=e6] [cursor=pointer]:
        - img [ref=e7]
        - generic [ref=e17]: 2D
      - button "3D" [active] [ref=e18] [cursor=pointer]:
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
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | /**
  4   |  * Tests for Home Page - Graph-first interface
  5   |  * Verifies that the main page displays the graph canvas by default
  6   |  * and list view is accessible from it
  7   |  */
  8   | 
  9   | test.describe('Home Page - Graph First', () => {
  10  |   
  11  |   test.beforeEach(async ({ page }) => {
  12  |     // Navigate to home page
  13  |     await page.goto('http://localhost:5173/');
  14  |     await page.waitForLoadState('networkidle');
  15  |     await page.waitForTimeout(1000);
  16  |   });
  17  | 
  18  |   test('should display graph canvas by default on home page', async ({ page }) => {
  19  |     // Verify graph container is visible
  20  |     const graphContainer = page.locator('.graph-container, .graph-canvas, canvas').first();
  21  |     await expect(graphContainer).toBeVisible({ timeout: 10000 });
  22  |     
  23  |     // Verify graph has height (is rendered)
  24  |     const containerHeight = await graphContainer.evaluate(el => (el as HTMLElement).style.height);
  25  |     expect(containerHeight).toBeTruthy();
  26  |   });
  27  | 
  28  |   test('should load notes and display them on graph', async ({ page, request }) => {
  29  |     // Create a test note via API
  30  |     const timestamp = Date.now();
  31  |     const note = await request.post('http://localhost:8080/notes', {
  32  |       data: { 
  33  |         title: `Home Page Test Note ${timestamp}`, 
  34  |         content: 'Test content for home page',
  35  |         type: 'star'
  36  |       }
  37  |     });
  38  |     expect(note.status()).toBe(201);
  39  |     
  40  |     // Reload page to see the note
  41  |     await page.reload();
  42  |     await page.waitForTimeout(2000);
  43  |     
  44  |     // Verify graph shows data (not empty state)
  45  |     const emptyState = page.locator('text=No graph data, text=Create some notes').first();
  46  |     const isEmptyVisible = await emptyState.isVisible().catch(() => false);
  47  |     
  48  |     // Either graph has nodes or we see note cards
  49  |     if (isEmptyVisible) {
  50  |       // If empty state is visible, that's also valid - means no notes with links yet
  51  |       await expect(emptyState).toBeVisible();
  52  |     } else {
  53  |       // Otherwise graph should be rendered
  54  |       const graphCanvas = page.locator('.graph-container canvas, .graph-canvas').first();
  55  |       await expect(graphCanvas).toBeVisible();
  56  |     }
  57  |   });
  58  | 
  59  |   test('should display list view when toggled from graph view', async ({ page }) => {
  60  |     // Create a note first
  61  |     await page.click('.create-btn, button:has-text("+")');
  62  |     await page.fill('input[name="title"]', 'List View Test Note');
  63  |     await page.fill('textarea[name="content"]', 'Content for list view test');
  64  |     await page.click('button[type="submit"]');
  65  |     await page.waitForTimeout(1500);
  66  |     
  67  |     // Reload to ensure we're on fresh home page
  68  |     await page.goto('http://localhost:5173/');
  69  |     await page.waitForTimeout(1000);
  70  |     
  71  |     // Toggle to list view (using FloatingControls or view toggle)
  72  |     // Try to find and click the view toggle button
  73  |     const viewToggle = page.locator('button:has-text("List"), button:has-text("View"), .view-toggle').first();
  74  |     if (await viewToggle.isVisible().catch(() => false)) {
  75  |       await viewToggle.click();
  76  |     }
  77  |     
  78  |     // Verify list view elements
  79  |     const noteCards = page.locator('.note-card');
  80  |     const notesGrid = page.locator('.notes-grid').first();
  81  |     
  82  |     // Either notes grid is visible or we see the graph
  83  |     const hasListView = await notesGrid.isVisible().catch(() => false);
  84  |     const hasNoteCards = await noteCards.first().isVisible().catch(() => false);
  85  |     
  86  |     // At least one of these should be visible
> 87  |     expect(hasListView || hasNoteCards).toBe(true);
      |                                         ^ Error: expect(received).toBe(expected) // Object.is equality
  88  |   });
  89  | 
  90  |   test('should show note count in stats bar', async ({ page, request }) => {
  91  |     // Create test notes
  92  |     const timestamp = Date.now();
  93  |     await request.post('http://localhost:8080/notes', {
  94  |       data: { title: `Stats Test 1 ${timestamp}`, content: 'Content 1', type: 'star' }
  95  |     });
  96  |     await request.post('http://localhost:8080/notes', {
  97  |       data: { title: `Stats Test 2 ${timestamp}`, content: 'Content 2', type: 'planet' }
  98  |     });
  99  |     
  100 |     // Reload page
  101 |     await page.reload();
  102 |     await page.waitForTimeout(2000);
  103 |     
  104 |     // Verify stats bar shows note count
  105 |     const statsBar = page.locator('.stats-bar, .stats-total').first();
  106 |     await expect(statsBar).toBeVisible();
  107 |     
  108 |     // Check that count is greater than 0
  109 |     const statsText = await statsBar.textContent();
  110 |     const countMatch = statsText?.match(/(\d+)\s+note/);
  111 |     if (countMatch) {
  112 |       const count = parseInt(countMatch[1], 10);
  113 |       expect(count).toBeGreaterThan(0);
  114 |     }
  115 |   });
  116 | 
  117 |   test('should filter notes by type from home page', async ({ page, request }) => {
  118 |     // Create notes of different types
  119 |     const timestamp = Date.now();
  120 |     await request.post('http://localhost:8080/notes', {
  121 |       data: { title: `Star Note ${timestamp}`, content: 'Star content', type: 'star' }
  122 |     });
  123 |     await request.post('http://localhost:8080/notes', {
  124 |       data: { title: `Planet Note ${timestamp}`, content: 'Planet content', type: 'planet' }
  125 |     });
  126 |     
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
```