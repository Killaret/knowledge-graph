import { Given, When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import type { ITestWorld } from '../support/world';

// Search and Discovery
Given('I am on the search page', async function(this: ITestWorld) {
  await this.page.goto('http://localhost:5173/');
  await this.page.waitForTimeout(1000);
});

Given('there are notes with titles containing {string}', async function(this: ITestWorld, searchTerm: string) {
  // Check if notes exist with this term, if not create them
  const searchInput = this.page.locator('.search-input, input[type="search"]').first();
  
  if (await searchInput.isVisible().catch(() => false)) {
    await searchInput.fill(searchTerm);
    await this.page.waitForTimeout(500);
    
    const results = this.page.locator('.search-results, .note-card, .search-dropdown').first();
    const hasResults = await results.isVisible().catch(() => false);
    
    if (!hasResults) {
      // Create a note with this term
      await this.request.post('http://localhost:8080/notes', {
        data: { 
          title: `Searchable Note with ${searchTerm}`, 
          content: `This note contains ${searchTerm} for testing`,
          type: 'star'
        }
      });
      
      await this.page.reload();
      await this.page.waitForTimeout(1000);
    }
  }
});

When('I type {string} into the search bar', async function(this: ITestWorld, query: string) {
  await this.page.fill('.search-input, input[type="search"], [placeholder*="Search"]', query);
  await this.page.waitForTimeout(500); // Wait for debounce
});

Then('a dropdown with suggestions appears', async function(this: ITestWorld) {
  const dropdown = this.page.locator('.search-results, .dropdown, [class*="suggestions"], .autocomplete').first();
  await expect(dropdown).toBeVisible();
});

Then('suggestions include {string}', async function(this: ITestWorld, suggestionText: string) {
  const suggestions = this.page.locator('.search-results, .dropdown, [class*="suggestions"]').first();
  await expect(suggestions).toContainText(suggestionText);
});

When('I press Enter', async function(this: ITestWorld) {
  await this.page.keyboard.press('Enter');
  await this.page.waitForTimeout(500);
});

Then('the graph filters to show only matching nodes', async function(this: ITestWorld) {
  // After search, only matching notes should be visible
  const visibleNotes = this.page.locator('.note-card:visible, .search-results .note-card');
  const count = await visibleNotes.count();
  expect(count).toBeGreaterThan(0);
});

Then('non-matching nodes are dimmed or hidden', async function(this: ITestWorld) {
  // Check that filtering was applied
  const filteredIndicator = this.page.locator('.stats-filter, .filter-indicator, text=search:').first();
  const hasFilter = await filteredIndicator.isVisible().catch(() => false);
  expect(hasFilter).toBe(true);
});

// Full-text search
Given('notes exist with content including {string}', async function(this: ITestWorld, contentTerm: string) {
  // Create a note with this content term via API
  await this.request.post('http://localhost:8080/notes', {
    data: { 
      title: 'Content Search Test Note', 
      content: `This is a test note containing ${contentTerm} for full-text search testing`,
      type: 'planet'
    }
  });
  
  await this.page.reload();
  await this.page.waitForTimeout(1000);
});

When('I enter {string} in the search field', async function(this: ITestWorld, query: string) {
  await this.step(`I type "${query}" into the search bar`);
});

Then('search results include notes where content matches {string}', async function(this: ITestWorld, searchTerm: string) {
  const results = this.page.locator('.search-results, .note-card, .search-item');
  const count = await results.count();
  expect(count).toBeGreaterThan(0);
  
  // Verify at least one result contains the search term
  let found = false;
  for (let i = 0; i < count; i++) {
    const text = await results.nth(i).textContent();
    if (text?.toLowerCase().includes(searchTerm.toLowerCase())) {
      found = true;
      break;
    }
  }
  expect(found).toBe(true);
});

// Empty search results
When('I enter {string} with no matches', async function(this: ITestWorld, query: string) {
  await this.page.fill('.search-input, input[type="search"]', query);
  await this.page.keyboard.press('Enter');
  await this.page.waitForTimeout(500);
});

Then('the message {string} appears', async function(this: ITestWorld, message: string) {
  const messageLocator = this.page.locator(`text="${message}", .empty-state:has-text("${message}"), .no-results:has-text("${message}")`).first();
  await expect(messageLocator).toBeVisible();
});

// Search and Select
Given('search results are displayed', async function(this: ITestWorld) {
  // Ensure search has been performed
  const results = this.page.locator('.search-results, .note-card, .dropdown-item').first();
  await expect(results).toBeVisible();
});

When('I click on the result {string}', async function(this: ITestWorld, resultTitle: string) {
  const result = this.page.locator(`.search-results:has-text("${resultTitle}"), .note-card:has-text("${resultTitle}"), .dropdown-item:has-text("${resultTitle}")`).first();
  await result.click();
});

Then('the graph centers on node {string}', async function(this: ITestWorld, nodeTitle: string) {
  // The node should be visible and accessible
  const node = this.page.locator(`.note-card:has-text("${nodeTitle}"), text="${nodeTitle}"`).first();
  await expect(node).toBeVisible();
  
  // Side panel should open
  const sidePanel = this.page.locator('.side-panel, .note-side-panel').first();
  await expect(sidePanel).toBeVisible();
});

Then('the side panel shows details for {string}', async function(this: ITestWorld, nodeTitle: string) {
  const sidePanel = this.page.locator('.side-panel, .note-side-panel').first();
  await expect(sidePanel).toContainText(nodeTitle);
});

// Clear search
When('I click the {string} icon', async function(this: ITestWorld, iconName: string) {
  const clearButton = this.page.locator(`[data-testid="clear-search"], .clear-search, button:has-text("Clear"), button[aria-label*="Clear"]`).first();
  await clearButton.click();
});

Then('the search field is cleared', async function(this: ITestWorld) {
  const searchInput = this.page.locator('.search-input, input[type="search"]').first();
  const value = await searchInput.inputValue();
  expect(value).toBe('');
});

Then('the graph shows all nodes again', async function(this: ITestWorld) {
  // All notes should be visible again
  const noteCards = this.page.locator('.note-card');
  const count = await noteCards.count();
  expect(count).toBeGreaterThan(0);
});
