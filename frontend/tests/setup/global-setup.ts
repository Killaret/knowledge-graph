import type { FullConfig } from "@playwright/test";

/**
 * Global setup for Playwright tests
 * Sets up SKIP_AUTH mode for testing
 */

async function globalSetup(_config: FullConfig) {
  // Default to SKIP_AUTH=true for tests, but allow an explicit override from env
  if (!process.env.SKIP_AUTH) {
    process.env.SKIP_AUTH = "true";
  }
  console.log(`[Global Setup] SKIP_AUTH=${process.env.SKIP_AUTH} for tests`);
}

export default globalSetup;
