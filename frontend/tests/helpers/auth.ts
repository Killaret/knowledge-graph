import type { APIRequestContext, Page } from "@playwright/test";

export function getBackendUrl(): string {
  return process.env.BACKEND_URL || "http://127.0.0.1:18083";
}

const TEST_USER = {
  login: "testuser",
  email: "testuser@example.com",
  password: "TestPassword123!",
};

export const BDD_USER = {
  login: "bdduser",
  email: "bdduser@example.com",
  password: "BDDPassword123!",
};

async function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function loginOrCreateUser(
  request: APIRequestContext,
  user: { login: string; email: string; password: string }
): Promise<string> {
  const backendUrl = getBackendUrl();
  const maxAttempts = 5;

  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    // Try to log in first
    const loginResp = await request.post(`${backendUrl}/api/v1/auth/login`, {
      data: {
        login: user.login,
        password: user.password,
      },
    });

    if (loginResp.ok()) {
      const tokens = await loginResp.json();
      return tokens.access_token as string;
    }

    // User may not exist; register and then log in
    const registerResp = await request.post(`${backendUrl}/api/v1/auth/register`, {
      data: {
        login: user.login,
        email: user.email,
        password: user.password,
      },
    });

    if (registerResp.ok()) {
      const secondLoginResp = await request.post(`${backendUrl}/api/v1/auth/login`, {
        data: {
          login: user.login,
          password: user.password,
        },
      });

      if (secondLoginResp.ok()) {
        const tokens = await secondLoginResp.json();
        return tokens.access_token as string;
      }

      const errorText = await secondLoginResp.text();
      throw new Error(
        `Failed to login user after registration: ${secondLoginResp.status()} - ${errorText}`
      );
    }

    const status = registerResp.status();
    const errorText = await registerResp.text();

    // Another worker may have created the user in the meantime; retry login.
    if (status === 409 || status === 500) {
      await sleep(100 * (attempt + 1));
      continue;
    }

    throw new Error(`Failed to register user: ${status} - ${errorText}`);
  }

  throw new Error(`Failed to login or create user after ${maxAttempts} attempts`);
}

/**
 * Log in as the seeded test user and return an access token.
 * If the user does not exist, register them first.
 *
 * Uses a small retry loop so parallel Playwright workers do not race when
 * creating the shared test user.
 */
export async function loginOrCreateTestUser(request: APIRequestContext): Promise<string> {
  return loginOrCreateUser(request, TEST_USER);
}

/**
 * Log in as the BDD test user and return an access token.
 * BDD uses a separate account from the seeded manual-test user so that
 * scenario cleanup does not wipe the seeded test data.
 */
export async function loginOrCreateBDDUser(request: APIRequestContext): Promise<string> {
  return loginOrCreateUser(request, BDD_USER);
}

/**
 * Log in as the seeded test user and prepare the Playwright page for an
 * authenticated session. The access token is exposed to the frontend via
 * window.__ACCESS_TOKEN__ so auth init can bootstrap without a refresh cookie.
 */
export async function loginAsTestUser(page: Page, request: APIRequestContext): Promise<string> {
  const token = await loginOrCreateTestUser(request);

  // Inject the token into the page context so it is available before scripts run.
  // Use page-level init script to ensure it runs on the next navigation of the
  // existing Playwright page; context-level init script covers any new pages.
  await page.addInitScript((t: string) => {
    (window as any).__ACCESS_TOKEN__ = t;
  }, token);
  await page.context().addInitScript((t: string) => {
    (window as any).__ACCESS_TOKEN__ = t;
  }, token);

  return token;
}
