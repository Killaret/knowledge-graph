import '@testing-library/jest-dom/vitest';
import { vi, beforeAll, afterEach, afterAll } from 'vitest';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';
import { cleanup } from '@testing-library/svelte';

// MSW server для мокирования HTTP запросов
export const server = setupServer();

// Запускаем сервер перед всеми тестами
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));

// Очищаем обработчики после каждого теста
afterEach(() => {
	cleanup();
	server.resetHandlers();
});

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
	},
	navigating: {
		subscribe: vi.fn()
	}
}));

// Подавляем ошибки Three.js в тестовом окружении
vi.mock('three', async () => {
	const actual = await vi.importActual<typeof import('three')>('three');
	return {
		...actual,
		WebGLRenderer: vi.fn().mockImplementation(() => ({
			render: vi.fn(),
			setSize: vi.fn(),
			domElement: document.createElement('canvas'),
		}))
	};
});
