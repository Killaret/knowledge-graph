import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, cleanup } from '@testing-library/svelte';
import * as renderer from './GraphCanvas/renderer';

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

describe('GraphCanvas - Node Type Rendering', () => {
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
    cleanup();
  });

  describe('Node Type Colors and Styles', () => {
    it('drawStar uses correct colors (#ffdd88 fill, #cc9900 stroke)', () => {
      const ctx = {
        beginPath: vi.fn(),
        lineTo: vi.fn(),
        closePath: vi.fn(),
        fill: vi.fn(),
        stroke: vi.fn(),
        fillStyle: '',
        strokeStyle: '',
        lineWidth: 0,
      } as unknown as CanvasRenderingContext2D;

      renderer.drawStar(ctx, 100, 100, 20, 0);

      expect(ctx.fillStyle).toBe('#ffdd88');
      expect(ctx.strokeStyle).toBe('#cc9900');
      expect(ctx.lineWidth).toBe(2);
    });

    it('drawPlanet uses correct default color (#c9b37c) for main body', () => {
      const fillStyles: string[] = [];
      const ctx = {
        beginPath: vi.fn(),
        arc: vi.fn(),
        ellipse: vi.fn(),
        fill: vi.fn(() => {
          fillStyles.push((ctx as any).fillStyle);
        }),
        set fillStyle(value: string) { (ctx as any)._fillStyle = value; },
        get fillStyle() { return (ctx as any)._fillStyle || ''; },
      } as unknown as CanvasRenderingContext2D & { _fillStyle: string };

      renderer.drawPlanet(ctx, 100, 100, 20, 0);

      // First fillStyle should be the main planet color #c9b37c
      expect(fillStyles[0]).toBe('#c9b37c');
    });

    it('drawComet uses correct color (#aaffdd)', () => {
      const ctx = {
        beginPath: vi.fn(),
        arc: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        fill: vi.fn(),
        fillStyle: '',
        strokeStyle: '',
        lineWidth: 0,
      } as unknown as CanvasRenderingContext2D;

      renderer.drawComet(ctx, 100, 100, 20, 0);

      expect(ctx.fillStyle).toBe('#aaffdd');
      expect(ctx.strokeStyle).toBe('rgba(170, 255, 221, 0.6)');
      expect(ctx.lineWidth).toBe(4);
    });

    it('drawGalaxy uses purple spiral colors', () => {
      const ctx = {
        save: vi.fn(),
        restore: vi.fn(),
        translate: vi.fn(),
        rotate: vi.fn(),
        beginPath: vi.fn(),
        ellipse: vi.fn(),
        fill: vi.fn(),
        fillStyle: '',
      } as unknown as CanvasRenderingContext2D;

      renderer.drawGalaxy(ctx, 100, 100, 20, 0);

      // Galaxy uses rgba with purple/white colors
      expect(ctx.fillStyle).toContain('rgba(200, 180, 255');
    });

    it('drawAsteroid uses correct rocky color (#8b7355)', () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        closePath: vi.fn(),
        fill: vi.fn(),
        stroke: vi.fn(),
        fillStyle: '',
        strokeStyle: '',
        lineWidth: 0,
      } as unknown as CanvasRenderingContext2D;

      renderer.drawAsteroid(ctx, 100, 100, 20, 0);

      expect(ctx.fillStyle).toBe('#8b7355');
      expect(ctx.strokeStyle).toBe('#5c4a3a');
    });

    it('drawNode sets correct fillStyle for each node type', () => {
      function createMockCtx() {
        const fillStyles: string[] = [];
        return {
          save: vi.fn(),
          restore: vi.fn(),
          translate: vi.fn(),
          rotate: vi.fn(),
          beginPath: vi.fn(),
          lineTo: vi.fn(),
          closePath: vi.fn(),
          arc: vi.fn(),
          ellipse: vi.fn(),
          fill: vi.fn(function() { fillStyles.push(this.fillStyle); }),
          stroke: vi.fn(),
          moveTo: vi.fn(),
          set fillStyle(v: string) { (this as any)._fillStyle = v; },
          get fillStyle() { return (this as any)._fillStyle || ''; },
          strokeStyle: '',
          lineWidth: 0,
          shadowBlur: 0,
          shadowColor: '',
          font: '',
          textAlign: 'center',
          textBaseline: 'middle',
          fillText: vi.fn(),
          measureText: vi.fn(() => ({ width: 50 })),
          getFillStyles: () => fillStyles,
        } as unknown as CanvasRenderingContext2D & { getFillStyles: () => string[] };
      }

      // Test star sets correct color
      const starCtx = createMockCtx();
      renderer.drawNode(starCtx, { id: '1', x: 100, y: 100, title: 'Star', type: 'star' }, 20, 0, false);
      expect(starCtx.getFillStyles()[0]).toBe('#ffdd88');

      // Test planet sets correct color (first fill is main body)
      const planetCtx = createMockCtx();
      renderer.drawNode(planetCtx, { id: '2', x: 100, y: 100, title: 'Planet', type: 'planet' }, 20, 0, false);
      expect(planetCtx.getFillStyles()[0]).toBe('#c9b37c');

      // Test comet sets correct color
      const cometCtx = createMockCtx();
      renderer.drawNode(cometCtx, { id: '3', x: 100, y: 100, title: 'Comet', type: 'comet' }, 20, 0, false);
      expect(cometCtx.getFillStyles()[0]).toBe('#aaffdd');
    });
  });

  it('renders star nodes with coordinates', async () => {
    render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Star', type: 'star' }], links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));
    expect(mockState.simulationNodes.length).toBe(1);
    expect(mockState.simulationNodes[0].type).toBe('star');
    expect(mockState.simulationNodes[0].x).toBeDefined();
    expect(mockState.simulationNodes[0].y).toBeDefined();
  });

  it('renders planet nodes with coordinates', async () => {
    render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Planet', type: 'planet' }], links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));
    expect(mockState.simulationNodes.length).toBe(1);
    expect(mockState.simulationNodes[0].type).toBe('planet');
    expect(mockState.simulationNodes[0].x).toBeDefined();
    expect(mockState.simulationNodes[0].y).toBeDefined();
  });

  it('renders comet nodes with coordinates', async () => {
    render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Comet', type: 'comet' }], links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));
    expect(mockState.simulationNodes.length).toBe(1);
    expect(mockState.simulationNodes[0].type).toBe('comet');
    expect(mockState.simulationNodes[0].x).toBeDefined();
    expect(mockState.simulationNodes[0].y).toBeDefined();
  });

  it('renders galaxy nodes with coordinates', async () => {
    render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Galaxy', type: 'galaxy' }], links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));
    expect(mockState.simulationNodes.length).toBe(1);
    expect(mockState.simulationNodes[0].type).toBe('galaxy');
    expect(mockState.simulationNodes[0].x).toBeDefined();
    expect(mockState.simulationNodes[0].y).toBeDefined();
  });

  it('renders asteroid nodes with coordinates', async () => {
    render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Asteroid', type: 'asteroid' }], links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));
    expect(mockState.simulationNodes.length).toBe(1);
    expect(mockState.simulationNodes[0].type).toBe('asteroid');
    expect(mockState.simulationNodes[0].x).toBeDefined();
    expect(mockState.simulationNodes[0].y).toBeDefined();
  });

  it('renders unknown type nodes without type property', async () => {
    render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Unknown' }], links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));
    expect(mockState.simulationNodes.length).toBe(1);
    expect(mockState.simulationNodes[0].type).toBeUndefined();
  });

  it('falls back to unknown for undefined type', async () => {
    render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'No Type', type: undefined }], links: [] } });
    await new Promise(resolve => setTimeout(resolve, 300));
    expect(mockState.simulationNodes.length).toBe(1);
    expect(mockState.simulationNodes[0].type).toBeUndefined();
  });
});
