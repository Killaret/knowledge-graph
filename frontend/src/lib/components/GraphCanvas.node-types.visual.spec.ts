/**
 * Visual correctness tests for GraphCanvas node types
 * These tests verify that celestial bodies are rendered with correct visual styles
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render } from '@testing-library/svelte';

// Mock canvas context to capture drawing calls
interface CanvasCall {
  method: string;
  value?: unknown;
  x?: number;
  y?: number;
  r?: number;
  rx?: number;
  ry?: number;
  rotation?: number;
  startAngle?: number;
  endAngle?: number;
  angle?: number;
  dash?: number[];
  text?: string;
}
const mockCanvasCalls: CanvasCall[] = [];

const mockState = {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
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

  const forceLink = vi.fn((_links?: any[]) => {
    const linkForce: any = {
      id: () => linkForce,
      distance: () => linkForce,
      strength: () => linkForce,
      links: () => [],
    };
    return linkForce;
  });

  return {
    forceSimulation,
    forceLink,
    forceManyBody: vi.fn(() => ({ strength: vi.fn() })),
    forceCenter: vi.fn(() => ({ strength: vi.fn() })),
    forceCollide: vi.fn(() => ({ radius: vi.fn() })),
  };
});

import GraphCanvas from './GraphCanvas.svelte';

describe('GraphCanvas - Visual Node Type Correctness', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    mockCanvasCalls.length = 0;
    mockState.simulationNodes = [];

    // Enhanced mock that captures all canvas operations
    const createMockContext = () => ({
      clearRect: vi.fn(),
      save: vi.fn(),
      restore: vi.fn(),
      translate: vi.fn(),
      scale: vi.fn(),
      beginPath: vi.fn(() => mockCanvasCalls.push({ method: 'beginPath' })),
      moveTo: vi.fn((x: number, y: number) => mockCanvasCalls.push({ method: 'moveTo', x, y })),
      lineTo: vi.fn((x: number, y: number) => mockCanvasCalls.push({ method: 'lineTo', x, y })),
      stroke: vi.fn(() => mockCanvasCalls.push({ method: 'stroke' })),
      fill: vi.fn(() => mockCanvasCalls.push({ method: 'fill' })),
      closePath: vi.fn(),
      arc: vi.fn((x, y, r, s, e) => mockCanvasCalls.push({ method: 'arc', x, y, r, startAngle: s, endAngle: e })),
      ellipse: vi.fn((x, y, rx, ry, rot, s, e) => 
        mockCanvasCalls.push({ method: 'ellipse', x, y, rx, ry, rotation: rot, startAngle: s, endAngle: e })),
      rotate: vi.fn((angle: number) => mockCanvasCalls.push({ method: 'rotate', angle })),
      fillRect: vi.fn(),
      strokeRect: vi.fn(),
      setLineDash: vi.fn((dash: number[]) => mockCanvasCalls.push({ method: 'setLineDash', dash })),
      fillText: vi.fn((text: string, x: number, y: number) => 
        mockCanvasCalls.push({ method: 'fillText', text, x, y })),
      fillStyle: '',
      strokeStyle: '',
      font: '',
      textAlign: 'center',
      textBaseline: 'middle',
      lineWidth: 1,
      shadowBlur: 0,
      shadowColor: '',
    });

    // Use defineProperty to track style changes
    const mockCtx = createMockContext();
    const styleTracker = {
      fillStyle: '',
      strokeStyle: '',
      lineWidth: 1,
      shadowBlur: 0,
      shadowColor: '',
    };

    (Object.keys(styleTracker) as Array<keyof typeof styleTracker>).forEach((key) => {
      Object.defineProperty(mockCtx, key, {
        get: () => styleTracker[key],
        set: (value: unknown) => {
          styleTracker[key] = value as never;
          mockCanvasCalls.push({ method: `set_${key}`, value });
        },
      });
    });

    HTMLCanvasElement.prototype.getContext = vi.fn().mockReturnValue(mockCtx);

    global.ResizeObserver = vi.fn().mockImplementation(() => ({
      observe: vi.fn(),
      disconnect: vi.fn(),
      unobserve: vi.fn(),
    }));

    vi.stubGlobal('requestAnimationFrame', vi.fn((cb: FrameRequestCallback) => {
      setTimeout(cb, 16);
      return 1;
    }));
    vi.stubGlobal('cancelAnimationFrame', vi.fn());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Star visual properties', () => {
    it('should use correct fill and stroke colors for star', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Star', type: 'star' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c: any) => c.method === 'set_fillStyle');
      const strokeStyleCalls = mockCanvasCalls.filter((c: any) => c.method === 'set_strokeStyle');

      // Star should use golden/yellow colors
      expect(fillStyleCalls.some((c: any) => c.value === '#ffdd88')).toBe(true);
      expect(strokeStyleCalls.some((c: any) => c.value === '#cc9900')).toBe(true);
    });

    it('should draw star shape with 5 points', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Star', type: 'star' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const lineToCalls = mockCanvasCalls.filter((c: any) => c.method === 'lineTo');
      const beginPathCalls = mockCanvasCalls.filter((c: any) => c.method === 'beginPath');

      // Star uses lineTo for drawing points - should have multiple calls
      expect(lineToCalls.length).toBeGreaterThanOrEqual(10); // 5 points, 2 lines each
      expect(beginPathCalls.length).toBeGreaterThanOrEqual(1);
    });
  });

  describe('Planet visual properties', () => {
    it('should use correct fill colors for planet', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Planet', type: 'planet' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_fillStyle');
      
      // Planet should use earthy/golden colors
      expect(fillStyleCalls.some((c) => c.value === '#c9b37c' || c.value === '#a57c2c')).toBe(true);
    });

    it('should draw circular base with ellipse bands', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Planet', type: 'planet' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const arcCalls = mockCanvasCalls.filter((c) => c.method === 'arc');
      const ellipseCalls = mockCanvasCalls.filter((c) => c.method === 'ellipse');

      // Planet should have base circle
      expect(arcCalls.length).toBeGreaterThanOrEqual(1);
      // And ellipse bands
      expect(ellipseCalls.length).toBeGreaterThanOrEqual(4);
    });
  });

  describe('Comet visual properties', () => {
    it('should use correct fill colors for comet', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Comet', type: 'comet' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_fillStyle');
      const strokeStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_strokeStyle');

      // Comet should use cyan/aqua colors
      expect(fillStyleCalls.some((c: CanvasCall) => c.value === '#aaffdd')).toBe(true);
      expect(strokeStyleCalls.some((c: CanvasCall) => (c.value as string).includes('170, 255, 221'))).toBe(true);
    });

    it('should draw comet with tail', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Comet', type: 'comet' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const moveToCalls = mockCanvasCalls.filter((c) => c.method === 'moveTo');
      const lineToCalls = mockCanvasCalls.filter((c) => c.method === 'lineTo');

      // Comet needs head circle and tail line
      expect(moveToCalls.length).toBeGreaterThanOrEqual(1);
      expect(lineToCalls.length).toBeGreaterThanOrEqual(1);
    });
  });

  describe('Galaxy visual properties', () => {
    it('should use correct translucent colors for galaxy', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Galaxy', type: 'galaxy' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_fillStyle');

      // Galaxy should use purple/lavender translucent colors
      expect(fillStyleCalls.some((c: CanvasCall) => (c.value as string).includes('200, 180, 255'))).toBe(true);
      expect(fillStyleCalls.some((c: CanvasCall) => (c.value as string).includes('0.3') || (c.value as string).includes('0.2') || (c.value as string).includes('0.1'))).toBe(true);
    });

    it('should draw multiple ellipses for galaxy arms', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Galaxy', type: 'galaxy' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const ellipseCalls = mockCanvasCalls.filter((c) => c.method === 'ellipse');

      // Galaxy should have 3 elliptical arms
      expect(ellipseCalls.length).toBeGreaterThanOrEqual(3);
    });
  });

  describe('Asteroid visual properties', () => {
    it('should use correct rocky colors for asteroid', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Asteroid', type: 'asteroid' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_fillStyle');
      const strokeStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_strokeStyle');

      // Asteroid should use brown/grey rocky colors
      expect(fillStyleCalls.some((c) => c.value === '#8b7355')).toBe(true);
      expect(strokeStyleCalls.some((c) => c.value === '#5c4a3a')).toBe(true);
    });

    it('should draw irregular polygon shape', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Asteroid', type: 'asteroid' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const lineToCalls = mockCanvasCalls.filter((c) => c.method === 'lineTo');

      // Asteroid uses 7 points for irregular shape
      expect(lineToCalls.length).toBeGreaterThanOrEqual(7);
    });
  });

  describe('Nebula visual properties', () => {
    it('should use correct turquoise colors for nebula', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Nebula', type: 'nebula' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_fillStyle');

      // Nebula should use turquoise colors
      expect(fillStyleCalls.some((c: CanvasCall) => (c.value as string).includes('100, 220, 220'))).toBe(true);
      expect(fillStyleCalls.some((c: CanvasCall) => (c.value as string).includes('0.25') || (c.value as string).includes('0.2') || (c.value as string).includes('0.15'))).toBe(true);
    });

    it('should draw multiple ellipses for nebula clouds', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Nebula', type: 'nebula' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const ellipseCalls = mockCanvasCalls.filter((c) => c.method === 'ellipse');

      // Nebula should have 4 elliptical clouds
      expect(ellipseCalls.length).toBeGreaterThanOrEqual(4);
    });
  });

  describe('Debris visual properties', () => {
    it('should use correct grey colors for debris', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Debris', type: 'debris' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_fillStyle');

      // Debris should use semi-transparent grey
      expect(fillStyleCalls.some((c) => c.value === 'rgba(150, 150, 150, 0.6)')).toBe(true);
    });

    it('should draw scattered particle circles', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Debris', type: 'debris' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const arcCalls = mockCanvasCalls.filter((c) => c.method === 'arc');

      // Debris should have 5 scattered particles
      expect(arcCalls.length).toBeGreaterThanOrEqual(5);
    });
  });

  describe('Black hole visual properties', () => {
    it('should use black fill for event horizon', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Black Hole', type: 'blackhole' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_fillStyle');

      // Black hole event horizon is pure black
      expect(fillStyleCalls.some((c) => c.value === '#000000')).toBe(true);
    });

    it('should draw accretion disk with orange glow', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Black Hole', type: 'blackhole' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const strokeStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_strokeStyle');

      // Accretion disk uses orange glow rings
      expect(strokeStyleCalls.some((c) => c.value === '#ff6600' || c.value === '#ff3300')).toBe(true);
    });

    it('should draw multiple concentric rings', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Black Hole', type: 'blackhole' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const arcCalls = mockCanvasCalls.filter((c) => c.method === 'arc');

      // Black hole: event horizon + 2 accretion rings
      expect(arcCalls.length).toBeGreaterThanOrEqual(3);
    });
  });

  describe('Moon visual properties', () => {
    it('should use correct grey colors for moon', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Moon', type: 'moon' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_fillStyle');
      const strokeStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_strokeStyle');

      // Moon body is light grey
      expect(fillStyleCalls.some((c) => c.value === '#cccccc')).toBe(true);
      // Moon stroke is darker grey
      expect(strokeStyleCalls.some((c) => c.value === '#999999')).toBe(true);
    });

    it('should draw moon with crater', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Moon', type: 'moon' }], links: [] } });
      await new Promise((resolve) => setTimeout(resolve, 300));

      const fillStyleCalls = mockCanvasCalls.filter((c) => c.method === 'set_fillStyle');

      // Moon has body (#cccccc) and crater (#aaaaaa)
      expect(fillStyleCalls.some((c) => c.value === '#cccccc')).toBe(true);
      expect(fillStyleCalls.some((c) => c.value === '#aaaaaa')).toBe(true);
    });
  });

  describe('Type uniqueness', () => {
    it('should use unique color schemes for each celestial type', async () => {
      const types = ['star', 'planet', 'comet', 'galaxy', 'asteroid', 'nebula', 'debris', 'blackhole', 'moon'] as const;
      const typeColors: Record<string, string[]> = {};

      for (const type of types) {
        mockCanvasCalls.length = 0;
        render(GraphCanvas, { 
          props: { nodes: [{ id: '1', title: type, type }], links: [] } 
        });
        await new Promise((resolve) => setTimeout(resolve, 300));

        const colors = mockCanvasCalls
          .filter((c: CanvasCall) => c.method === 'set_fillStyle' && c.value)
          .map((c: CanvasCall) => c.value as string);
        typeColors[type] = [...new Set(colors)];
      }

      // Verify each type has unique primary color
      expect(typeColors['star']).toContain('#ffdd88');
      expect(typeColors['planet'].some(c => c === '#c9b37c' || c === '#a57c2c')).toBe(true);
      expect(typeColors['comet']).toContain('#aaffdd');
      expect(typeColors['galaxy'].some(c => c.includes('200, 180, 255'))).toBe(true);
      expect(typeColors['asteroid']).toContain('#8b7355');
      expect(typeColors['nebula'].some(c => c.includes('100, 220, 220'))).toBe(true);
      expect(typeColors['debris']).toContain('rgba(150, 150, 150, 0.6)');
      expect(typeColors['blackhole']).toContain('#000000');
      expect(typeColors['moon']).toContain('#cccccc');
    });
  });
});
