import { defineConfig, devices } from '@playwright/test';
import { createArgosReporterOptions } from '@argos-ci/playwright/reporter';

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
    ARGOS_TOKEN?: string;
    ARGOS_UPLOAD_LOCAL?: string;
  };
};

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 2 : undefined,
  timeout: 120 * 1000, // 120s per test (increased from 90s)
  globalSetup: './tests/setup/global-setup.ts',
  reporter: [
    ['line'],
    ['@argos-ci/playwright/reporter', createArgosReporterOptions({
      uploadToArgos: !!process.env.CI || !!process.env.ARGOS_UPLOAD_LOCAL,
      token: process.env.ARGOS_TOKEN,
    })],
  ],
  use: {
    baseURL: process.env.FRONTEND_URL || 'http://localhost:5173',
    trace: 'on-first-retry',
    actionTimeout: 60000, // Increased from 30000ms
    navigationTimeout: 60000, // Increased from 30000ms
    bypassCSP: true, // Required for Argos stabilization script injection
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
    // Default project: all tests (auth-functional will auto-skip if backend has SKIP_AUTH)
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        // Inject SKIP_AUTH flag for all tests
        launchOptions: {
          args: ['--disable-web-security'],
        },
      },
      // Note: auth-functional tests will auto-skip if backend has SKIP_AUTH enabled
      // Skip 3D and visual tests in default project
      grepInvert: /@3d|@visual/,
      dependencies: ['setup'],
    },
    // Visual regression project (runs only @visual tests)
    {
      name: 'visual',
      use: {
        ...devices['Desktop Chrome'],
        launchOptions: {
          args: ['--disable-web-security'],
        },
      },
      grep: /@visual/,
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
  // Auto-start dev server for tests - enable with PLAYWRIGHT_DEV_SERVER=true
  webServer: process.env.PLAYWRIGHT_DEV_SERVER === 'true' ? {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: true,
    timeout: 120 * 1000,
    env: {
      SKIP_AUTH: 'true',
    },
  } : undefined,
});

