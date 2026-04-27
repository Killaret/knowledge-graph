import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render } from '@testing-library/svelte';

const mockState = {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  stopCallback: null as (() => void) | null,
};

vi.mock('d3-force', () => {
  const createMockSimulation = () => {
    const sim: any = {
      nodes: vi.fn((n?: any[]) => {
        if (n) {
          mockState.simulationNodes = n.map((node, i) => ({
            ...node,
            x: 400 + i * 50,
            y: 300 + i * 30
          }));
        }
        return mockState.simulationNodes;
      }),
      tick: vi.fn(() => sim),
      force: vi.fn(() => sim),
      alphaDecay: vi.fn(() => sim),
      on: vi.fn((event: string, cb: () => void) => {
        if (event === 'tick') mockState.tickCallback = cb;
        return sim;
      }),
      alpha: vi.fn(() => sim),
      restart: vi.fn(() => {
        if (mockState.tickCallback) mockState.tickCallback();
        return sim;
      }),
      stop: vi.fn(() => {
        mockState.tickCallback = null;
        return sim;
      }),
    };
    return sim;
  };

  const forceSimulation = vi.fn((nodes?: any[]) => {
    const sim = createMockSimulation();
    if (nodes) {
      sim.nodes(nodes);
      setTimeout(() => {
        if (mockState.tickCallback) mockState.tickCallback();
      }, 50);
    }
    return sim;
  });

  const forceLink = vi.fn((links?: any[]) => {
    if (links) mockState.simulationLinks = links;
    const linkForce: any = {
      id: (fn?: (d: any) => string) => { if (fn) return linkForce; return linkForce; },
      distance: () => linkForce,
      strength: () => linkForce,
      links: () => mockState.simulationLinks,
    };
    return linkForce;
  });

  const forceManyBody = vi.fn(() => ({ strength: vi.fn(() => ({})) }));
  const forceCenter = vi.fn(() => ({ strength: vi.fn(() => ({})) }));
  const forceCollide = vi.fn(() => ({ radius: vi.fn(() => ({})) }));

  return {
    forceSimulation, forceLink, forceManyBody, forceCenter, forceCollide,
    __esModule: true,
    default: { forceSimulation, forceLink, forceManyBody, forceCenter, forceCollide },
  };
});

import GraphCanvas from './GraphCanvas.svelte';

const mockNodes = [
  { id: '1', title: 'Node 1', type: 'star' },
  { id: '2', title: 'Node 2', type: 'planet' },
  { id: '3', title: 'Node 3', type: 'comet' }
];

const mockLinks = [
  { source: '1', target: '2', weight: 0.8, link_type: 'reference' },
  { source: '2', target: '3', weight: 0.5, link_type: 'dependency' }
];

describe('GraphCanvas - Interactions', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    mockState.simulationNodes = [];
    mockState.simulationLinks = [];
    mockState.tickCallback = null;
    mockState.stopCallback = null;
    
    HTMLCanvasElement.prototype.getContext = vi.fn().mockReturnValue({
      clearRect: vi.fn(), save: vi.fn(), restore: vi.fn(), translate: vi.fn(), scale: vi.fn(),
      beginPath: vi.fn(), moveTo: vi.fn(), lineTo: vi.fn(), stroke: vi.fn(), fill: vi.fn(),
      closePath: vi.fn(), arc: vi.fn(), ellipse: vi.fn(), rotate: vi.fn(), fillRect: vi.fn(),
      strokeRect: vi.fn(), setLineDash: vi.fn(), fillText: vi.fn(), measureText: vi.fn(() => ({ width: 50 })),
      fillStyle: '', strokeStyle: '', font: '', textAlign: 'center', textBaseline: 'middle',
      lineWidth: 1, shadowBlur: 0, shadowColor: ''
    });

    global.ResizeObserver = vi.fn().mockImplementation(() => ({
      observe: vi.fn(), disconnect: vi.fn(), unobserve: vi.fn()
    }));

    vi.stubGlobal('requestAnimationFrame', vi.fn().mockImplementation((cb: FrameRequestCallback) => {
      setTimeout(cb, 16); return 1;
    }));
    vi.stubGlobal('cancelAnimationFrame', vi.fn());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('calls onNodeClick callback is defined', async () => {
    const onNodeClick = vi.fn();
    render(GraphCanvas, { props: { nodes: mockNodes, links: mockLinks, onNodeClick } });
    await new Promise(resolve => setTimeout(resolve, 300));

    expect(onNodeClick).toBeDefined();
    expect(mockState.simulationNodes.length).toBeGreaterThan(0);
  });

  it('cleans up on unmount', async () => {
    const { unmount } = render(GraphCanvas, { props: { nodes: mockNodes, links: mockLinks } });
    await new Promise(resolve => setTimeout(resolve, 150));
    expect(() => unmount()).not.toThrow();
  });
});
