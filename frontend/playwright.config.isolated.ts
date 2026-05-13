import { defineConfig, devices } from '@playwright/test';
import type { PlaywrightTestConfig } from '@playwright/test';

/**
 * Isolated Playwright Configuration for Docker Environment
 * Completely isolated from Vitest to avoid dependency conflicts
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:8081',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    // Setup project for auth bypass
    {
      name: 'setup',
      testMatch: '**/setup/*.setup.ts',
    },
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        // Inject SKIP_AUTH flag for all tests
        launchOptions: {
          args: ['--disable-web-security'],
        },
      },
      dependencies: ['setup'],
    },
  ],
  // Completely disable webServer - use existing Docker server
  webServer: undefined,
});
