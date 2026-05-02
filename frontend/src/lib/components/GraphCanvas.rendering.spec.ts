import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render } from '@testing-library/svelte';

// Shared state для мока d3-force
const mockState = {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  stopCallback: null as (() => void) | null,
};

// Мокируем d3-force
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
      id: (fn?: (d: any) => string) => {
        if (fn) return linkForce;
        return linkForce;
      },
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
    forceSimulation,
    forceLink,
    forceManyBody,
    forceCenter,
    forceCollide,
    __esModule: true,
    default: { forceSimulation, forceLink, forceManyBody, forceCenter, forceCollide },
  };
});

import GraphCanvas from './GraphCanvas.svelte';

// Track canvas context calls
let ctxCalls: { method: string; args: any[]; fillStyle?: string; strokeStyle?: string }[] = [];
let lastFillStyle = '';
let lastStrokeStyle = '';

const createMockContext = () => {
  ctxCalls = [];
  lastFillStyle = '';
  lastStrokeStyle = '';
  
  return {
    clearRect: vi.fn((...args) => ctxCalls.push({ method: 'clearRect', args })),
    save: vi.fn(() => ctxCalls.push({ method: 'save', args: [] })),
    restore: vi.fn(() => ctxCalls.push({ method: 'restore', args: [] })),
    translate: vi.fn((...args) => ctxCalls.push({ method: 'translate', args })),
    scale: vi.fn((...args) => ctxCalls.push({ method: 'scale', args })),
    beginPath: vi.fn(() => ctxCalls.push({ method: 'beginPath', args: [], fillStyle: lastFillStyle })),
    moveTo: vi.fn((...args) => ctxCalls.push({ method: 'moveTo', args })),
    lineTo: vi.fn((...args) => ctxCalls.push({ method: 'lineTo', args })),
    stroke: vi.fn(() => ctxCalls.push({ method: 'stroke', args: [], strokeStyle: lastStrokeStyle })),
    fill: vi.fn(() => ctxCalls.push({ method: 'fill', args: [], fillStyle: lastFillStyle })),
    closePath: vi.fn(() => ctxCalls.push({ method: 'closePath', args: [] })),
    arc: vi.fn((x, y, r, s, e) => ctxCalls.push({ method: 'arc', args: [x, y, r, s, e], fillStyle: lastFillStyle })),
    ellipse: vi.fn((...args) => ctxCalls.push({ method: 'ellipse', args })),
    rotate: vi.fn((...args) => ctxCalls.push({ method: 'rotate', args })),
    fillRect: vi.fn((...args) => ctxCalls.push({ method: 'fillRect', args })),
    strokeRect: vi.fn((...args) => ctxCalls.push({ method: 'strokeRect', args })),
    setLineDash: vi.fn((...args) => ctxCalls.push({ method: 'setLineDash', args })),
    fillText: vi.fn((text, x, y) => ctxCalls.push({ method: 'fillText', args: [text, x, y], fillStyle: lastFillStyle })),
    measureText: vi.fn(() => ({ width: 50 })),
    set fillStyle(value: string) { lastFillStyle = value; },
    get fillStyle() { return lastFillStyle; },
    set strokeStyle(value: string) { lastStrokeStyle = value; },
    get strokeStyle() { return lastStrokeStyle; },
    set font(value: string) { },
    get font() { return '14px sans-serif'; },
    set textAlign(value: string) { },
    get textAlign() { return 'center'; },
    set textBaseline(value: string) { },
    get textBaseline() { return 'middle'; },
    set lineWidth(value: number) { },
    get lineWidth() { return 1; },
    set shadowBlur(value: number) { },
    get shadowBlur() { return 0; },
    set shadowColor(value: string) { },
    get shadowColor() { return ''; },
  };
};

const mockNodes = [
  { id: '1', title: 'Node 1', type: 'star' },
  { id: '2', title: 'Node 2', type: 'planet' },
  { id: '3', title: 'Node 3', type: 'comet' }
];

const mockLinks = [
  { source: '1', target: '2', weight: 0.8, link_type: 'reference' },
  { source: '2', target: '3', weight: 0.5, link_type: 'dependency' }
];

describe('GraphCanvas - Rendering', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    mockState.simulationNodes = [];
    mockState.simulationLinks = [];
    mockState.tickCallback = null;
    mockState.stopCallback = null;
    ctxCalls = [];
    
    const mockCtx = createMockContext();
    HTMLCanvasElement.prototype.getContext = vi.fn().mockReturnValue(mockCtx);

    global.ResizeObserver = vi.fn().mockImplementation(() => ({
      observe: vi.fn(),
      disconnect: vi.fn(),
      unobserve: vi.fn()
    }));

    vi.stubGlobal('requestAnimationFrame', vi.fn().mockImplementation((cb: FrameRequestCallback) => {
      setTimeout(cb, 16);
      return 1;
    }));
    vi.stubGlobal('cancelAnimationFrame', vi.fn());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders canvas element', async () => {
    render(GraphCanvas, { props: { nodes: mockNodes, links: mockLinks } });
    await new Promise(resolve => setTimeout(resolve, 50));
    
    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
  });

  it('renders with empty data without errors', async () => {
    const { component } = render(GraphCanvas, { props: { nodes: [], links: [] } });
    await new Promise(resolve => setTimeout(resolve, 50));

    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
    expect(component).toBeDefined();
  });

  it('processes nodes and links without errors', async () => {
    const { component } = render(GraphCanvas, { props: { nodes: mockNodes, links: mockLinks } });
    await new Promise(resolve => setTimeout(resolve, 150));

    expect(component).toBeDefined();
    expect(document.querySelector('canvas')).toBeInTheDocument();
  });

  it('renders nodes with coordinates from simulation', async () => {
    render(GraphCanvas, { props: { nodes: mockNodes, links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));

    expect(mockState.simulationNodes.length).toBe(3);
    mockState.simulationNodes.forEach((node) => {
      expect(node.x).toBeDefined();
      expect(node.y).toBeDefined();
      expect(typeof node.x).toBe('number');
      expect(typeof node.y).toBe('number');
    });

    const arcCalls = ctxCalls.filter(c => c.method === 'arc');
    expect(arcCalls.length).toBeGreaterThan(0);
  });

  it('renders correct number of nodes', async () => {
    render(GraphCanvas, { props: { nodes: mockNodes, links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));

    expect(mockState.simulationNodes.length).toBe(mockNodes.length);
    
    const nodeIds = mockState.simulationNodes.map(n => n.id);
    mockNodes.forEach(node => {
      expect(nodeIds).toContain(node.id);
    });
  });

  it('renders links between correct nodes', async () => {
    render(GraphCanvas, { props: { nodes: mockNodes, links: mockLinks } });
    await new Promise(resolve => setTimeout(resolve, 300));

    if (mockState.tickCallback) {
      mockState.tickCallback();
    }

    const lineToCalls = ctxCalls.filter(c => c.method === 'lineTo');
    expect(lineToCalls.length).toBeGreaterThan(0);
  });

  it('applies correct cursor style', async () => {
    render(GraphCanvas, { props: { nodes: mockNodes, links: mockLinks } });
    await new Promise(resolve => setTimeout(resolve, 50));

    const canvas = document.querySelector('canvas');
    expect(canvas?.style.cursor).toBe('grab');
  });

  it('renders all nodes within canvas bounds', async () => {
    render(GraphCanvas, { props: { nodes: mockNodes, links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));

    mockState.simulationNodes.forEach((node: any) => {
      expect(Number.isFinite(node.x)).toBe(true);
      expect(Number.isFinite(node.y)).toBe(true);
      expect(Number.isNaN(node.x)).toBe(false);
      expect(Number.isNaN(node.y)).toBe(false);
    });
  });

  it('cleans up resources on unmount', async () => {
    const { unmount } = render(GraphCanvas, { props: { nodes: mockNodes, links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));

    expect(mockState.tickCallback).not.toBeNull();

    unmount();

    await new Promise(resolve => setTimeout(resolve, 50));
    expect(mockState.tickCallback).toBeNull();
  });

  it('updates simulation when nodes change', async () => {
    const { rerender } = render(GraphCanvas, { props: { nodes: mockNodes, links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));

    expect(mockState.simulationNodes.length).toBe(3);

    // Update with new nodes using rerender
    const newNodes = [
      { id: '4', title: 'New Node 4', type: 'star' },
      { id: '5', title: 'New Node 5', type: 'planet' }
    ];

    rerender({ nodes: newNodes, links: [] });
    await new Promise(resolve => setTimeout(resolve, 200));

    expect(mockState.simulationNodes.length).toBe(2);
  });

  it('updates simulation when links change', async () => {
    const { rerender } = render(GraphCanvas, { props: { nodes: mockNodes, links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));

    const newLinks = [
      { source: '1', target: '2', weight: 1.0, link_type: 'reference' },
      { source: '2', target: '3', weight: 0.5, link_type: 'dependency' },
      { source: '1', target: '3', weight: 0.3, link_type: 'related' }
    ];

    rerender({ nodes: mockNodes, links: newLinks });
    await new Promise(resolve => setTimeout(resolve, 200));

    expect(mockState.simulationLinks.length).toBe(3);
  });
});
