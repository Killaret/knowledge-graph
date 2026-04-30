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

// Мокируем requestAnimationFrame/cancelAnimationFrame - СИНХРОННО для стабильности тестов
const rafCallbacks = new Map<number, FrameRequestCallback>();
let rafId = 0;

global.requestAnimationFrame = vi.fn((callback: FrameRequestCallback) => {
	rafId++;
	rafCallbacks.set(rafId, callback);
	// Вызываем синхронно, но не рекурсивно
	queueMicrotask(() => {
		if (rafCallbacks.has(rafId)) {
			callback(performance.now());
			rafCallbacks.delete(rafId);
		}
	});
	return rafId;
});

global.cancelAnimationFrame = vi.fn((id: number) => {
	rafCallbacks.delete(id);
});

// Мок ResizeObserver
global.ResizeObserver = class ResizeObserver {
	observe() {}
	unobserve() {}
	disconnect() {}
};

// Мок CanvasRenderingContext2D
const mockCanvasContext = {
	beginPath: vi.fn(),
	arc: vi.fn(),
	fill: vi.fn(),
	stroke: vi.fn(),
	moveTo: vi.fn(),
	lineTo: vi.fn(),
	clearRect: vi.fn(),
	save: vi.fn(),
	restore: vi.fn(),
	translate: vi.fn(),
	scale: vi.fn(),
	rotate: vi.fn(),
	setLineDash: vi.fn(),
	fillText: vi.fn(),
	strokeText: vi.fn(),
	measureText: vi.fn().mockReturnValue({ width: 100 }),
	closePath: vi.fn(),
	rect: vi.fn(),
	fillRect: vi.fn(),
	strokeRect: vi.fn(),
	clip: vi.fn(),
	createLinearGradient: vi.fn().mockReturnValue({
		addColorStop: vi.fn()
	}),
	createRadialGradient: vi.fn().mockReturnValue({
		addColorStop: vi.fn()
	}),
	getImageData: vi.fn().mockReturnValue({ data: new Uint8ClampedArray(4) }),
	putImageData: vi.fn(),
	drawImage: vi.fn(),
	fillStyle: '',
	strokeStyle: '',
	lineWidth: 1,
	lineDashOffset: 0,
	globalAlpha: 1,
	font: '10px sans-serif',
	textAlign: 'left',
	textBaseline: 'alphabetic',
	shadowColor: '',
	shadowBlur: 0,
	shadowOffsetX: 0,
	shadowOffsetY: 0
};

// Мок getContext для canvas
Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
	value: vi.fn((contextId: string) => {
		if (contextId === '2d') {
			return { ...mockCanvasContext };
		}
		return null;
	}),
	configurable: true
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

// Мокируем GraphCanvas animation модуль
vi.mock('$lib/components/GraphCanvas/animation.ts', () => ({
	startAnimationLoop: vi.fn().mockReturnValue({
		stop: vi.fn()
	}),
	updateNodeAngles: vi.fn(),
	clearAnimationState: vi.fn()
}));
