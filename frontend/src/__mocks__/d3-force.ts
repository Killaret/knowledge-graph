import { vi } from 'vitest';

// Хранилище состояния для мока
export const mockState = {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  stopCallback: null as (() => void) | null,
  currentSimulation: null as any,
  createSimulation: null as any,
};

// Инициализируем createSimulation после объявления mockState
mockState.createSimulation = () => createMockSimulation();

// Фабрика создания мок-симуляции
export function createMockSimulation() {
  return {
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
      setTimeout(() => {
        if (mockState.tickCallback) {
          mockState.simulationNodes = mockState.simulationNodes.map((n: any, i: number) => ({
            ...n,
            x: n.x + 10 + i * 5,
            y: n.y + 15 + i * 3
          }));
          mockState.tickCallback();
        }
      }, 10);
      return this;
    }),
    stop: vi.fn().mockImplementation(function(this: any) {
      mockState.stopCallback?.();
      mockState.tickCallback = null;
      return this;
    })
  };
}

// Экспорты модуля d3-force
export const forceSimulation = vi.fn().mockImplementation((nodes?: any[]) => {
  const sim = createMockSimulation();
  if (nodes) {
    sim.nodes(nodes);
  }
  return sim;
});

export const forceLink = vi.fn().mockImplementation((links?: any[]) => {
  if (links) {
    mockState.simulationLinks = links;
  }
  return {
    id: vi.fn().mockReturnThis(),
    distance: vi.fn().mockReturnThis(),
    strength: vi.fn().mockReturnThis()
  };
});

export const forceManyBody = vi.fn().mockReturnValue({
  strength: vi.fn().mockReturnThis()
});

export const forceCenter = vi.fn().mockReturnThis();

export const forceCollide = vi.fn().mockReturnValue({
  radius: vi.fn().mockReturnThis()
});
