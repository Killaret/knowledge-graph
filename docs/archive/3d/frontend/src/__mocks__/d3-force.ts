import { vi } from 'vitest';

// Хранилище состояния для мока - используем глобальный объект для shared state
const globalState = (globalThis as any).__D3_FORCE_MOCK_STATE__ || {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  stopCallback: null as (() => void) | null,
  currentSimulation: null as any,
  createSimulation: null as any,
};

// Сохраняем на globalThis
(globalThis as any).__D3_FORCE_MOCK_STATE__ = globalState;

// Экспортируем ссылку на глобальный state
export const mockState = globalState;

// Инициализируем createSimulation после объявления mockState
mockState.createSimulation = () => createMockSimulation();

// Фабрика создания мок-симуляции
export function createMockSimulation() {
  const sim = {
    nodes: vi.fn().mockImplementation((nodes?: any[]) => {
      if (nodes) {
        mockState.simulationNodes = nodes.map((n: any, i: number) => ({
          ...n,
          x: 400 + i * 50,
          y: 300 + i * 30
        }));
      }
      return mockState.simulationNodes;
    }),
    tick: vi.fn().mockImplementation(() => {
      mockState.simulationNodes = mockState.simulationNodes.map((n: any, i: number) => ({
        ...n,
        x: (n.x || 400) + i * 2,
        y: (n.y || 300) + i * 2
      }));
      return sim;
    }),
    force: vi.fn().mockReturnThis(),
    alphaDecay: vi.fn().mockReturnThis(),
    on: vi.fn().mockImplementation(function(this: any, event: string, callback: () => void) {
      if (event === 'tick') {
        mockState.tickCallback = callback;
      }
      return this;
    }),
    alpha: vi.fn().mockReturnThis(),
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
  console.log('[MOCK] forceSimulation called with nodes:', nodes?.length);
  const sim = createMockSimulation();
  if (nodes) {
    sim.nodes(nodes);
    console.log('[MOCK] simulationNodes after init:', mockState.simulationNodes.length);
  }
  return sim;
});

export const forceLink = vi.fn().mockImplementation((links?: any[]) => {
  if (links) {
    mockState.simulationLinks = links;
  }
  const linkForce: any = {
    id: (fn?: (d: any) => string) => {
      if (fn) return linkForce;
      return linkForce;
    },
    distance: () => linkForce,
    strength: () => linkForce,
    links: () => mockState.simulationLinks
  };
  return linkForce;
});

export const forceManyBody = vi.fn().mockReturnValue({
  strength: vi.fn().mockReturnThis()
});

export const forceCenter = vi.fn().mockImplementation(() => ({
  strength: () => ({}) // Returns object with strength method
}));

export const forceCollide = vi.fn().mockReturnValue({
  radius: vi.fn().mockReturnThis()
});
