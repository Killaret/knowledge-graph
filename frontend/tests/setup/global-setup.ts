import type { FullConfig } from '@playwright/test';

/**
 * Global setup for Playwright tests
 * Sets up SKIP_AUTH mode for testing
 */

async function globalSetup(config: FullConfig) {
  // Set SKIP_AUTH environment variable
  process.env.SKIP_AUTH = 'true';
  console.log('[Global Setup] SKIP_AUTH enabled for tests');
}

export default globalSetup;
