import { Before, After, BeforeAll, AfterAll } from "@cucumber/cucumber";
import { chromium, type Browser } from "playwright";
import { spawn, type ChildProcess } from "child_process";
import type { ITestWorld } from "./world";
import { loginOrCreateTestUser } from "../../helpers/auth";

let browser: Browser;
let devServer: ChildProcess | null = null;

async function waitForServer(url: string, timeout = 60000): Promise<void> {
  const start = Date.now();
  while (Date.now() - start < timeout) {
    try {
      const response = await fetch(url);
      if (response.status === 200) return;
    } catch {
      // Server not ready yet
    }
    await new Promise((r) => setTimeout(r, 500));
  }
  throw new Error(`Server at ${url} did not start within ${timeout}ms`);
}

BeforeAll(async function () {
  // Start Vite dev server if not already running
  const frontendUrl = process.env.FRONTEND_URL || "http://localhost:5173";
  try {
    await fetch(frontendUrl);
    // Server already running
  } catch {
    // Start the dev server
    devServer = spawn("npm", ["run", "dev"], {
      cwd: process.cwd(),
      stdio: "pipe",
      shell: true,
      detached: false,
    });
    // Wait for server to be ready
    await waitForServer(frontendUrl);
  }

  browser = await chromium.launch({
    headless: true,
    slowMo: 50,
  });
});

AfterAll(async function () {
  await browser?.close();
  // Kill dev server if we started it
  if (devServer) {
    devServer.kill("SIGTERM");
    // Force kill if still running
    setTimeout(() => devServer?.kill("SIGKILL"), 5000);
  }
});

Before(async function (this: ITestWorld) {
  let extraHeaders: Record<string, string> | undefined;

  // For real-auth stacks, log in as the seeded test user and inject the
  // access token as a default header for both page and API requests.
  let accessToken: string | undefined;
  if (process.env.SKIP_AUTH !== "true") {
    const tempContext = await browser.newContext({
      viewport: { width: 1280, height: 720 },
    });
    const tempRequest = tempContext.request;
    accessToken = await loginOrCreateTestUser(tempRequest);
    await tempContext.close();
    extraHeaders = { Authorization: `Bearer ${accessToken}` };
  }

  // Create new context and page for each scenario
  this.context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    extraHTTPHeaders: extraHeaders,
  });

  // Add SKIP_AUTH init script only when running against a SKIP_AUTH stack
  if (process.env.SKIP_AUTH === "true") {
    await this.context.addInitScript(() => {
      try {
        localStorage.setItem("__SKIP_AUTH__", "true");
      } catch {
        // localStorage may be unavailable in some contexts (e.g. about:blank)
      }
      (window as any).__SKIP_AUTH__ = true;
    });
  } else if (accessToken) {
    // For real-auth stacks, expose the access token to the frontend so
    // auth init can bootstrap the user session without a refresh cookie.
    await this.context.addInitScript((token: string) => {
      (window as any).__ACCESS_TOKEN__ = token;
    }, accessToken);
  }

  this.page = await this.context.newPage();
  this.page.on("console", (msg) => {
    if (msg.type() === "error" || msg.type() === "warning") {
      console.log(`[BROWSER ${msg.type().toUpperCase()}]`, msg.text());
    }
  });
  this.request = this.context.request;
  this.testNotes = [];
});

After(async function (this: ITestWorld) {
  // Cleanup test notes
  for (const note of this.testNotes) {
    try {
      await this.request.delete(
        `${process.env.BACKEND_URL || "http://127.0.0.1:18083"}/api/v1/notes/${note.id}`
      );
    } catch {
      // Ignore cleanup errors
    }
  }

  // Close context
  await this.context?.close();
});
