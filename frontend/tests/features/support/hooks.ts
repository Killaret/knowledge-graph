import { Before, After, BeforeAll, AfterAll } from '@cucumber/cucumber';
import { chromium, type Browser, type BrowserContext } from 'playwright';
import type { ITestWorld } from './world';

let browser: Browser;

BeforeAll(async function() {
  browser = await chromium.launch({
    headless: true,
    slowMo: 50
  });
});

AfterAll(async function() {
  await browser?.close();
});

Before(async function(this: ITestWorld) {
  // Create new context and page for each scenario
  this.context = await browser.newContext({
    viewport: { width: 1280, height: 720 }
  });
  this.page = await this.context.newPage();
  this.request = this.context.request;
  this.testNotes = [];
});

After(async function(this: ITestWorld) {
  // Cleanup test notes
  for (const note of this.testNotes) {
    try {
      await this.request.delete(`http://localhost:8080/notes/${note.id}`);
    } catch (e) {
      // Ignore cleanup errors
    }
  }
  
  // Close context
  await this.context?.close();
});
