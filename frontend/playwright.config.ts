import { defineConfig, devices } from "@playwright/test";
import { createArgosReporterOptions } from "@argos-ci/playwright/reporter";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

// Load Argos token from gitignored argos.json if ARGOS_TOKEN is not set.
// This lets local runs upload screenshots without hardcoding the secret in the repo.
if (!process.env.ARGOS_TOKEN) {
  try {
    const argosConfigPath = resolve("argos.json");
    if (existsSync(argosConfigPath)) {
      const argosConfig = JSON.parse(readFileSync(argosConfigPath, "utf-8"));
      if (typeof argosConfig?.token === "string") {
        process.env.ARGOS_TOKEN = argosConfig.token;
      }
    }
  } catch {
    // Ignore missing or malformed argos.json
  }
}

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
  testDir: "./tests",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 2 : undefined,
  timeout: 120 * 1000, // 120s per test (increased from 90s)
  globalSetup: "./tests/setup/global-setup.ts",
  reporter: [
    ["line"],
    [
      "@argos-ci/playwright/reporter",
      createArgosReporterOptions({
        uploadToArgos:
          (!!process.env.CI && !!process.env.ARGOS_TOKEN) || !!process.env.ARGOS_UPLOAD_LOCAL,
        token: process.env.ARGOS_TOKEN,
      }),
    ],
  ],
  use: {
    baseURL: process.env.FRONTEND_URL || "http://localhost:5173",
    trace: "on-first-retry",
    actionTimeout: 60000, // Increased from 30000ms
    navigationTimeout: 60000, // Increased from 30000ms
    bypassCSP: true, // Required for Argos stabilization script injection
    // Inject SKIP_AUTH flag for testing
    launchOptions: {
      args: ["--disable-web-security"],
    },
  },
  projects: [
    // Setup project for auth bypass (used by skip-auth projects)
    {
      name: "setup",
      testMatch: "**/setup/*.setup.ts",
    },
    // SKIP_AUTH mode project: skip-auth tests only, excludes real-auth tests
    {
      name: "chromium-skip-auth",
      use: {
        ...devices["Desktop Chrome"],
        launchOptions: {
          args: ["--disable-web-security"],
        },
      },
      grepInvert: /@3d|@visual|@auth-real|@manual|@canvas/,
      dependencies: ["setup"],
    },
    // Real auth project: only @auth-real tests (requires backend SKIP_AUTH=false)
    {
      name: "chromium-real-auth",
      use: {
        ...devices["Desktop Chrome"],
        launchOptions: {
          args: ["--disable-web-security"],
        },
      },
      grep: /@auth-real/,
    },
    // Visual regression project (runs only @visual tests)
    {
      name: "visual",
      use: {
        ...devices["Desktop Chrome"],
        launchOptions: {
          args: ["--disable-web-security"],
        },
      },
      grep: /@visual/,
      dependencies: ["setup"],
    },
  ],
  // Auto-start dev server for tests - enable with PLAYWRIGHT_DEV_SERVER=true
  webServer:
    process.env.PLAYWRIGHT_DEV_SERVER === "true"
      ? {
          command: "npm run dev",
          url: "http://localhost:5173",
          reuseExistingServer: true,
          timeout: 120 * 1000,
          env: {
            SKIP_AUTH: "true",
          },
        }
      : undefined,
});
