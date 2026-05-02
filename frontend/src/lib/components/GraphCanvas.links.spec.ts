import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render } from '@testing-library/svelte';

const mockState = {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  stopCallback: null as (() => void) | null,
};

let ctxCalls: { method: string; args: any[]; fillStyle?: string; strokeStyle?: string }[] = [];
let lastFillStyle = '';
let lastStrokeStyle = '';

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
    set font(value: string) { }, get font() { return '14px sans-serif'; },
    set textAlign(value: string) { }, get textAlign() { return 'center'; },
    set textBaseline(value: string) { }, get textBaseline() { return 'middle'; },
    set lineWidth(value: number) { }, get lineWidth() { return 1; },
    set shadowBlur(value: number) { }, get shadowBlur() { return 0; },
    set shadowColor(value: string) { }, get shadowColor() { return ''; },
  };
};

const mockNodes = [
  { id: '1', title: 'Node 1', type: 'star' },
  { id: '2', title: 'Node 2', type: 'planet' },
  { id: '3', title: 'Node 3', type: 'comet' }
];

describe('GraphCanvas - Link Type Rendering', () => {
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

  it('renders reference links', async () => {
    render(GraphCanvas, { 
      props: { 
        nodes: mockNodes, 
        links: [{ source: '1', target: '2', weight: 0.8, link_type: 'reference' }] 
      } 
    });
    await new Promise(resolve => setTimeout(resolve, 300));

    if (mockState.tickCallback) mockState.tickCallback();
    expect(ctxCalls.filter(c => c.method === 'lineTo').length).toBeGreaterThan(0);
  });

  it('renders dependency links', async () => {
    render(GraphCanvas, { 
      props: { 
        nodes: mockNodes, 
        links: [{ source: '1', target: '2', weight: 0.8, link_type: 'dependency' }] 
      } 
    });
    await new Promise(resolve => setTimeout(resolve, 300));

    if (mockState.tickCallback) mockState.tickCallback();
    expect(ctxCalls.filter(c => c.method === 'stroke').length).toBeGreaterThan(0);
  });

  it('renders related links', async () => {
    render(GraphCanvas, { 
      props: { 
        nodes: mockNodes, 
        links: [{ source: '1', target: '2', weight: 0.6, link_type: 'related' }] 
      } 
    });
    await new Promise(resolve => setTimeout(resolve, 300));

    if (mockState.tickCallback) mockState.tickCallback();
    expect(ctxCalls.filter(c => c.method === 'lineTo').length).toBeGreaterThan(0);
  });

  it('renders links with different weights', async () => {
    render(GraphCanvas, { 
      props: { 
        nodes: mockNodes, 
        links: [
          { source: '1', target: '2', weight: 1.0, link_type: 'reference' },
          { source: '2', target: '3', weight: 0.3, link_type: 'reference' }
        ] 
      } 
    });
    await new Promise(resolve => setTimeout(resolve, 300));

    if (mockState.tickCallback) mockState.tickCallback();
    const lineCalls = ctxCalls.filter(c => c.method === 'lineTo');
    expect(lineCalls.length).toBeGreaterThan(0);
  });
});
