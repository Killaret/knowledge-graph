# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d.spec.ts >> Graph Visualization - Progressive Rendering >> should render 3D graph immediately without spinner
- Location: tests\graph-3d.spec.ts:17:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('[data-testid="graph-stats"]').first()
Expected: visible
Timeout: 2000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 2000ms
  - waiting for locator('[data-testid="graph-stats"]').first()

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e5] [cursor=pointer]:
    - checkbox "Показать все заметки (выключено)" [ref=e6]
    - generic [ref=e7]: Показать все заметки (выключено)
  - generic [ref=e8]:
    - generic [ref=e9]:
      - strong [ref=e10]: "1"
      - text: nodes
    - generic [ref=e11]:
      - strong [ref=e12]: "0"
      - text: links
    - generic [ref=e13]: (Local view)
  - generic [ref=e15]:
    - button "Reset View" [ref=e17] [cursor=pointer]:
      - img [ref=e18]
      - generic [ref=e20]: Reset View
    - generic [ref=e22] [cursor=pointer]: 3D Graph Test Note
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | import { createNote, createLink, getBackendUrl } from './helpers/testData';
  3   | 
  4   | /**
  5   |  * Tests for Graph Visualization with Progressive Rendering
  6   |  * These tests verify the new architecture with immediate loading and fog effect
  7   |  */
  8   | 
  9   | test.describe('Graph Visualization - Progressive Rendering', { tag: ['@smoke', '@3d', '@progressive'] }, () => {
  10  |   
  11  |   test.beforeEach(async ({ page }) => {
  12  |     // Navigate to home page first
  13  |     await page.goto('/');
  14  |     await page.waitForLoadState('networkidle');
  15  |   });
  16  | 
  17  |   test('should render 3D graph immediately without spinner', async ({ page, request }) => {
  18  |     // Create a note via API using helper
  19  |     const note = await createNote(request, {
  20  |       title: '3D Graph Test Note',
  21  |       content: 'Test note for 3D graph'
  22  |     });
  23  |     const noteId = note.id;
  24  | 
  25  |     // Navigate to 3D graph page directly
  26  |     await page.goto(`/graph/3d/${noteId}`);
  27  |     await page.waitForLoadState('networkidle');
  28  |     
  29  |     // Graph should appear immediately (no lazy loading spinner)
  30  |     const graphContainer = page.locator('[data-testid="graph-3d-container"]').first();
  31  |     await expect(graphContainer).toBeVisible({ timeout: 3000 });
  32  |     
  33  |     // Loading overlay may be present briefly but should disappear
  34  |     const loadingOverlay = page.locator('[data-testid="loading-overlay"]');
  35  |     await expect(loadingOverlay).toBeHidden({ timeout: 8000 });
  36  |     
  37  |     // Stats bar should show immediately with node count
  38  |     const statsBar = page.locator('[data-testid="graph-stats"]').first();
> 39  |     await expect(statsBar).toBeVisible({ timeout: 2000 });
      |                            ^ Error: expect(locator).toBeVisible() failed
  40  |     
  41  |     const statsText = await statsBar.textContent();
  42  |     expect(statsText).toMatch(/\d+\s*nodes?/i);
  43  |   });
  44  | 
  45  |   test('should show graph container with correct 3D styling', async ({ page, request }) => {
  46  |     // Create a note via API using helper
  47  |     const note = await createNote(request, {
  48  |       title: 'Styling Test Note',
  49  |       content: 'Testing 3D graph styling'
  50  |     });
  51  |     const noteId = note.id;
  52  | 
  53  |     // Navigate to 3D graph page
  54  |     await page.goto(`/graph/3d/${noteId}`);
  55  |     await page.waitForLoadState('networkidle');
  56  |     await page.waitForTimeout(2000);
  57  |     
  58  |     // Verify 3D graph container is visible
  59  |     const graphContainer = page.locator('[data-testid="graph-3d-container"]').first();
  60  |     await expect(graphContainer).toBeVisible();
  61  |     
  62  |     // Verify container has correct CSS
  63  |     const containerStyles = await graphContainer.evaluate(el => {
  64  |       const styles = window.getComputedStyle(el);
  65  |       return {
  66  |         position: styles.position,
  67  |         width: styles.width,
  68  |         height: styles.height,
  69  |         overflow: styles.overflow,
  70  |         backgroundColor: styles.backgroundColor
  71  |       };
  72  |     });
  73  |     
  74  |     expect(containerStyles.position).toBe('relative');
  75  |     expect(containerStyles.overflow).toBe('hidden');
  76  |     expect(containerStyles.width).not.toBe('0px');
  77  |     expect(containerStyles.height).not.toBe('0px');
  78  |   });
  79  | 
  80  |   test('should handle back button navigation from 3D graph page', async ({ page, request }) => {
  81  |     // Create a note via API using helper
  82  |     const note = await createNote(request, {
  83  |       title: 'Navigation Test Note',
  84  |       content: 'Testing navigation from 3D graph'
  85  |     });
  86  |     const noteId = note.id;
  87  | 
  88  |     // Navigate to 3D graph page
  89  |     await page.goto(`/graph/3d/${noteId}`);
  90  |     await page.waitForLoadState('networkidle');
  91  |     await page.waitForTimeout(1000);
  92  |     
  93  |     // Navigate back using browser back
  94  |     await page.goBack();
  95  |     await page.waitForTimeout(1000);
  96  |     
  97  |     // Should be back on home or note page
  98  |     const currentUrl = page.url();
  99  |     const baseUrl = process.env.FRONTEND_URL || 'http://localhost:5173';
  100 |     expect(currentUrl).toMatch(new RegExp(`${baseUrl.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}(\/|\/notes\/.+)`));
  101 |   });
  102 | 
  103 |   test('should display stats bar with node and link counts', async ({ page, request }) => {
  104 |     // Create a note with connections using helper
  105 |     const note1 = await createNote(request, { title: 'Stats Test Node 1', content: 'Node 1' });
  106 |     const note1Id = note1.id;
  107 | 
  108 |     const note2 = await createNote(request, { title: 'Stats Test Node 2', content: 'Node 2' });
  109 |     const note2Id = note2.id;
  110 | 
  111 |     // Create link between notes
  112 |     await createLink(request, note1Id, note2Id, 0.8);
  113 | 
  114 |     // Navigate to 3D graph
  115 |     await page.goto(`/graph/3d/${note1Id}`);
  116 |     await page.waitForLoadState('networkidle');
  117 |     await page.waitForTimeout(2000);
  118 |     
  119 |     // Verify stats bar shows correct counts
  120 |     const statsBar = page.locator('.stats-bar').first();
  121 |     await expect(statsBar).toBeVisible();
  122 |     
  123 |     const statsText = await statsBar.textContent();
  124 |     expect(statsText).toContain('nodes');
  125 |     expect(statsText).toContain('links');
  126 |     
  127 |     // Should show "Local view" mode
  128 |     expect(statsText).toContain('Local view');
  129 |   });
  130 | 
  131 | });
  132 | 
```