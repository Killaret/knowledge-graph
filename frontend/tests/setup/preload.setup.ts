// Setup файл для тестов PreloadService - ОТКЛЮЧЕН для Playwright
// import { vi, afterEach } from 'vitest'; // Временно отключено для Playwright

// Глобальные моки для всех тестов
/*
vi.mock('$app/environment', () => ({
  browser: true
}));

vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));
*/

// Мок для localStorage
/*
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
};

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
});

// Очистка после каждого теста
/*
afterEach(() => {
  vi.clearAllMocks();
  localStorageMock.getItem.mockReturnValue(null);
});
*/
