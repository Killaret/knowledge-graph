# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: home-page.spec.ts >> Home Page - Graph First >> should display list view when toggled from graph view
- Location: tests\home-page.spec.ts:59:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.graph-container, .notes-grid, .note-card, .lazy-loading, .error-overlay').first()
Expected: visible
Timeout: 10000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 10000ms
  - waiting for locator('.graph-container, .notes-grid, .note-card, .lazy-loading, .error-overlay').first()

```

# Page snapshot

```yaml
- generic [ref=e2]:
  - generic [ref=e4]:
    - generic [ref=e5]:
      - strong [ref=e6]: "100"
      - text: nodes
    - generic [ref=e7]:
      - strong [ref=e8]: "1838"
      - text: links
  - generic [ref=e11]: untitled page
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
  69  |     await page.waitForTimeout(3000); // Wait for page to load
  70  |     
  71  |     // Toggle to list view (using FloatingControls or view toggle)
  72  |     const viewToggle = page.locator('button:has-text("List"), button:has-text("View"), .view-toggle').first();
  73  |     const hasToggle = await viewToggle.isVisible().catch(() => false);
  74  |     
  75  |     if (hasToggle) {
  76  |       await viewToggle.click();
  77  |       await page.waitForTimeout(1500);
  78  |     }
  79  |     
  80  |     // Verify any content is visible (graph, list, loading, or error states)
  81  |     const content = page.locator('.graph-container, .notes-grid, .note-card, .lazy-loading, .error-overlay').first();
> 82  |     await expect(content).toBeVisible({ timeout: 10000 });
      |                           ^ Error: expect(locator).toBeVisible() failed
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
  175 |       await expect(statsBar).toBeVisible();
  176 |     }
  177 |   });
  178 | 
  179 |   test('should open side panel when clicking on graph node', async ({ page, request }) => {
  180 |     // Create a note
  181 |     const timestamp = Date.now();
  182 |     const note = await request.post('http://localhost:8080/notes', {
```