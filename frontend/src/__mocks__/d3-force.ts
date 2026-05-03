import { vi } from 'vitest';

// Хранилище состояния для мока - используем глобальный объект для shared state
const globalState = (globalThis as any).__D3_FORCE_MOCK_STATE__ || {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  stopCallback: null as (() => void) | null,
  currentSimulation: null as any,
};

// Сохраняем на globalThis
(globalThis as any).__D3_FORCE_MOCK_STATE__ = globalState;

// Экспортируем ссылку на глобальный state
export const mockState = globalState;

// Фабрика создания мок-симуляции
export function createMockSimulation() {
  const sim = {
    nodes: vi.fn().mockImplementation((nodes?: any[]) => {
      if (nodes) {
        // Immediately assign coordinates to nodes
        mockState.simulationNodes = nodes.map((n: any, i: number) => ({
          ...n,
          x: n.x ?? 400 + i * 50,
          y: n.y ?? 300 + i * 30,
          vx: 0,
          vy: 0
        }));
      }
      return mockState.simulationNodes;
    }),
    tick: vi.fn().mockImplementation((_iterations?: number) => {
      // Immediately update node positions
      mockState.simulationNodes = mockState.simulationNodes.map((n: any, _i: number) => ({
        ...n,
        x: (n.x || 400) + (Math.random() - 0.5) * 10,
        y: (n.y || 300) + (Math.random() - 0.5) * 10
      }));
      // Trigger tick callback synchronously
      if (mockState.tickCallback) {
        mockState.tickCallback();
      }
      return sim;
    }),
    force: vi.fn().mockReturnThis(),
    alphaDecay: vi.fn().mockReturnThis(),
    velocityDecay: vi.fn().mockReturnThis(),
    on: vi.fn().mockImplementation(function(this: any, event: string, callback: () => void) {
      if (event === 'tick') {
        mockState.tickCallback = callback;
      }
      return this;
    }),
    alpha: vi.fn().mockImplementation(function(this: any, value?: number) {
      if (value !== undefined) {
        return this;
      }
      return 1;
    }),
    restart: vi.fn().mockImplementation(function(this: any) {
      if (mockState.tickCallback) {
        mockState.tickCallback();
      }
      return this;
    }),
    stop: vi.fn().mockImplementation(function(this: any) {
      mockState.stopCallback?.();
      mockState.tickCallback = null;
      return this;
    })
  };
  return sim;
}

// Экспорты модуля d3-force
export const forceSimulation = vi.fn().mockImplementation((nodes?: any[]) => {
  const sim = createMockSimulation();
  if (nodes) {
    sim.nodes(nodes);
    // Immediately trigger tick to assign coordinates
    queueMicrotask(() => {
      if (mockState.tickCallback) {
        mockState.tickCallback();
      }
    });
  }
  return sim;
});

export const forceLink = vi.fn().mockImplementation((links?: any[]) => {
  if (links) {
    mockState.simulationLinks = links.map((l: any) => ({
      ...l,
      source: typeof l.source === 'string' ? { id: l.source } : l.source,
      target: typeof l.target === 'string' ? { id: l.target } : l.target
    }));
  }
  const linkForce: any = {
    id: vi.fn().mockReturnThis(),
    distance: vi.fn().mockReturnThis(),
    strength: vi.fn().mockReturnThis(),
    links: () => mockState.simulationLinks
  };
  return linkForce;
});

export const forceManyBody = vi.fn().mockReturnValue({
  strength: vi.fn().mockReturnThis()
});

export const forceCenter = vi.fn().mockImplementation(() => ({
  strength: vi.fn().mockReturnThis()
}));

export const forceCollide = vi.fn().mockReturnValue({
  radius: vi.fn().mockReturnThis()
});

// Helper to reset state between tests
export function resetMockState() {
  mockState.simulationNodes = [];
  mockState.simulationLinks = [];
  mockState.tickCallback = null;
  mockState.stopCallback = null;
}
