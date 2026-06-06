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
  timeout: 90 * 1000, // 90s per test
  globalSetup: './tests/setup/global-setup.ts',
  reporter: 'html',
  use: {
    baseURL: process.env.FRONTEND_URL || 'http://localhost:5173',
    trace: 'on-first-retry',
    actionTimeout: 30000,
    navigationTimeout: 30000,
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
    // Default project: all tests EXCEPT auth-functional (auth_skipped=true)
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        // Inject SKIP_AUTH flag for all tests
        launchOptions: {
          args: ['--disable-web-security'],
        },
      },
      // Exclude auth-functional tests — they require real authentication
      grepInvert: /auth-functional/,
      dependencies: ['setup'],
    },
    // Auth project: only auth-functional tests (auth_skipped=false)
    {
      name: 'chromium-auth',
      use: {
        ...devices['Desktop Chrome'],
        // No SKIP_AUTH injection — real auth required
        launchOptions: {
          args: ['--disable-web-security'],
        },
      },
      // Only auth-functional tests
      grep: /auth-functional/,
      // Note: this project requires backend with SKIP_AUTH=false
      // Run separately: SKIP_AUTH=false docker compose up backend
    },
  ],
  // Auto-start dev server for tests - DISABLED for Docker usage
  // webServer: {
  //   command: 'npm run dev',
  //   url: 'http://localhost:5173',
  //   reuseExistingServer: true,
  //   timeout: 120 * 1000,
  //   env: {
  //     SKIP_AUTH: 'true',
  //   },
  // },
});
