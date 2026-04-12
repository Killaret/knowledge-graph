import { Before, After, BeforeAll, AfterAll } from '@cucumber/cucumber';
import type { ITestWorld } from './world';

// Global setup before all tests
BeforeAll(async function() {
  // Any global setup - e.g., starting test server
  console.log('Starting BDD test suite...');
});

// Cleanup after all tests
AfterAll(async function() {
  console.log('BDD test suite completed.');
});

// Setup before each scenario
Before(async function(this: ITestWorld) {
  await this.setup();

  // Set base URL from environment or default
  const baseUrl = process.env.BASE_URL || 'http://localhost:5173';

  // Navigate to base page and wait for it to load
  await this.page.goto(baseUrl);
  await this.page.waitForLoadState('networkidle');
});

// Cleanup after each scenario
After(async function(this: ITestWorld, scenario) {
  // Take screenshot if scenario failed
  if (scenario.result?.status === 'FAILED') {
    const screenshotPath = `./test-results/screenshots/${scenario.pickle.name.replace(/[^a-z0-9]/gi, '_')}.png`;
    await this.page.screenshot({ path: screenshotPath, fullPage: true });
    console.log(`Screenshot saved to ${screenshotPath}`);
  }

  await this.teardown();
});
