// Auth store with Svelte 5 runes. This module orchestrates authentication flows;
// low-level state lives in auth-session.svelte.ts to keep the API client cycle-free.
import { browser } from "$app/environment";
import { goto } from "$app/navigation";
import * as authApi from "$shared/api/auth";
import * as usersApi from "$shared/api/users";
import { clearPreloadCache, preloadAuthenticatedGraph } from "$shared/services/PreloadService";
import { setLocale, type Locale } from "$shared/utils/i18n";
import type { User } from "$shared/types";
import {
  authState,
  clearAuthState,
  isAuthenticated,
  saveTokens,
  setApiKey,
  setSessionHint,
  hasSessionHint,
  getApiKey,
  skipAuthMode,
} from "$shared/stores/auth-session.svelte";

// Re-export state getters so components can keep importing from this module
export {
  currentUser,
  accessToken,
  refreshToken,
  isInitialized,
  isLoading,
  error,
  apiKey,
  getApiKey,
  isAuthenticated,
  isAdmin,
  skipAuthMode,
} from "$shared/stores/auth-session.svelte";

// Guard concurrent initAuth calls so refresh token is not used twice
let initAuthPromise: Promise<void> | null = null;

const API_KEY = "api_key";

/**
 * Apply user settings like locale after login/init.
 */
async function applyUserSettings(): Promise<void> {
  try {
    const { settings } = (await usersApi.getSettings()) ?? { settings: [] };
    const localeSetting = settings?.find((s) => s.key === "preferred_language");
    if (localeSetting?.value && (localeSetting.value === "en" || localeSetting.value === "ru")) {
      setLocale(localeSetting.value as Locale);
    }
  } catch (e) {
    // Settings are not critical for auth flow
    if (import.meta.env.DEV) {
      console.warn("Failed to load user settings:", e);
    }
  }
}

/**
 * Initialize auth state from the session.
 * Access/refresh tokens are now HttpOnly cookies and are not accessible to JS.
 */
export async function initAuth(): Promise<void> {
  if (!browser) {
    authState.isInitialized = true;
    return;
  }

  if (initAuthPromise) {
    return initAuthPromise;
  }

  // Auth init only needs to run once per app session. Subsequent calls from
  // other pages/layouts can rely on the existing state (login sets it directly).
  if (authState.isInitialized && authState.currentUser) {
    return;
  }

  initAuthPromise = (async () => {
    // Capture whether we already had a token at the start. This is used to
    // avoid wiping a token set by a concurrent login while initAuth is still
    // running.
    let hadTokenAtStart = false;

    try {
      // Restore API key (user-provided, not a JWT)
      const storedApiKey = browser ? localStorage.getItem(API_KEY) : null;
      setApiKey(storedApiKey);

      // Check for SKIP_AUTH mode from query parameter on first load (dev only)
      if (import.meta.env.DEV) {
        const url = new URL(window.location.href);
        if (url.searchParams.get("skip_auth") === "true") {
          // Persist SKIP_AUTH to localStorage
          localStorage.setItem("__SKIP_AUTH__", "true");
          console.log("[Auth] SKIP_AUTH mode enabled via query param");
        }
      }

      // In SKIP_AUTH mode we do not need real tokens or refresh flow.
      if (skipAuthMode()) {
        saveTokens({
          access_token: "skip-auth-token",
          refresh_token: "skip-auth-refresh",
          token_type: "Bearer",
          expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
        });
        await applyUserSettings();
        authState.currentUser = {
          id: "skip-auth-user",
          login: "testuser",
          email: "test@example.com",
          role: "user",
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        } as User;
        void preloadAuthenticatedGraph();
        return;
      }

      // Allow test harnesses to inject an access token via window.__ACCESS_TOKEN__
      // (e.g. Playwright/Cucumber real-auth scenarios). This avoids the need to
      // share HttpOnly refresh cookies across origins.
      if (browser && !(window as any).__SKIP_AUTH__) {
        const injectedToken = (window as any).__ACCESS_TOKEN__ as string | undefined;
        if (injectedToken && !authState.accessToken) {
          saveTokens({
            access_token: injectedToken,
            refresh_token: "",
            token_type: "Bearer",
            expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
          });
        }
      }

      // If we already have an access token, try to use it directly. This
      // handles the common case where the user just logged in and the
      // HttpOnly refresh cookie may not be available (cross-origin / test env).
      hadTokenAtStart = !!authState.accessToken;
      const hadApiKey = !!getApiKey();
      const hadSessionHint = hasSessionHint();

      // If there is no trace of a previous session (no in-memory token, no API
      // key, no session hint), the browser is anonymous. Avoid hitting
      // /v1/auth/refresh so public pages do not log a 401 on every load.
      if (!hadTokenAtStart && !hadApiKey && !hadSessionHint) {
        authState.isInitialized = true;
        return;
      }

      // If the user has an API key, try to validate it directly. API keys do
      // not support cookie refresh, so do not fall through to the refresh flow.
      if (hadApiKey) {
        try {
          const user = await usersApi.getMe();
          await applyUserSettings();
          authState.currentUser = user;
          void preloadAuthenticatedGraph();
        } catch {
          // API key is invalid or the account no longer exists.
          clearAuthState();
        }
        return;
      }

      // If we have an in-memory token, try it first. If the request fails, the
      // API client's afterResponse hook will attempt a cookie refresh; we do not
      // need to fall through to an explicit refresh here.
      if (hadTokenAtStart) {
        try {
          const user = await usersApi.getMe();
          await applyUserSettings();
          authState.currentUser = user;
          void preloadAuthenticatedGraph();
          return;
        } catch {
          // existing token didn't work and the client already tried to refresh
          clearAuthState();
        }
        return;
      }

      // If we reach here, we have a session hint but no in-memory token or API
      // key. The user may still have a valid HttpOnly refresh cookie, so try to
      // use it to obtain a fresh access token.
      const tokens = await authApi.refreshTokens();
      saveTokens(tokens);

      const user = await usersApi.getMe();
      await applyUserSettings();
      authState.currentUser = user;
      void preloadAuthenticatedGraph();
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to initialize auth:", e);
      }
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
export async function register(login: string, password: string, email?: string): Promise<boolean> {
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
  try {
    await authApi.logout();
  } catch (e) {
    // Ignore errors during logout
    if (import.meta.env.DEV) {
      console.error("Logout error:", e);
    }
  }

  // Clear preload cache on logout (never block logout if preload helpers fail)
  try {
    clearPreloadCache();
  } catch (e) {
    if (import.meta.env.DEV) {
      console.error("Failed to clear preload cache on logout:", e);
    }
  }

  clearAuthState();
  goto("/auth/login");
}

/**
 * Handle Yandex OAuth callback
 */
export async function handleYandexCallback(code: string, state: string): Promise<boolean> {
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
    authState.error = e instanceof Error ? e.message : "Yandex authentication failed";
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
    setApiKey(key);

    // Try to get user info with API key
    const user = await usersApi.getMe();
    authState.currentUser = user;
    void applyUserSettings();

    return true;
  } catch (e) {
    authState.error = e instanceof Error ? e.message : "Invalid API key";
    setApiKey(null);
    return false;
  } finally {
    authState.isLoading = false;
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
    if (import.meta.env.DEV) {
      console.error("Failed to update user info:", e);
    }
  }
}
