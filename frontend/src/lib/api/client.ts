// API client wrapper для ky
import ky from 'ky';

// Определяем базовый URL в зависимости от среды
const isTest = typeof process !== 'undefined' && process.env.NODE_ENV === 'test';
const prefixUrl = isTest ? 'http://localhost/api' : '/api';

// Базовый URL с прокси /api → http://localhost:8080
export const api = ky.create({ 
  prefixUrl,
  timeout: 30000,
  retry: {
    limit: 2,
    methods: ['get', 'post', 'put', 'delete'],
    statusCodes: [408, 413, 429, 500, 502, 503, 504]
  }
});

export default api;
