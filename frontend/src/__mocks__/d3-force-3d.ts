import { vi } from "vitest";

const globalState = (globalThis as any).__D3_FORCE_3D_MOCK_STATE__ || {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  stopCallback: null as (() => void) | null,
  currentSimulation: null as any,
};

(globalThis as any).__D3_FORCE_3D_MOCK_STATE__ = globalState;

export const mockState = globalState;

function createMockSimulation() {
  let alphaValue = 1;
  let alphaMinValue = 0.001;

  const sim = {
    nodes: vi.fn().mockImplementation((nodes?: any[]) => {
      if (nodes) {
        nodes.forEach((n: any, i: number) => {
          n.x = n.x ?? 400 + i * 50;
          n.y = n.y ?? 300 + i * 30;
          n.z = n.z ?? (Math.random() - 0.5) * 100;
          n.vx = 0;
          n.vy = 0;
          n.vz = 0;
        });
        mockState.simulationNodes = nodes;
      }
      return mockState.simulationNodes;
    }),
    tick: vi.fn().mockImplementation((iterations?: number) => {
      const count = iterations ?? 1;
      for (let k = 0; k < count; k++) {
        alphaValue *= 0.98;
        for (const n of mockState.simulationNodes) {
          n.x = (n.x || 400) + (Math.random() - 0.5) * 10;
          n.y = (n.y || 300) + (Math.random() - 0.5) * 10;
          n.z = (n.z || 0) + (Math.random() - 0.5) * 10;
        }
      }
      if (mockState.tickCallback) {
        mockState.tickCallback();
      }
      return sim;
    }),
    force: vi.fn().mockImplementation((name: string, _force?: any) => {
      if (arguments.length > 1) return sim;
      if (name === "link") {
        return {
          id: vi.fn().mockReturnThis(),
          distance: vi.fn().mockReturnThis(),
          strength: vi.fn().mockReturnThis(),
          links: () => mockState.simulationLinks,
        };
      }
      return undefined;
    }),
    alphaDecay: vi.fn().mockImplementation(function (this: any, value?: number) {
      if (value !== undefined) {
        alphaValue *= 0.98;
      }
      return this;
    }),
    alpha: vi.fn().mockImplementation((_value?: number) => {
      if (_value !== undefined) {
        alphaValue = _value;
        return sim;
      }
      return alphaValue;
    }),
    alphaMin: vi.fn().mockImplementation(function (this: any, value?: number) {
      if (value !== undefined) {
        alphaMinValue = value;
        return this;
      }
      return alphaMinValue;
    }),
    velocityDecay: vi.fn().mockReturnThis(),
    alphaTarget: vi.fn().mockReturnThis(),
    on: vi.fn().mockImplementation(function (this: any, event: string, callback: () => void) {
      if (event === "tick") {
        mockState.tickCallback = callback;
      }
      return this;
    }),
    restart: vi.fn().mockImplementation(function (this: any) {
      alphaValue = 1;
      if (mockState.tickCallback) {
        mockState.tickCallback();
      }
      return this;
    }),
    stop: vi.fn().mockImplementation(function (this: any) {
      mockState.stopCallback?.();
      mockState.tickCallback = null;
      alphaValue = alphaMinValue / 2;
      return this;
    }),
  };
  return sim;
}

export const forceSimulation = vi.fn().mockImplementation((nodes?: any[]) => {
  const sim = createMockSimulation();
  if (nodes) {
    sim.nodes(nodes);
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
      source: typeof l.source === "string" ? { id: l.source } : l.source,
      target: typeof l.target === "string" ? { id: l.target } : l.target,
    }));
  }
  const linkForce: any = {
    id: vi.fn().mockReturnThis(),
    distance: vi.fn().mockReturnThis(),
    strength: vi.fn().mockReturnThis(),
    links: () => mockState.simulationLinks,
  };
  return linkForce;
});

export const forceManyBody = vi.fn().mockReturnValue({
  strength: vi.fn().mockReturnThis(),
  distanceMin: vi.fn().mockReturnThis(),
  distanceMax: vi.fn().mockReturnThis(),
});

export const forceCenter = vi.fn().mockImplementation(() => ({
  strength: vi.fn().mockReturnThis(),
  x: vi.fn().mockReturnThis(),
  y: vi.fn().mockReturnThis(),
  z: vi.fn().mockReturnThis(),
}));

export function resetMockState() {
  mockState.simulationNodes = [];
  mockState.simulationLinks = [];
  mockState.tickCallback = null;
  mockState.stopCallback = null;
}
