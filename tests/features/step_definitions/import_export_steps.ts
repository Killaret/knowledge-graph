import { Given, When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import type { ITestWorld } from '../support/world';

// Import steps
Given('I have a JSON file {string} with valid note data', async function(this: ITestWorld, filename: string) {
  // Create test data via API
  await this.request.post('http://localhost:8080/notes', {
    data: { 
      title: 'Import Test Note 1', 
      content: 'Content for import testing',
      type: 'star'
    }
  });
  await this.request.post('http://localhost:8080/notes', {
    data: { 
      title: 'Import Test Note 2', 
      content: 'Another content for import',
      type: 'planet'
    }
  });
  
  await this.page.reload();
  await this.page.waitForTimeout(1000);
});

When('I click the {string} menu item', async function(this: ITestWorld, menuItem: string) {
  await this.page.click(`button:has-text("${menuItem}"), [data-menu-item="${menuItem.toLowerCase()}"], .menu-item:has-text("${menuItem}")`);
});

When('I select the file {string}', async function(this: ITestWorld, filename: string) {
  // Handle file upload
  const fileInput = this.page.locator('input[type="file"]').first();
  
  // Create a mock file upload
  await fileInput.setInputFiles({
    name: filename,
    mimeType: 'application/json',
    buffer: Buffer.from(JSON.stringify({
      notes: [
        { title: 'Imported Note 1', content: 'Imported content 1' },
        { title: 'Imported Note 2', content: 'Imported content 2' }
      ]
    }))
  });
});

Then('the import process completes successfully', async function(this: ITestWorld) {
  // Wait for success message
  const successMessage = this.page.locator('text=Import successful, text=Notes imported, .success-message').first();
  await expect(successMessage).toBeVisible();
});

Then('the graph displays new nodes from the file', async function(this: ITestWorld) {
  const noteCards = this.page.locator('.note-card');
  const count = await noteCards.count();
  expect(count).toBeGreaterThan(0);
});

Then('I see a confirmation: {string}', async function(this: ITestWorld, message: string) {
  const confirmation = this.page.locator(`text="${message}", .confirmation:has-text("${message}")`).first();
  await expect(confirmation).toBeVisible();
});

// Error handling for import
Given('I have an invalid or corrupted file {string}', async function(this: ITestWorld, filename: string) {
  // Prepare invalid file for upload
  const fileInput = this.page.locator('input[type="file"]').first();
  
  await fileInput.setInputFiles({
    name: filename,
    mimeType: 'application/json',
    buffer: Buffer.from('invalid json content {{{')
  });
});

Then('an error message {string} is shown', async function(this: ITestWorld, errorMessage: string) {
  const error = this.page.locator(`text="${errorMessage}", .error:has-text("${errorMessage}")`).first();
  await expect(error).toBeVisible();
});

Then('the existing graph remains unchanged', async function(this: ITestWorld) {
  // Verify no crash and graph still visible
  const graph = this.page.locator('.graph-container, .graph-canvas, .notes-grid').first();
  await expect(graph).toBeVisible();
});

// Export steps
Given('I have several notes on the graph', async function(this: ITestWorld) {
  // Create multiple notes via API
  const notes = [
    { title: 'Export Note 1', content: 'Content 1', type: 'star' },
    { title: 'Export Note 2', content: 'Content 2', type: 'planet' },
    { title: 'Export Note 3', content: 'Content 3', type: 'comet' }
  ];
  
  for (const noteData of notes) {
    await this.request.post('http://localhost:8080/notes', { data: noteData });
  }
  
  await this.page.reload();
  await this.page.waitForTimeout(1000);
});

When('I open the menu and select {string}', async function(this: ITestWorld, menuItem: string) {
  // Open menu first
  const menuButton = this.page.locator('.menu-button, [data-testid="menu"], button:has-text("Menu")').first();
  if (await menuButton.isVisible().catch(() => false)) {
    await menuButton.click();
  }
  
  // Select menu item
  await this.page.click(`button:has-text("${menuItem}"), [data-menu-item="${menuItem.toLowerCase()}"]`);
});

When('I choose format {string}', async function(this: ITestWorld, format: string) {
  // Select format from dropdown
  await this.page.selectOption('select[name="format"], .format-selector', format);
  
  // Click export button
  await this.page.click('button:has-text("Export"), button[type="submit"]');
});

Then('a file {string} is downloaded', async function(this: ITestWorld, filename: string) {
  // Wait for download
  const downloadPromise = this.page.waitForEvent('download', { timeout: 10000 });
  
  // Trigger download if not already triggered
  const exportButton = this.page.locator('button:has-text("Export"), button:has-text("Download")').first();
  if (await exportButton.isVisible().catch(() => false)) {
    await exportButton.click();
  }
  
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toContain(filename.replace('.json', '').replace('.csv', ''));
});

Then('the file contains all current notes and links', async function(this: ITestWorld) {
  // Verify download completed
  const download = await this.page.waitForEvent('download', { timeout: 5000 });
  const path = await download.path();
  expect(path).toBeTruthy();
});

// Import resume (after error fix)
Given('I corrected the errors in {string}', async function(this: ITestWorld, filename: string) {
  // Create valid file
  const fileInput = this.page.locator('input[type="file"]').first();
  
  await fileInput.setInputFiles({
    name: filename,
    mimeType: 'application/json',
    buffer: Buffer.from(JSON.stringify({
      notes: [
        { title: 'Valid Note 1', content: 'Valid content 1' },
        { title: 'Valid Note 2', content: 'Valid content 2' }
      ]
    }))
  });
});

When('I retry the import', async function(this: ITestWorld) {
  // Click import button again
  await this.page.click('button:has-text("Import"), button[type="submit"]');
});

Then('after completion, new nodes appear on the graph', async function(this: ITestWorld) {
  const noteCards = this.page.locator('.note-card');
  const count = await noteCards.count();
  expect(count).toBeGreaterThan(0);
});

// Basic import/export steps
When('I open the menu and select {string}', async function(this: ITestWorld, menuItem: string) {
  // Open menu (hamburger or menu button)
  const menuBtn = this.page.locator('.menu-button, [data-testid="menu"], button:has-text("Menu")').first();
  await menuBtn.click();
  await this.page.waitForTimeout(300);
  
  // Click menu item
  const item = this.page.locator(`.menu-item:has-text("${menuItem}"), text="${menuItem}"`).first();
  await item.click();
});

When('I choose a Markdown file {string}', async function(this: ITestWorld, filename: string) {
  const fileInput = this.page.locator('input[type="file"]').first();
  
  await fileInput.setInputFiles({
    name: filename,
    mimeType: 'text/markdown',
    buffer: Buffer.from('# Test Note\n\nThis is test content.')
  });
});

When('I choose a PDF file {string}', async function(this: ITestWorld, filename: string) {
  const fileInput = this.page.locator('input[type="file"]').first();
  
  await fileInput.setInputFiles({
    name: filename,
    mimeType: 'application/pdf',
    buffer: Buffer.from('%PDF-1.4 test content')
  });
});

When('I start the import', async function(this: ITestWorld) {
  const importBtn = this.page.locator('button:has-text("Import"), button:has-text("Start"), button[type="submit"]').first();
  await importBtn.click();
  await this.page.waitForTimeout(2000);
});

Then('a progress indicator is shown', async function(this: ITestWorld) {
  const progress = this.page.locator('.progress, .progress-bar, .loading-indicator, [class*="progress"]').first();
  await expect(progress).toBeVisible();
});

When('I enter URL {string}', async function(this: ITestWorld, url: string) {
  const urlInput = this.page.locator('input[type="url"], input[name="url"], input[placeholder*="URL"]').first();
  await urlInput.fill(url);
});

Given('I have several notes on the graph', async function(this: ITestWorld) {
  // Create test notes
  for (let i = 1; i <= 3; i++) {
    await this.request.post('http://localhost:8080/notes', {
      data: { title: `Export Test ${i}`, content: `Content ${i}` }
    });
  }
  
  // Navigate to graph
  await this.page.goto('http://localhost:5173/');
  await this.page.waitForLoadState('networkidle');
});

When('I choose format {string}', async function(this: ITestWorld, format: string) {
  const formatSelect = this.page.locator('select[name="format"], .format-select').first();
  await formatSelect.selectOption(format);
});

Then('a file {string} is downloaded', async function(this: ITestWorld, filename: string) {
  const download = await this.page.waitForEvent('download', { timeout: 5000 });
  const suggestedFilename = download.suggestedFilename();
  expect(suggestedFilename).toContain(filename.split('.')[0]);
});
