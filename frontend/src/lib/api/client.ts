// API client wrapper для ky
import ky from 'ky';

// Определяем базовый URL в зависимости от среды
// import.meta.env доступен в Vite/SvelteKit
const isDev = import.meta.env.DEV;
const isProd = import.meta.env.PROD;

// Проверяем тестовое окружение через process.env.VITEST
// Vitest устанавливает эту переменную автоматически
const isTest = typeof process !== 'undefined' && process.env?.VITEST === 'true';

// В dev режиме (Vite) используем относительный /api (проксируется на backend)
// В production и тестах используем полный URL к backend
const prefixUrl = isDev && !isTest
  ? '/api' 
  : 'http://localhost:8080/api';

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
  }
});

export default api;
