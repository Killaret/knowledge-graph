import type { APIRequestContext, Page } from "@playwright/test";

export function getBackendUrl(): string {
  return process.env.BACKEND_URL || "http://127.0.0.1:18083";
}

const TEST_USER = {
  login: "testuser",
  email: "testuser@example.com",
  password: "TestPassword123!",
};

/**
 * Log in as the seeded test user and return an access token.
 * If the user does not exist, register them first.
 */
export async function loginOrCreateTestUser(request: APIRequestContext): Promise<string> {
  const backendUrl = getBackendUrl();

  // Try to log in first
  const loginResp = await request.post(`${backendUrl}/api/v1/auth/login`, {
    data: {
      login: TEST_USER.login,
      password: TEST_USER.password,
    },
  });

  if (loginResp.ok()) {
    const tokens = await loginResp.json();
    return tokens.access_token as string;
  }

  // User may not exist; register and then log in
  const registerResp = await request.post(`${backendUrl}/api/v1/auth/register`, {
    data: {
      login: TEST_USER.login,
      email: TEST_USER.email,
      password: TEST_USER.password,
    },
  });

  if (!registerResp.ok()) {
    const errorText = await registerResp.text();
    throw new Error(`Failed to register test user: ${registerResp.status()} - ${errorText}`);
  }

  // Login after registration
  const secondLoginResp = await request.post(`${backendUrl}/api/v1/auth/login`, {
    data: {
      login: TEST_USER.login,
      password: TEST_USER.password,
    },
  });

  if (!secondLoginResp.ok()) {
    const errorText = await secondLoginResp.text();
    throw new Error(
      `Failed to login test user after registration: ${secondLoginResp.status()} - ${errorText}`
    );
  }

  const tokens = await secondLoginResp.json();
  return tokens.access_token as string;
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
