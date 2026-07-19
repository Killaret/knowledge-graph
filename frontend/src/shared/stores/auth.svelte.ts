// Auth store with Svelte 5 runes
import { browser } from "$app/environment";
import { goto } from "$app/navigation";
import * as authApi from "$shared/api/auth";
import * as usersApi from "$shared/api/users";
import {
  clearPreloadCache,
  preloadAuthenticatedGraph,
} from "$shared/services/PreloadService";
import { setLocale, type Locale } from "$shared/utils/i18n";
import type { User, AuthTokens } from "$shared/types";

// Global reactive state - wrapped in object for export
const authState = $state({
  currentUser: null as User | null,
  accessToken: null as string | null,
  refreshToken: null as string | null,
  isInitialized: false,
  isLoading: false,
  error: null as string | null,
  apiKey: null as string | null,
});

// Guard concurrent initAuth calls so refresh token is not used twice
let initAuthPromise: Promise<void> | null = null;

// Export reactive state through getter functions
export function currentUser(): User | null {
  return authState.currentUser;
}
export function accessToken(): string | null {
  return authState.accessToken;
}
export function refreshToken(): string | null {
  return authState.refreshToken;
}
export function isInitialized(): boolean {
  return authState.isInitialized;
}
export function isLoading(): boolean {
  return authState.isLoading;
}
export function error(): string | null {
  return authState.error;
}
export function apiKey(): string | null {
  return authState.apiKey;
}

/**
 * Apply user settings like locale after login/init.
 */
async function applyUserSettings(): Promise<void> {
  try {
    const { settings } = await usersApi.getSettings();
    const localeSetting = settings.find((s) => s.key === "preferred_language");
    if (
      localeSetting?.value &&
      (localeSetting.value === "en" || localeSetting.value === "ru")
    ) {
      setLocale(localeSetting.value as Locale);
    }
  } catch (e) {
    // Settings are not critical for auth flow
    console.warn("Failed to load user settings:", e);
  }
}

/**
 * Check if SKIP_AUTH mode is enabled.
 * The window flag __SKIP_AUTH__ is respected in all builds to support isolated visual regression tests.
 * LocalStorage and query params remain dev-only to avoid accidental production bypass.
 */
export function skipAuthMode(): boolean {
  if (!browser) return false;
  if ((window as any).__SKIP_AUTH__ === true) return true;
  if (!import.meta.env.DEV) return false;
  return localStorage.getItem("__SKIP_AUTH__") === "true";
}

// LocalStorage keys
const ACCESS_TOKEN_KEY = "access_token";
const REFRESH_TOKEN_KEY = "refresh_token";
const API_KEY = "api_key";

/**
 * Initialize auth state from localStorage
 */
export async function initAuth(): Promise<void> {
  if (!browser) {
    authState.isInitialized = true;
    return;
  }

  if (initAuthPromise) {
    return initAuthPromise;
  }

  initAuthPromise = (async () => {
    try {
      // Check for SKIP_AUTH mode from query parameter on first load (dev only)
      if (import.meta.env.DEV) {
        const url = new URL(window.location.href);
        if (url.searchParams.get("skip_auth") === "true") {
          // Persist SKIP_AUTH to localStorage
          localStorage.setItem("__SKIP_AUTH__", "true");
          console.log("[Auth] SKIP_AUTH mode enabled via query param");
        }
      }

      // Try to load tokens from localStorage
      const storedAccessToken = localStorage.getItem(ACCESS_TOKEN_KEY);
      const storedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
      const storedApiKey = localStorage.getItem(API_KEY);

      authState.apiKey = storedApiKey;
      authState.accessToken = storedAccessToken;
      authState.refreshToken = storedRefreshToken;

      if (!storedRefreshToken) {
        // Stale access token alone is not enough to be authenticated.
        authState.accessToken = null;
        authState.refreshToken = null;
        authState.currentUser = null;
        if (browser) {
          localStorage.removeItem(ACCESS_TOKEN_KEY);
          localStorage.removeItem(REFRESH_TOKEN_KEY);
        }
        return;
      }

      // Try to refresh the token and get user info
      const refreshed = await refreshAccessToken();

      if (refreshed) {
        try {
          const user = await usersApi.getMe();
          await applyUserSettings();
          authState.currentUser = user;
          void preloadAuthenticatedGraph();
        } catch {
          // If getting user fails, clear auth state
          clearAuthState();
        }
      }
    } catch (e) {
      console.error("Failed to initialize auth:", e);
      clearAuthState();
    } finally {
      authState.isInitialized = true;
      initAuthPromise = null;
    }
  })();

  return initAuthPromise;
}

/**
 * Login with credentials
 */
export async function login(login: string, password: string): Promise<boolean> {
  authState.isLoading = true;
  authState.error = null;

  if (skipAuthMode()) {
    saveTokens({
      access_token: "skip-auth-token",
      refresh_token: "skip-auth-refresh",
      token_type: "Bearer",
      expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    });
    authState.currentUser = {
      id: "skip-auth-user",
      login: login || "testuser",
      email: "test@example.com",
      role: "user",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    } as User;
    authState.isLoading = false;
    return true;
  }

  try {
    const tokens = await authApi.login(login, password);

    // Save tokens
    saveTokens(tokens);

    // Get user info
    const user = await usersApi.getMe();
    authState.currentUser = user;
    void preloadAuthenticatedGraph();
    void applyUserSettings();

    return true;
  } catch (e) {
    authState.error = e instanceof Error ? e.message : "Login failed";
    clearAuthState();
    return false;
  } finally {
    authState.isLoading = false;
  }
}

/**
 * Register new user
 */
export async function register(
  login: string,
  password: string,
  email?: string,
): Promise<boolean> {
  authState.isLoading = true;
  authState.error = null;

  try {
    const tokens = await authApi.register(login, password, email);

    // Save tokens
    saveTokens(tokens);

    // Get user info
    const user = await usersApi.getMe();
    authState.currentUser = user;
    void preloadAuthenticatedGraph();
    void applyUserSettings();

    return true;
  } catch (e) {
    authState.error = e instanceof Error ? e.message : "Registration failed";
    clearAuthState();
    return false;
  } finally {
    authState.isLoading = false;
  }
}

/**
 * Logout user
 */
export async function logout(): Promise<void> {
  if (authState.refreshToken) {
    try {
      await authApi.logout(authState.refreshToken);
    } catch (e) {
      // Ignore errors during logout
      console.error("Logout error:", e);
    }
  }

  // Clear preload cache on logout (never block logout if preload helpers fail)
  try {
    clearPreloadCache();
  } catch (e) {
    console.error("Failed to clear preload cache on logout:", e);
  }

  clearAuthState();
  goto("/auth/login");
}

/**
 * Refresh access token
 */
export async function refreshAccessToken(): Promise<boolean> {
  if (!authState.refreshToken) {
    return false;
  }

  try {
    const tokens = await authApi.refreshTokens(authState.refreshToken);
    saveTokens(tokens);
    return true;
  } catch (e) {
    console.error("Token refresh failed:", e);
    clearAuthState();
    return false;
  }
}

/**
 * Handle Yandex OAuth callback
 */
export async function handleYandexCallback(
  code: string,
  state: string,
): Promise<boolean> {
  authState.isLoading = true;
  authState.error = null;

  try {
    const tokens = await authApi.handleYandexCallback(code, state);

    // Save tokens
    saveTokens(tokens);

    // Get user info
    const user = await usersApi.getMe();
    authState.currentUser = user;
    void preloadAuthenticatedGraph();
    void applyUserSettings();

    return true;
  } catch (e) {
    authState.error =
      e instanceof Error ? e.message : "Yandex authentication failed";
    clearAuthState();
    return false;
  } finally {
    authState.isLoading = false;
  }
}

/**
 * Login with API Key
 */
export async function loginWithApiKey(key: string): Promise<boolean> {
  authState.isLoading = true;
  authState.error = null;

  try {
    // Save API key
    authState.apiKey = key;
    if (browser) {
      localStorage.setItem(API_KEY, key);
    }

    // Try to get user info with API key
    const user = await usersApi.getMe();
    authState.currentUser = user;
    void applyUserSettings();

    return true;
  } catch (e) {
    authState.error = e instanceof Error ? e.message : "Invalid API key";
    authState.apiKey = null;
    if (browser) {
      localStorage.removeItem(API_KEY);
    }
    return false;
  } finally {
    authState.isLoading = false;
  }
}

/**
 * Get current access token
 */
export function getAccessToken(): string | null {
  return authState.accessToken;
}

/**
 * Get API key
 */
export function getApiKey(): string | null {
  return authState.apiKey;
}

/**
 * Check if user is authenticated
 * SKIP_AUTH bypasses auth for testing via the window flag, localStorage, or query param.
 * The window flag is allowed in all builds for isolated visual regression tests.
 */
export function isAuthenticated(): boolean {
  if (browser) {
    // Check window flag injected by Playwright
    if ((window as any).__SKIP_AUTH__ === true) {
      return true;
    }
    // Check localStorage and query parameters only in dev to avoid production bypass
    if (import.meta.env.DEV) {
      if (localStorage.getItem("__SKIP_AUTH__") === "true") {
        return true;
      }
      const url = new URL(window.location.href);
      if (url.searchParams.get("skip_auth") === "true") {
        // Persist to localStorage for subsequent navigations
        localStorage.setItem("__SKIP_AUTH__", "true");
        return true;
      }
    }
  }
  return !!authState.accessToken || !!authState.apiKey;
}

/**
 * Check if user is admin
 */
export function isAdmin(): boolean {
  return authState.currentUser?.role === "admin";
}

/**
 * Save tokens to state and localStorage
 */
function saveTokens(tokens: AuthTokens): void {
  authState.accessToken = tokens.access_token;
  authState.refreshToken = tokens.refresh_token;

  if (browser) {
    localStorage.setItem(ACCESS_TOKEN_KEY, tokens.access_token);
    localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token);
  }
}

/**
 * Clear all auth state
 */
function clearAuthState(): void {
  authState.accessToken = null;
  authState.refreshToken = null;
  authState.currentUser = null;
  authState.apiKey = null;
  authState.error = null;

  if (browser) {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(API_KEY);
  }
}

/**
 * Update user info
 */
export async function updateUserInfo(): Promise<void> {
  if (!isAuthenticated()) {
    return;
  }

  try {
    const user = await usersApi.getMe();
    authState.currentUser = user;
    void applyUserSettings();
  } catch (e) {
    console.error("Failed to update user info:", e);
  }
}
