# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: type-filters.spec.ts >> Type Filters - Home Page Filtering >> should filter notes by comet type
- Location: tests\type-filters.spec.ts:93:3

# Error details

```
Error: page.waitForLoadState: Target page, context or browser has been closed
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | import { createNote, getBackendUrl } from './helpers/testData';
  3   | 
  4   | /**
  5   |  * Tests for Type Filtering with metadata.type fallback
  6   |  * Verifies that filters work correctly when type is stored in metadata
  7   |  * instead of the root type field
  8   |  */
  9   | 
  10  | test.describe('Type Filters - Home Page Filtering', { tag: ['@smoke', '@filters'] }, () => {
  11  | 
  12  |   test.beforeEach(async ({ page }) => {
  13  |     await page.goto('/');
> 14  |     await page.waitForLoadState('networkidle');
      |                ^ Error: page.waitForLoadState: Target page, context or browser has been closed
  15  |     await page.waitForTimeout(1000);
  16  |   });
  17  | 
  18  |   test('should filter notes by star type', async ({ page, request }) => {
  19  |     const timestamp = Date.now();
  20  | 
  21  |     // Create star type note using helper
  22  |     await createNote(request, {
  23  |       title: `Star Filter Test ${timestamp}`,
  24  |       content: 'Star content',
  25  |       type: 'star'
  26  |     });
  27  | 
  28  |     // Create planet type note
  29  |     await createNote(request, {
  30  |       title: `Planet Filter Test ${timestamp}`,
  31  |       content: 'Planet content',
  32  |       type: 'planet'
  33  |     });
  34  |     
  35  |     // Reload to see new notes
  36  |     await page.reload();
  37  |     await page.waitForTimeout(2000);
  38  |     
  39  |     // Click on Stars filter button
  40  |     const starsFilter = page.locator('button:has-text("⭐"), button:has-text("Stars"), button[data-filter="star"]').first();
  41  |     if (await starsFilter.isVisible().catch(() => false)) {
  42  |       await starsFilter.click();
  43  |       await page.waitForTimeout(1000);
  44  |       
  45  |       // Verify filter is applied - stats should show filtered state
  46  |       const statsBar = page.locator('.stats-bar').first();
  47  |       if (await statsBar.isVisible().catch(() => false)) {
  48  |         const statsText = await statsBar.textContent();
  49  |         // Should show filter indicator or specific count
  50  |         const hasFilterIndicator = statsText?.toLowerCase().includes('filter') || 
  51  |                                    statsText?.toLowerCase().includes('star');
  52  |         expect(hasFilterIndicator).toBe(true);
  53  |       }
  54  |     }
  55  |   });
  56  | 
  57  |   test('should filter notes by planet type', async ({ page, request }) => {
  58  |     const timestamp = Date.now();
  59  | 
  60  |     // Create notes with different types using helper
  61  |     await createNote(request, {
  62  |       title: `Planet Type ${timestamp}`,
  63  |       content: 'Planet content',
  64  |       type: 'planet'
  65  |     });
  66  | 
  67  |     await createNote(request, {
  68  |       title: `Comet Type ${timestamp}`,
  69  |       content: 'Comet content',
  70  |       type: 'comet'
  71  |     });
  72  |     
  73  |     await page.reload();
  74  |     await page.waitForTimeout(2000);
  75  |     
  76  |     // Click on Planets filter
  77  |     const planetsFilter = page.locator('button:has-text("🪐"), button:has-text("Planets"), button[data-filter="planet"]').first();
  78  |     if (await planetsFilter.isVisible().catch(() => false)) {
  79  |       await planetsFilter.click();
  80  |       await page.waitForTimeout(1000);
  81  |       
  82  |       // Verify filter is active
  83  |       const statsBar = page.locator('.stats-bar').first();
  84  |       if (await statsBar.isVisible().catch(() => false)) {
  85  |         const statsText = await statsBar.textContent();
  86  |         const hasFilterIndicator = statsText?.toLowerCase().includes('filter') || 
  87  |                                    statsText?.toLowerCase().includes('planet');
  88  |         expect(hasFilterIndicator).toBe(true);
  89  |       }
  90  |     }
  91  |   });
  92  | 
  93  |   test('should filter notes by comet type', async ({ page, request }) => {
  94  |     const timestamp = Date.now();
  95  | 
  96  |     // Create comet type note using helper
  97  |     await createNote(request, {
  98  |       title: `Comet Filter Test ${timestamp}`,
  99  |       content: 'Comet content',
  100 |       type: 'comet'
  101 |     });
  102 | 
  103 |     // Create galaxy type note
  104 |     await createNote(request, {
  105 |       title: `Galaxy Filter Test ${timestamp}`,
  106 |       content: 'Galaxy content',
  107 |       type: 'galaxy'
  108 |     });
  109 |     
  110 |     // Reload to see new notes
  111 |     await page.reload();
  112 |     await page.waitForTimeout(2000);
  113 |     
  114 |     // Click on Comets filter
```