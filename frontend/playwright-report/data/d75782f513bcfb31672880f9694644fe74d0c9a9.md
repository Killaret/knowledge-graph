# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notes.spec.ts >> Knowledge Graph Frontend >> should delete a note
- Location: tests\notes.spec.ts:68:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/
Call log:
  - navigating to "http://localhost:5173/", waiting until "load"

```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | test.describe('Knowledge Graph Frontend', () => {
  4   |   test.beforeEach(async ({ page }) => {
> 5   |     await page.goto('http://localhost:5173');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/
  6   |     await page.waitForLoadState('networkidle');
  7   |   });
  8   | 
  9   |   test('should create a new note', async ({ page, request }) => {
  10  |     // Wait for floating controls to be visible
  11  |     await expect(page.locator('.floating-controls')).toBeVisible({ timeout: 10000 });
  12  |     
  13  |     // Click create button in floating controls
  14  |     await expect(page.locator('.create-btn')).toBeVisible();
  15  |     await page.click('.create-btn');
  16  |     
  17  |     // Wait for modal to open
  18  |     await page.waitForSelector('.modal, [role="dialog"]', { timeout: 10000 });
  19  |     await page.waitForTimeout(500);
  20  |     
  21  |     // Fill in create modal
  22  |     await page.waitForSelector('input[name="title"]', { timeout: 5000 });
  23  |     await page.fill('input[name="title"]', 'Playwright Test ' + Date.now());
  24  |     await page.fill('textarea[name="content"]', 'Automated content');
  25  |     
  26  |     // Click Save button
  27  |     await page.click('button[type="submit"]');
  28  |     
  29  |     // Wait for modal to close
  30  |     await page.waitForTimeout(2000);
  31  |     
  32  |     // Verify via API that note was created
  33  |     const notesResponse = await request.get('http://localhost:8080/notes');
  34  |     const notesData = await notesResponse.json();
  35  |     expect(notesData.total).toBeGreaterThan(0);
  36  |     
  37  |     // Note: Due to API serialization issue, we verify creation via API only
  38  |     // The UI list may not refresh correctly until backend is fixed
  39  |   });
  40  | 
  41  |   test('should edit a note', async ({ page, request }) => {
  42  |     // Create a note via API first
  43  |     const timestamp = Date.now();
  44  |     const note = await request.post('http://localhost:8080/notes', {
  45  |       data: { title: 'Edit Test ' + timestamp, content: 'Original content', type: 'star' }
  46  |     });
  47  |     const noteId = (await note.json()).id;
  48  |     
  49  |     // Navigate to edit page directly
  50  |     await page.goto(`http://localhost:5173/notes/${noteId}/edit`);
  51  |     await page.waitForTimeout(1000);
  52  |     
  53  |     // Update note
  54  |     await page.waitForSelector('input[name="title"]', { timeout: 5000 });
  55  |     await page.fill('input[name="title"]', 'Edited ' + timestamp);
  56  |     await page.fill('textarea[name="content"]', 'Updated content');
  57  |     await page.click('button[type="submit"]');
  58  |     
  59  |     // Wait for redirect to note page
  60  |     await page.waitForURL(`http://localhost:5173/notes/${noteId}`, { timeout: 5000 });
  61  |     
  62  |     // Verify via API
  63  |     const updatedNote = await request.get(`http://localhost:8080/notes/${noteId}`);
  64  |     const noteData = await updatedNote.json();
  65  |     expect(noteData.title).toBe('Edited ' + timestamp);
  66  |   });
  67  | 
  68  |   test('should delete a note', async ({ page, request }) => {
  69  |     // Create a note via API first
  70  |     const timestamp = Date.now();
  71  |     const note = await request.post('http://localhost:8080/notes', {
  72  |       data: { title: 'Delete Test ' + timestamp, content: 'Test content for deletion' }
  73  |     });
  74  |     const noteId = (await note.json()).id;
  75  |     
  76  |     // Navigate directly to note page
  77  |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  78  |     await page.waitForTimeout(1000);
  79  |     
  80  |     // Setup dialog handler before click
  81  |     page.on('dialog', async dialog => {
  82  |       await dialog.accept();
  83  |     });
  84  |     
  85  |     // Click Delete button
  86  |     await page.click('button:has-text("Delete")');
  87  |     
  88  |     // Wait for redirect to home page (give more time)
  89  |     await page.waitForURL('http://localhost:5173/', { timeout: 10000 });
  90  |     await page.waitForTimeout(1000);
  91  |     
  92  |     // Verify via API that note is deleted
  93  |     const checkResponse = await request.get(`http://localhost:8080/notes/${noteId}`);
  94  |     expect(checkResponse.status()).toBe(404);
  95  |   });
  96  | 
  97  |   test('should open graph for a note with links', async ({ page, request }) => {
  98  |     // Create two notes and a link via API
  99  |     const note1 = await request.post('http://localhost:8080/notes', {
  100 |       data: { title: 'Node A', content: 'A' }
  101 |     });
  102 |     const note2 = await request.post('http://localhost:8080/notes', {
  103 |       data: { title: 'Node B', content: 'B' }
  104 |     });
  105 |     const id1 = (await note1.json()).id;
```