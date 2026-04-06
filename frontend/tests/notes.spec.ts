import { test, expect } from '@playwright/test';

test.describe('Knowledge Graph Frontend', () => {
  test.beforeEach(async ({ page }) => {
    // Очистка БД через API (опционально)
    await page.goto('http://localhost:5173');
  });

  test('should create a new note', async ({ page }) => {
    await page.click('a:has-text("+ New Note")');
    await page.fill('input[placeholder="Title"]', 'Playwright Test');
    await page.fill('textarea', 'Automated content');
    await page.click('button:has-text("Create")');
    await expect(page).toHaveURL(/\/notes\/[a-f0-9-]+/);
    await expect(page.locator('h1')).toHaveText('Playwright Test');
  });

  test('should edit a note', async ({ page }) => {
    // Сначала создадим заметку через API или UI
    await page.click('a:has-text("+ New Note")');
    await page.fill('input[placeholder="Title"]', 'To Edit');
    await page.fill('textarea', 'Original');
    await page.click('button:has-text("Create")');
    await expect(page).toHaveURL(/\/notes\/[a-f0-9-]+/);

    await page.click('a:has-text("Edit")');
    await page.fill('input[placeholder="Title"]', 'Edited');
    await page.fill('textarea', 'New content');
    await page.click('button:has-text("Update")');
    await expect(page.locator('h1')).toHaveText('Edited');
    await expect(page.locator('.content')).toHaveText('New content');
  });

  test('should delete a note', async ({ page }) => {
    await page.click('a:has-text("+ New Note")');
    await page.fill('input[placeholder="Title"]', 'To Delete');
    await page.click('button:has-text("Create")');
    await page.waitForURL(/\/notes\/[a-f0-9-]+/);
    page.on('dialog', dialog => dialog.accept());
    await page.click('button:has-text("Delete")');
    await expect(page).toHaveURL('http://localhost:5173/');
    await expect(page.locator('text=To Delete')).not.toBeVisible();
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
});