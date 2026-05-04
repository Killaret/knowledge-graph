// Auth store with Svelte 5 runes
import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import * as authApi from '$lib/api/auth';
import * as usersApi from '$lib/api/users';
import type { User, AuthTokens } from '$lib/types';

// Global reactive state - wrapped in object for export
const authState = $state({
  currentUser: null as User | null,
  accessToken: null as string | null,
  refreshToken: null as string | null,
  isInitialized: false,
  isLoading: false,
  error: null as string | null,
  apiKey: null as string | null
});

// Export reactive state through getter functions
export function currentUser(): User | null { return authState.currentUser; }
export function accessToken(): string | null { return authState.accessToken; }
export function refreshToken(): string | null { return authState.refreshToken; }
export function isInitialized(): boolean { return authState.isInitialized; }
export function isLoading(): boolean { return authState.isLoading; }
export function error(): string | null { return authState.error; }
export function apiKey(): string | null { return authState.apiKey; }

// LocalStorage keys
const ACCESS_TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';
const API_KEY = 'api_key';

/**
 * Initialize auth state from localStorage
 */
export async function initAuth(): Promise<void> {
  if (!browser) {
    authState.isInitialized = true;
    return;
  }

  try {
    // Try to load tokens from localStorage
    const storedAccessToken = localStorage.getItem(ACCESS_TOKEN_KEY);
    const storedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
    const storedApiKey = localStorage.getItem(API_KEY);

    if (storedApiKey) {
      authState.apiKey = storedApiKey;
    }

    if (storedAccessToken) {
      authState.accessToken = storedAccessToken;
    }

    if (storedRefreshToken) {
      authState.refreshToken = storedRefreshToken;
      
      // Try to refresh the token and get user info
      const refreshed = await refreshAccessToken();
      
      if (refreshed) {
        try {
          const user = await usersApi.getMe();
          authState.currentUser = user;
        } catch (e) {
          // If getting user fails, clear auth state
          clearAuthState();
        }
      }
    }
  } catch (e) {
    console.error('Failed to initialize auth:', e);
    clearAuthState();
  } finally {
    authState.isInitialized = true;
  }
}

/**
 * Login with credentials
 */
export async function login(login: string, password: string): Promise<boolean> {
  authState.isLoading = true;
  authState.error = null;

  try {
    const tokens = await authApi.login(login, password);
    
    // Save tokens
    saveTokens(tokens);
    
    // Get user info
    const user = await usersApi.getMe();
    authState.currentUser = user;
    
    return true;
  } catch (e) {
    authState.error = e instanceof Error ? e.message : 'Login failed';
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
    
    return true;
  } catch (e) {
    authState.error = e instanceof Error ? e.message : 'Registration failed';
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
      console.error('Logout error:', e);
    }
  }
  
  clearAuthState();
  goto('/auth/login');
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
    console.error('Token refresh failed:', e);
    clearAuthState();
    return false;
  }
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
    
    return true;
  } catch (e) {
    authState.error = e instanceof Error ? e.message : 'Yandex authentication failed';
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
    
    return true;
  } catch (e) {
    authState.error = e instanceof Error ? e.message : 'Invalid API key';
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
 */
export function isAuthenticated(): boolean {
  return !!authState.accessToken || !!authState.apiKey;
}

/**
 * Check if user is admin
 */
export function isAdmin(): boolean {
  return authState.currentUser?.role === 'admin';
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
  } catch (e) {
    console.error('Failed to update user info:', e);
  }
}
