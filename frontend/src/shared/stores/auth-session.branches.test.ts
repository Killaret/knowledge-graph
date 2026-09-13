import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

const mockEnv = vi.hoisted(() => ({ browser: true }));

vi.mock("$app/environment", () => mockEnv);

async function importAuthSession() {
  vi.resetModules();
  const mod = await import("./auth-session.svelte");
  return mod;
}

describe("auth-session branch coverage", () => {
  beforeEach(() => {
    mockEnv.browser = true;
    localStorage.clear();
    delete (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__;
    delete (import.meta.env as any).VITE_SKIP_AUTH;
    delete (import.meta.env as any).DEV;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("accessToken falls back to injected window token", async () => {
    const { accessToken } = await importAuthSession();

    expect(accessToken()).toBeNull();

    (window as { __ACCESS_TOKEN__?: string }).__ACCESS_TOKEN__ = "injected";
    expect(accessToken()).toBe("injected");

    delete (window as { __ACCESS_TOKEN__?: string }).__ACCESS_TOKEN__;
  });

  it("setApiKey and clearAuthState avoid localStorage when not in browser", async () => {
    mockEnv.browser = false;
    const { setApiKey, getApiKey, clearAuthState, authState } = await importAuthSession();

    setApiKey("secret");
    expect(getApiKey()).toBe("secret");
    expect(localStorage.getItem("api_key")).toBeNull();

    authState.currentUser = { id: "u1" } as any;
    clearAuthState();
    expect(getApiKey()).toBeNull();
    expect(authState.currentUser).toBeNull();
  });

  it("setSessionHint and hasSessionHint are no-ops when not in browser", async () => {
    mockEnv.browser = false;
    const { setSessionHint, hasSessionHint } = await importAuthSession();

    setSessionHint(true);
    expect(hasSessionHint()).toBe(false);
  });

  it("skipAuthMode and isAuthenticated respect VITE_SKIP_AUTH build flag", async () => {
    (import.meta.env as any).VITE_SKIP_AUTH = "true";
    const { skipAuthMode, isAuthenticated } = await importAuthSession();

    expect(skipAuthMode()).toBe(true);
    expect(isAuthenticated()).toBe(true);

    delete (import.meta.env as any).VITE_SKIP_AUTH;
  });

  it("skipAuthMode and isAuthenticated respect localStorage skip_auth in DEV", async () => {
    (import.meta.env as any).DEV = true;
    localStorage.setItem("__SKIP_AUTH__", "true");
    const { skipAuthMode, isAuthenticated } = await importAuthSession();

    expect(skipAuthMode()).toBe(true);
    expect(isAuthenticated()).toBe(true);
  });

  it("isAuthenticated parses skip_auth query param in DEV", async () => {
    const originalLocation = window.location;
    (import.meta.env as any).DEV = true;
    Object.defineProperty(window, "location", {
      value: { href: "http://127.0.0.1:3000/?skip_auth=true" },
      configurable: true,
      writable: true,
    });

    const { isAuthenticated } = await importAuthSession();

    expect(isAuthenticated()).toBe(true);
    expect(localStorage.getItem("__SKIP_AUTH__")).toBe("true");

    Object.defineProperty(window, "location", {
      value: originalLocation,
      configurable: true,
      writable: true,
    });
  });

  it("isAuthenticated is true for current user, api key or token", async () => {
    const { isAuthenticated, authState } = await importAuthSession();

    expect(isAuthenticated()).toBe(false);

    authState.currentUser = { id: "u1", role: "user" } as any;
    expect(isAuthenticated()).toBe(true);

    authState.currentUser = null;
    authState.apiKey = "key";
    expect(isAuthenticated()).toBe(true);

    authState.apiKey = null;
    authState.accessToken = "token";
    expect(isAuthenticated()).toBe(true);
  });

  it("isAdmin returns false when no user", async () => {
    const { isAdmin } = await importAuthSession();
    expect(isAdmin()).toBe(false);
  });
});
