// API client wrapper для ky
import ky from 'ky';
import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { getAccessToken, getApiKey, refreshAccessToken } from '$shared/stores/auth.svelte';

// Redirect to login after an unrecoverable auth failure (browser only)
function redirectToLogin(): void {
  if (browser) {
    void goto('/auth/login');
  }
}

// Проверяем тестовое окружение через process.env.VITEST
// Vitest устанавливает эту переменную автоматически
const isTest = typeof process !== 'undefined' && process.env?.VITEST === 'true';

// Получаем backend URL из env (для Docker) или используем default
let backendUrl = '';
try {
  const envUrl = (import.meta as any).env?.VITE_API_URL;
  if (envUrl) backendUrl = envUrl;
} catch {
  // Fallback to default
}

// В dev режиме (Vite) используем относительный путь (проксируется на backend)
// Production использует прямой backend URL или относительный путь
// Если backendUrl относительный (начинается с /), используем его как есть
const isRelativeUrl = backendUrl.startsWith('/');
const prefixUrl = isTest && !isRelativeUrl
  ? `${backendUrl}/api`
  : (isRelativeUrl ? backendUrl : `${backendUrl}/api`);

// Flag to prevent infinite refresh loops
let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

// Базовый URL с прокси /api → http://localhost:8080
// Retry настроен для устойчивости к временным сетевым сбоям:
// - limit: 3 попытки (4 total: initial + 3 retries)
// - methods: все HTTP методы кроме DELETE (DELETE не идемпотентен)
// - statusCodes: сетевые и серверные ошибки, rate limiting
// Ky использует встроенный exponential backoff с jitter между попытками
export const api = ky.create({
  prefixUrl,
  timeout: 30000,
  retry: {
    limit: 3,
    methods: ['get', 'post', 'put', 'patch'],
    statusCodes: [408, 413, 429, 500, 502, 503, 504]
  },
  hooks: {
    beforeRequest: [
      async (request) => {
        // Add Authorization header if access token exists
        const token = getAccessToken();
        if (token) {
          request.headers.set('Authorization', `Bearer ${token}`);
        }
        
        // Add API Key header if exists (for API key auth)
        const key = getApiKey();
        if (key) {
          request.headers.set('X-API-Key', key);
        }
      }
    ],
    afterResponse: [
      async (request, options, response) => {
        // The auth refresh endpoint is used by the refresh flow itself;
        // do not try to refresh on a refresh request to avoid recursion/deadlock.
        const requestPath = new URL(request.url).pathname;
        if (requestPath.endsWith('/api/v1/auth/refresh')) {
          return response;
        }

        // Handle 401 Unauthorized
        if (response.status === 401) {
          // Prevent infinite loops - if we're already refreshing, wait for it and retry
          if (isRefreshing) {
            // Wait for the current refresh to complete
            if (refreshPromise) {
              const refreshed = await refreshPromise;
              if (refreshed) {
                // Retry the original request with the new token
                const newToken = getAccessToken();
                if (newToken) {
                  request.headers.set('Authorization', `Bearer ${newToken}`);
                  return ky(request);
                }
              }
              // Refresh failed while waiting — redirect to login
              redirectToLogin();
            }
            return response;
          }

          // Try to refresh the token
          isRefreshing = true;
          refreshPromise = refreshAccessToken();

          try {
            const refreshed = await refreshPromise;
            if (refreshed) {
              const newToken = getAccessToken();
              if (newToken) {
                request.headers.set('Authorization', `Bearer ${newToken}`);
                return ky(request);
              }
            }
            // Refresh failed (or no new token) — redirect to login
            redirectToLogin();
          } finally {
            isRefreshing = false;
            refreshPromise = null;
          }
        }

        return response;
      }
    ]
  }
});

export default api;
