// API client wrapper для ky
import ky from 'ky';

// Определяем базовый URL в зависимости от среды
const isTest = typeof process !== 'undefined' && process.env.NODE_ENV === 'test';
const prefixUrl = isTest ? 'http://localhost:8081/api' : '/api';

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
