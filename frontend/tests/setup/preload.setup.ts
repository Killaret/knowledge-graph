// Setup file for PreloadService tests - DISABLED for Playwright
// import { vi, afterEach } from 'vitest'; // Temporarily disabled for Playwright

// Global mocks for all tests
/*
vi.mock('$app/environment', () => ({
  browser: true
}));

vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));
*/

// Mock for localStorage
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

// Cleanup after each test
/*
afterEach(() => {
  vi.clearAllMocks();
  localStorageMock.getItem.mockReturnValue(null);
});
*/
