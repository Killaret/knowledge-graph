import { defineConfig, devices } from '@playwright/test';

// Type for Node.js process
declare const process: {
  env: {
    CI?: string;
    FORCE3D?: string;
    FRONTEND_URL?: string;
    BACKEND_URL?: string;
  };
};

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  timeout: 60 * 1000, // 60s per test
  reporter: 'html',
  use: {
    baseURL: process.env.FRONTEND_URL || 'http://localhost:5173',
    trace: 'on-first-retry',
    actionTimeout: 15000,
    navigationTimeout: 15000,
    // Expose backend URL for API requests
    extraHTTPHeaders: {
      'X-BACKEND-URL': process.env.BACKEND_URL || 'http://localhost:8080',
    },
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: process.env.BACKEND_URL
    ? {
        command: 'npm run dev',
        url: process.env.FRONTEND_URL || 'http://localhost:5173',
        reuseExistingServer: !process.env.CI,
        timeout: process.env.CI ? 180 * 1000 : 120 * 1000,
      }
    : [
        {
          command: 'npm run dev',
          url: process.env.FRONTEND_URL || 'http://localhost:5173',
          reuseExistingServer: !process.env.CI,
          timeout: process.env.CI ? 180 * 1000 : 120 * 1000,
        },
        {
          command: 'cd ../backend && go run cmd/server/main.go',
          url: 'http://localhost:8080',
          reuseExistingServer: !process.env.CI,
          timeout: 60 * 1000,
        },
      ],
});
