import { defineConfig, devices } from '@playwright/test';

// Type for Node.js process
declare const process: {
  env: {
    CI?: string;
    FORCE3D?: string;
    FRONTEND_URL?: string;
    BACKEND_URL?: string;
    DATABASE_URL?: string;
    REDIS_URL?: string;
    SKIP_AUTH?: string;
  };
};

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  timeout: 60 * 1000, // 60s per test
  // globalSetup: './tests/setup/global-setup.ts', // Отключено чтобы избежать конфликта
  reporter: 'list',
  use: {
    baseURL: process.env.FRONTEND_URL || 'http://localhost:8081',
    trace: 'on-first-retry',
    actionTimeout: 15000,
    navigationTimeout: 15000,
    // Inject SKIP_AUTH flag for testing
    launchOptions: {
      args: ['--disable-web-security'],
    },
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
      },
      dependencies: ['setup'],
    },
  ],
  // Completely disable webServer - use existing Docker server
  webServer: undefined,
});
