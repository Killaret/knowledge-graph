import { test, expect } from '@playwright/test';

test.describe('Knowledge Graph Frontend', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173');
    await page.waitForLoadState('networkidle');
  });

  test('should create a new note', async ({ page }) => {
    await page.click('[data-testid="fab-new-note"]');
    await page.fill('input[placeholder="Title"]', 'Playwright Test');
    await page.fill('textarea', 'Automated content');
    await page.click('button:has-text("Create")');
    // Wait for navigation to complete with explicit timeout
    await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
    await expect(page.locator('h1')).toHaveText('Playwright Test');
  });

  test('should edit a note', async ({ page }) => {
    // Сначала создадим заметку через API или UI
    await page.click('[data-testid="fab-new-note"]');
    await page.fill('input[placeholder="Title"]', 'To Edit');
    await page.fill('textarea', 'Original');
    await page.click('button:has-text("Create")');
    await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
    // Wait additional time for page to fully load
    await page.waitForTimeout(1000);

    await page.click('a:has-text("Edit")');
    await page.fill('input[placeholder="Title"]', 'Edited');
    await page.fill('textarea', 'New content');
    await page.click('button:has-text("Update")');
    await expect(page.locator('h1')).toHaveText('Edited');
    await expect(page.locator('.content')).toHaveText('New content');
  });

  test('should delete a note', async ({ page, request }) => {
    // Create a note via API first
    const timestamp = Date.now();
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Delete Test ' + timestamp, content: 'Test content for deletion' }
    });
    const noteId = (await note.json()).id;
    const noteTitle = 'Delete Test ' + timestamp;
    
    // Go to home page to see the note in the list
    await page.goto('http://localhost:5173/');
    await page.waitForSelector('.note-card', { timeout: 5000 });
    await expect(page.locator('text=' + noteTitle)).toBeVisible();
    
    // Delete the note via API
    await request.delete(`http://localhost:8080/notes/${noteId}`);
    
    // Wait and reload to see the changes
    await page.waitForTimeout(1000);
    await page.goto('http://localhost:5173/');
    await page.waitForLoadState('networkidle');
    
    // Check that the specific note is no longer present
    await expect(page.locator('text=' + noteTitle)).not.toBeVisible();
  });

  test('should open graph for a note with links', async ({ page, request }) => {
    // Сначала создадим две заметки и связь через API
    const note1 = await request.post('http://localhost:8080/notes', {
      data: { title: 'Node A', content: 'A' }
    });
    const note2 = await request.post('http://localhost:8080/notes', {
      data: { title: 'Node B', content: 'B' }
    });
    const id1 = (await note1.json()).id;
    const id2 = (await note2.json()).id;
    await request.post('http://localhost:8080/links', {
      data: { source_note_id: id1, target_note_id: id2, link_type: 'reference', weight: 1.0 }
    });

    await page.goto(`http://localhost:5173/graph/${id1}`);
    await expect(page.locator('canvas')).toBeVisible();
    // Ждём, пока d3-force немного стабилизируется
    await page.waitForTimeout(1000);
    // Проверяем, что canvas не пустой (можно по цвету пикселя, но сложно)
    const canvas = page.locator('canvas');
    await expect(canvas).toBeVisible();
  });

  test('should show back button on note detail page', async ({ page, request }) => {
    // Create a note via API first
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Back Button Test', content: 'Testing back button functionality' }
    });
    const noteId = (await note.json()).id;
    
    // Navigate to note detail page
    await page.goto(`http://localhost:5173/notes/${noteId}`);
    await page.waitForTimeout(1000);

    // Check that back button is visible
    await expect(page.locator('button:has-text("Back")')).toBeVisible();
    
    // Test back button functionality - should go back to home page
    await page.click('button:has-text("Back")');
    await expect(page).toHaveURL('http://localhost:5173/');
  });

  test('should show back button on graph page', async ({ page, request }) => {
    // Create a note via API first
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Graph Back Test', content: 'Testing back button on graph' }
    });
    const noteId = (await note.json()).id;

    // First navigate to home page to create history
    await page.goto('http://localhost:5173/');
    await page.waitForTimeout(1000);
    
    // Then navigate to note page
    await page.goto(`http://localhost:5173/notes/${noteId}`);
    await page.waitForTimeout(1000);
    
    // Then navigate to graph page
    await page.goto(`http://localhost:5173/graph/${noteId}`);
    await page.waitForTimeout(1000);
    
    // Check that back button is visible
    await expect(page.locator('button:has-text("Back")')).toBeVisible();
    
    // Test back button functionality - should go back to note page using browser history
    await page.click('button:has-text("Back")');
    await page.waitForTimeout(1000);
    
    // Should be back on note page
    await expect(page).toHaveURL(`http://localhost:5173/notes/${noteId}`);
  });

  test('should use browser back when history exists', async ({ page, request }) => {
    // Create a note via API first
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'History Test', content: 'Testing browser back functionality' }
    });
    const noteId = (await note.json()).id;
    const noteUrl = `http://localhost:5173/notes/${noteId}`;
    
    // Navigate to note page to create history
    await page.goto(noteUrl);
    await page.waitForTimeout(1000);
    
    // Navigate to home page to create history
    await page.goto('http://localhost:5173');
    await page.waitForTimeout(1000);
    
    // Go back to note page using browser history
    await page.goBack();
    await page.waitForTimeout(1000);
    await expect(page.locator('button:has-text("Back")')).toBeVisible();
    
    // Click back button - should use browser history to go home
    await page.click('button:has-text("Back")');
    await page.waitForTimeout(2000);
    // Check if we're back on home page
    await expect(page.locator('h1')).toHaveText('My Notes');
  });
});
