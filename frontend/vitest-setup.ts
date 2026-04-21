import '@testing-library/jest-dom';
import { vi } from 'vitest';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';

// MSW server для мокирования HTTP запросов
export const server = setupServer();

// Запускаем сервер перед всеми тестами
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));

// Очищаем обработчики после каждого теста
afterEach(() => server.resetHandlers());

// Останавливаем сервер после всех тестов
afterAll(() => server.close());

// Mock SvelteKit modules
vi.mock('$app/environment', () => ({
	browser: true,
	dev: true,
	building: false,
	version: 'test'
}));

vi.mock('$app/navigation', () => ({
	goto: vi.fn(),
	beforeNavigate: vi.fn(),
	afterNavigate: vi.fn()
}));

vi.mock('$app/stores', () => ({
	page: {
		subscribe: vi.fn((fn) => {
			fn({ url: new URL('http://localhost'), params: {} });
			return () => {};
		})
	}
}));
