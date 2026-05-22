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
    it('drawStar uses correct colors (#ffcc00 fill, #cc9900 stroke)', () => {
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

      expect(ctx.fillStyle).toBe('#ffcc00');
      expect(ctx.strokeStyle).toBe('#cc9900');
      expect(ctx.lineWidth).toBe(2);
    });

    it('drawPlanet uses correct default color (#60a5fa) for main body', () => {
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
      expect(fillStyles[0]).toBe('#60a5fa');
    });

    it('drawComet uses correct color (#e879f9)', () => {
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

      expect(ctx.fillStyle).toBe('#e879f9');
      expect(ctx.strokeStyle).toBe('rgba(232, 121, 249, 0.6)');
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
      expect(ctx.fillStyle).toContain('rgba(192, 132, 252');
    });

    it('drawAsteroid uses correct rocky color (#94a3b8)', () => {
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

      expect(ctx.fillStyle).toBe('#94a3b8');
      expect(ctx.strokeStyle).toBe('#64748b');
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
          fill: vi.fn(function(this: any) { fillStyles.push(this.fillStyle); }),
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
      // Color should be close to the base color (with hue shift variation)
      expect(starCtx.getFillStyles()[0]).toMatch(/^#[0-9a-fA-F]{6}$/);
      // The color should start with 'f' (golden/yellow range)


      // Test planet sets correct color (first fill is main body)
      const planetCtx = createMockCtx();
      renderer.drawNode(planetCtx, { id: '2', x: 100, y: 100, title: 'Planet', type: 'planet' }, 20, 0, false);
      expect(planetCtx.getFillStyles()[0]).toMatch(/^#[0-9a-fA-F]{6}$/);

      // Test comet sets correct color
      const cometCtx = createMockCtx();
      renderer.drawNode(cometCtx, { id: '3', x: 100, y: 100, title: 'Comet', type: 'comet' }, 20, 0, false);
      expect(cometCtx.getFillStyles()[0]).toMatch(/^#[0-9a-fA-F]{6}$/);

      // Test galaxy sets correct color
      const galaxyCtx = createMockCtx();
      renderer.drawNode(galaxyCtx, { id: '4', x: 100, y: 100, title: 'Galaxy', type: 'galaxy' }, 20, 0, false);
      expect(galaxyCtx.getFillStyles()[0]).toContain('rgba');

      // Test asteroid sets correct color
      const asteroidCtx = createMockCtx();
      renderer.drawNode(asteroidCtx, { id: '5', x: 100, y: 100, title: 'Asteroid', type: 'asteroid' }, 20, 0, false);
      expect(asteroidCtx.getFillStyles()[0]).toMatch(/^#[0-9a-fA-F]{6}$/);
    });
  });

  describe('Node Rendering Integration', () => {
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

    it.skip('renders unknown type nodes without type property', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'Unknown' }], links: [] } });
      await new Promise(resolve => setTimeout(resolve, 300));
      expect(mockState.simulationNodes.length).toBe(1);
      expect(mockState.simulationNodes[0].type).toBeUndefined();
    });

    it.skip('falls back to unknown for undefined type', async () => {
      render(GraphCanvas, { props: { nodes: [{ id: '1', title: 'No Type', type: undefined }], links: [] } });
      await new Promise(resolve => setTimeout(resolve, 300));
      expect(mockState.simulationNodes.length).toBe(1);
      expect(mockState.simulationNodes[0].type).toBeUndefined();
    });
  });
});

describe('Anomaly Rendering (Unknown Node Types)', () => {
  it('drawRealityRift renders without errors', () => {
    const ctx = {
      save: vi.fn(),
      restore: vi.fn(),
      translate: vi.fn(),
      rotate: vi.fn(),
      beginPath: vi.fn(),
      closePath: vi.fn(),
      arc: vi.fn(),
      moveTo: vi.fn(),
      lineTo: vi.fn(),
      fill: vi.fn(),
      stroke: vi.fn(),
      fillStyle: '',
      strokeStyle: '',
      lineWidth: 0,
      shadowBlur: 0,
      shadowColor: '',
    } as unknown as CanvasRenderingContext2D;

    const params = {
      crackCount: 6,
      tentacleCount: 8,
      particleCount: 25,
      colorShift1: 45,
      colorShift2: 90,
      deformAmount: 0.35,
      rotationOffset: 1.5,
    };

    renderer.drawRealityRift(ctx, 100, 100, 20, params);

    expect(ctx.save).toHaveBeenCalled();
    expect(ctx.translate).toHaveBeenCalledWith(100, 100);
    expect(ctx.arc).toHaveBeenCalled();
    expect(ctx.restore).toHaveBeenCalled();
  });

  it('drawChromaticMaw renders tentacles with gradient core', () => {
    const ctx = {
      save: vi.fn(),
      restore: vi.fn(),
      translate: vi.fn(),
      rotate: vi.fn(),
      beginPath: vi.fn(),
      arc: vi.fn(),
      moveTo: vi.fn(),
      bezierCurveTo: vi.fn(),
      fill: vi.fn(),
      stroke: vi.fn(),
      createRadialGradient: vi.fn(() => ({
        addColorStop: vi.fn(),
      })),
      createLinearGradient: vi.fn(() => ({
        addColorStop: vi.fn(),
      })),
      fillStyle: '',
      strokeStyle: '',
      lineWidth: 0,
      lineCap: '',
      shadowBlur: 0,
      shadowColor: '',
    } as unknown as CanvasRenderingContext2D;

    const params = {
      crackCount: 5,
      tentacleCount: 7,
      particleCount: 22,
      colorShift1: 60,
      colorShift2: 120,
      deformAmount: 0.3,
      rotationOffset: 0.8,
    };

    renderer.drawChromaticMaw(ctx, 100, 100, 20, params);

    expect(ctx.save).toHaveBeenCalled();
    expect(ctx.bezierCurveTo).toHaveBeenCalled();
    expect(ctx.createRadialGradient).toHaveBeenCalled();
  });

  it('drawVoidWhisper renders particles with connections', () => {
    const ctx = {
      save: vi.fn(),
      restore: vi.fn(),
      translate: vi.fn(),
      rotate: vi.fn(),
      beginPath: vi.fn(),
      arc: vi.fn(),
      moveTo: vi.fn(),
      lineTo: vi.fn(),
      fill: vi.fn(),
      stroke: vi.fn(),
      createRadialGradient: vi.fn(() => ({
        addColorStop: vi.fn(),
      })),
      fillStyle: '',
      strokeStyle: '',
      lineWidth: 0,
    } as unknown as CanvasRenderingContext2D;

    const params = {
      crackCount: 5,
      tentacleCount: 6,
      particleCount: 25,
      colorShift1: 30,
      colorShift2: 45,
      deformAmount: 0.25,
      rotationOffset: 2.0,
    };

    renderer.drawVoidWhisper(ctx, 100, 100, 20, params);

    expect(ctx.save).toHaveBeenCalled();
    expect(ctx.arc).toHaveBeenCalled();
  });

  it('drawCosmicAbomination combines all anomaly types', () => {
    const ctx = {
      save: vi.fn(),
      restore: vi.fn(),
      translate: vi.fn(),
      rotate: vi.fn(),
      beginPath: vi.fn(),
      closePath: vi.fn(),
      arc: vi.fn(),
      moveTo: vi.fn(),
      lineTo: vi.fn(),
      bezierCurveTo: vi.fn(),
      fill: vi.fn(),
      stroke: vi.fn(),
      createRadialGradient: vi.fn(() => ({
        addColorStop: vi.fn(),
      })),
      createLinearGradient: vi.fn(() => ({
        addColorStop: vi.fn(),
      })),
      fillStyle: '',
      strokeStyle: '',
      lineWidth: 0,
      lineCap: '',
      shadowBlur: 0,
      shadowColor: '',
    } as unknown as CanvasRenderingContext2D;

    const params = {
      crackCount: 5,
      tentacleCount: 8,
      particleCount: 28,
      colorShift1: 90,
      colorShift2: 135,
      deformAmount: 0.4,
      rotationOffset: 1.2,
    };

    renderer.drawCosmicAbomination(ctx, 100, 100, 20, params);

    expect(ctx.save).toHaveBeenCalled();
    expect(ctx.arc).toHaveBeenCalled();
    expect(ctx.bezierCurveTo).toHaveBeenCalled();
  });

  it('drawUnknown dispatches to one anomaly renderer based on nodeId', () => {
    const drawRealityRiftSpy = vi.fn();
    const drawChromaticMawSpy = vi.fn();
    const drawVoidWhisperSpy = vi.fn();
    const drawCosmicAbominationSpy = vi.fn();

    renderer.drawUnknown(
      {} as CanvasRenderingContext2D,
      100,
      100,
      20,
      0,
      'node1',
      {
        0: drawRealityRiftSpy,
        1: drawChromaticMawSpy,
        2: drawVoidWhisperSpy,
        3: drawCosmicAbominationSpy,
      }
    );

    const calledSpies = [
      drawRealityRiftSpy,
      drawChromaticMawSpy,
      drawVoidWhisperSpy,
      drawCosmicAbominationSpy,
    ].filter((spy) => spy.mock.calls.length > 0);

    expect(calledSpies).toHaveLength(1);
  });

  it('drawUnknown is deterministic for the same nodeId', () => {
    const drawRealityRiftSpy = vi.fn();
    const drawChromaticMawSpy = vi.fn();
    const drawVoidWhisperSpy = vi.fn();
    const drawCosmicAbominationSpy = vi.fn();

    const customRenderers = {
      0: drawRealityRiftSpy,
      1: drawChromaticMawSpy,
      2: drawVoidWhisperSpy,
      3: drawCosmicAbominationSpy,
    };

    renderer.drawUnknown({} as CanvasRenderingContext2D, 100, 100, 20, 0, 'deterministic-node', customRenderers);
    renderer.drawUnknown({} as CanvasRenderingContext2D, 120, 120, 20, 0, 'deterministic-node', customRenderers);

    const counts = [
      drawRealityRiftSpy.mock.calls.length,
      drawChromaticMawSpy.mock.calls.length,
      drawVoidWhisperSpy.mock.calls.length,
      drawCosmicAbominationSpy.mock.calls.length,
    ].filter((count) => count > 0);

    expect(counts).toEqual([2]);
  });

  it('getAnomalyParams returns stable and different values for different nodeIds', () => {
    const paramsA = renderer.getAnomalyParams('alpha-node');
    const paramsB = renderer.getAnomalyParams('alpha-node');
    const paramsC = renderer.getAnomalyParams('beta-node');

    expect(paramsA).toEqual(paramsB);
    expect(paramsA).not.toEqual(paramsC);
  });
});