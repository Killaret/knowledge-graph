import '@testing-library/jest-dom/vitest';
import { vi, beforeAll, afterEach, afterAll } from 'vitest';
import { setupServer } from 'msw/node';
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

// Note: $app/* modules are mocked via resolve aliases in vitest.config.ts
// pointing to src/lib/mocks/app/*.ts files

// d3-force is mocked via src/__mocks__/d3-force.ts (auto-loaded by Vitest)

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

// Мокируем Web Animations API (Element.prototype.animate) СРАЗУ при импорте
// Это необходимо для Svelte transitions в jsdom
if (typeof Element !== 'undefined') {
	Element.prototype.animate = vi.fn().mockImplementation(() => {
		const anim = {
			finished: Promise.resolve(),
			cancel: vi.fn(),
			pause: vi.fn(),
			play: vi.fn()
		};
		// Define onfinish with setter that auto-triggers
		Object.defineProperty(anim, 'onfinish', {
			set(cb: () => void) {
				if (cb) queueMicrotask(() => cb());
			},
			get() { return null; },
			configurable: true
		});
		return anim;
	});
}

// Мокируем requestAnimationFrame/cancelAnimationFrame
global.requestAnimationFrame = vi.fn((callback: FrameRequestCallback) => {
	return setTimeout(() => callback(performance.now()), 16) as unknown as number;
});
global.cancelAnimationFrame = vi.fn((id: number) => {
	clearTimeout(id);
});

// Мокируем CSS2DRenderer и CSS2DObject из three/examples/jsm
vi.mock('three/examples/jsm/renderers/CSS2DRenderer.js', () => ({
	CSS2DRenderer: vi.fn().mockImplementation(() => ({
		setSize: vi.fn(),
		render: vi.fn(),
		domElement: document.createElement('div')
	})),
	CSS2DObject: vi.fn().mockImplementation((element: HTMLElement) => ({
		element,
		position: { set: vi.fn() }
	}))
}));
