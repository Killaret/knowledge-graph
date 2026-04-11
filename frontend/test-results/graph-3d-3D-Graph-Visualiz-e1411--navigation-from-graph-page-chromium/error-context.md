# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d.spec.ts >> 3D Graph Visualization >> should handle back button navigation from graph page
- Location: tests\graph-3d.spec.ts:75:3

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: page.waitForURL: Test timeout of 30000ms exceeded.
=========================== logs ===========================
waiting for navigation until "load"
============================================================
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - button "Go back" [ref=e4] [cursor=pointer]: « Back
  - heading "Navigation Test Note" [level=1] [ref=e5]
  - generic [ref=e6]: "Created: 4/11/2026, 7:53:55 PM"
  - generic [ref=e7]: Testing navigation from graph
  - generic [ref=e8]:
    - link "Edit" [ref=e9] [cursor=pointer]:
      - /url: /notes/b746ac88-8f78-44f1-a749-bafcc7c8d419/edit
    - button "Delete" [ref=e10]
    - link "✨ Show constellation" [ref=e11] [cursor=pointer]:
      - /url: /graph/b746ac88-8f78-44f1-a749-bafcc7c8d419
  - heading "Similar notes" [level=2] [ref=e12]
  - list [ref=e13]:
    - listitem [ref=e14]:
      - link "Graph Back Test" [ref=e15] [cursor=pointer]:
        - /url: /notes/8e915eb2-c9eb-486a-b31d-9b95a0b88c15
      - text: "score: 0.690"
    - listitem [ref=e16]:
      - link "Graph Back Test" [ref=e17] [cursor=pointer]:
        - /url: /notes/2949f5db-1677-43be-bc66-122132eace02
      - text: "score: 0.690"
    - listitem [ref=e18]:
      - link "Graph Back Test" [ref=e19] [cursor=pointer]:
        - /url: /notes/f6b40f42-e807-46be-a418-fc8d6a3d1e81
      - text: "score: 0.690"
    - listitem [ref=e20]:
      - link "Graph Back Test" [ref=e21] [cursor=pointer]:
        - /url: /notes/1034a782-24f0-4ded-bdb9-b8583d314d48
      - text: "score: 0.690"
    - listitem [ref=e22]:
      - link "Graph Back Test" [ref=e23] [cursor=pointer]:
        - /url: /notes/60972bbb-bdab-45d3-a1c2-60f41f521293
      - text: "score: 0.690"
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | /**
  4   |  * Tests for 3D Graph Visualization
  5   |  * These tests verify that the 3D graph canvas renders correctly
  6   |  * and that the SmartGraph component properly selects 2D or 3D mode
  7   |  */
  8   | 
  9   | test.describe('3D Graph Visualization', () => {
  10  |   
  11  |   test.beforeEach(async ({ page }) => {
  12  |     // Navigate to home page first to create a note
  13  |     await page.goto('/');
  14  |     await page.waitForLoadState('networkidle');
  15  |   });
  16  | 
  17  |   test('should render graph page with canvas visible', async ({ page }) => {
  18  |     // Create a new note first
  19  |     await page.click('a:has-text("New Note")');
  20  |     await page.waitForURL(/\/notes\/new/);
  21  |     
  22  |     // Fill in note details
  23  |     await page.fill('input[name="title"]', 'Graph Test Note');
  24  |     await page.fill('textarea[name="content"]', 'This is a test note for graph visualization');
  25  |     await page.click('button[type="submit"]');
  26  |     
  27  |     // Wait for note creation and redirect to note page
  28  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  29  |     
  30  |     // Get the note ID from URL
  31  |     const noteUrl = page.url();
  32  |     const noteId = noteUrl.split('/').pop();
  33  |     
  34  |     // Navigate to graph page
  35  |     await page.goto(`/graph/${noteId}`);
  36  |     await page.waitForLoadState('networkidle');
  37  |     
  38  |     // Verify page title
  39  |     await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
  40  |     
  41  |     // Verify canvas or 2D fallback is visible
  42  |     const canvas = page.locator('canvas').first();
  43  |     const graph2D = page.locator('.graph-2d, .graph-3d, .graph-container canvas').first();
  44  |     
  45  |     await expect(canvas.or(graph2D)).toBeVisible({ timeout: 10000 });
  46  |   });
  47  | 
  48  |   test('should show graph container with correct styling', async ({ page }) => {
  49  |     // Create a new note
  50  |     await page.click('a:has-text("New Note")');
  51  |     await page.waitForURL(/\/notes\/new/);
  52  |     
  53  |     await page.fill('input[name="title"]', 'Styling Test Note');
  54  |     await page.fill('textarea[name="content"]', 'Testing graph styling');
  55  |     await page.click('button[type="submit"]');
  56  |     
  57  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  58  |     const noteUrl = page.url();
  59  |     const noteId = noteUrl.split('/').pop();
  60  |     
  61  |     // Navigate to graph page
  62  |     await page.goto(`/graph/${noteId}`);
  63  |     await page.waitForLoadState('networkidle');
  64  |     
  65  |     // Verify graph container has correct background
  66  |     const graphPage = page.locator('.graph-page');
  67  |     await expect(graphPage).toBeVisible();
  68  |     
  69  |     // Verify header elements
  70  |     await expect(page.locator('.graph-header')).toBeVisible();
  71  |     await expect(page.locator('.graph-header h1')).toHaveText('Knowledge Constellation');
  72  |     await expect(page.locator('.hint')).toBeVisible();
  73  |   });
  74  | 
  75  |   test('should handle back button navigation from graph page', async ({ page }) => {
  76  |     // Create a new note
  77  |     await page.click('a:has-text("New Note")');
  78  |     await page.waitForURL(/\/notes\/new/);
  79  |     
  80  |     await page.fill('input[name="title"]', 'Navigation Test Note');
  81  |     await page.fill('textarea[name="content"]', 'Testing navigation from graph');
  82  |     await page.click('button[type="submit"]');
  83  |     
  84  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  85  |     const noteUrl = page.url();
  86  |     const noteId = noteUrl.split('/').pop();
  87  |     
  88  |     // Navigate to graph page
  89  |     await page.goto(`/graph/${noteId}`);
  90  |     await page.waitForLoadState('networkidle');
  91  |     
  92  |     // Click back button
  93  |     await page.click('.back-button');
  94  |     
  95  |     // Should navigate back
> 96  |     await page.waitForURL(/\/$/);
      |                ^ Error: page.waitForURL: Test timeout of 30000ms exceeded.
  97  |     await expect(page).toHaveURL('/');
  98  |   });
  99  | 
  100 |   test('should display performance mode indicator', async ({ page }) => {
  101 |     // Create a new note
  102 |     await page.click('a:has-text("New Note")');
  103 |     await page.waitForURL(/\/notes\/new/);
  104 |     
  105 |     await page.fill('input[name="title"]', 'Performance Test Note');
  106 |     await page.fill('textarea[name="content"]', 'Testing performance mode indicator');
  107 |     await page.click('button[type="submit"]');
  108 |     
  109 |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  110 |     const noteUrl = page.url();
  111 |     const noteId = noteUrl.split('/').pop();
  112 |     
  113 |     // Navigate to graph page
  114 |     await page.goto(`/graph/${noteId}`);
  115 |     await page.waitForLoadState('networkidle');
  116 |     
  117 |     // Wait for graph to load (either 2D or 3D)
  118 |     await page.waitForTimeout(2000);
  119 |     
  120 |     // Check for performance hint (either "3D Mode" or "2D Mode")
  121 |     const performanceHint = page.locator('.performance-hint');
  122 |     if (await performanceHint.isVisible().catch(() => false)) {
  123 |       const hintText = await performanceHint.textContent();
  124 |       expect(hintText).toMatch(/(3D Mode|2D Mode)/);
  125 |     }
  126 |   });
  127 | 
  128 |   test('should handle graph page with no nodes gracefully', async ({ page }) => {
  129 |     // Create a note with no links
  130 |     await page.click('a:has-text("New Note")');
  131 |     await page.waitForURL(/\/notes\/new/);
  132 |     
  133 |     await page.fill('input[name="title"]', 'Isolated Note');
  134 |     await page.fill('textarea[name="content"]', 'This note has no connections');
  135 |     await page.click('button[type="submit"]');
  136 |     
  137 |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  138 |     const noteUrl = page.url();
  139 |     const noteId = noteUrl.split('/').pop();
  140 |     
  141 |     // Navigate to graph page
  142 |     await page.goto(`/graph/${noteId}`);
  143 |     await page.waitForLoadState('networkidle');
  144 |     
  145 |     // Page should still render without errors
  146 |     await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
  147 |     await expect(page.locator('.graph-container')).toBeVisible();
  148 |   });
  149 | });
  150 | 
```