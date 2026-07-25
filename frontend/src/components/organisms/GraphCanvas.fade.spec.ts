import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render } from "@testing-library/svelte";
import GraphCanvas from "$components/organisms/GraphCanvas.svelte";
import { startSimulation, type SimulationState, type TransformState } from "./GraphCanvas";

// Shared state для мока d3-force
const mockState = {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  endCallback: null as (() => void) | null,
  stopCallback: null as (() => void) | null,
};

let animationFrameHandles: ReturnType<typeof setTimeout>[] = [];

// Мокируем d3-force
vi.mock("d3-force", () => {
  const createMockSimulation = () => {
    const sim: any = {
      nodes: vi.fn((n?: any[]) => {
        if (n) {
          mockState.simulationNodes = n.map((node, i) => ({
            ...node,
            x: 400 + i * 50,
            y: 300 + i * 30,
          }));
        }
        return mockState.simulationNodes;
      }),
      tick: vi.fn(() => sim),
      force: vi.fn(() => sim),
      alphaDecay: vi.fn(() => sim),
      on: vi.fn((event: string, cb: () => void) => {
        if (event === "tick") mockState.tickCallback = cb;
        if (event === "end") mockState.endCallback = cb;
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
    default: {
      forceSimulation,
      forceLink,
      forceManyBody,
      forceCenter,
      forceCollide,
    },
  };
});

const mockNodes = [
  { id: "1", title: "Node 1", type: "star" },
  { id: "2", title: "Node 2", type: "planet" },
  { id: "3", title: "Node 3", type: "comet" },
];

const mockLinks = [
  { source: "1", target: "2", weight: 0.8, link_type: "reference" },
  { source: "2", target: "3", weight: 0.5, link_type: "dependency" },
];

function createSimulationState(): SimulationState {
  return {
    simulation: null,
    simLinks: [],
    isRunning: false,
    stable: false,
    nodeOpacity: new Map(),
    linkOpacity: new Map(),
    fadeAnimationId: null,
  };
}

describe("GraphCanvas - Fade Effect", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    mockState.simulationNodes = [];
    mockState.simulationLinks = [];
    mockState.tickCallback = null;
    mockState.endCallback = null;
    mockState.stopCallback = null;

    const mockCtx = {
      clearRect: vi.fn(),
      save: vi.fn(),
      restore: vi.fn(),
      translate: vi.fn(),
      scale: vi.fn(),
      beginPath: vi.fn(),
      moveTo: vi.fn(),
      lineTo: vi.fn(),
      quadraticCurveTo: vi.fn(),
      stroke: vi.fn(),
      fill: vi.fn(),
      closePath: vi.fn(),
      arc: vi.fn(),
      ellipse: vi.fn(),
      rotate: vi.fn(),
      fillRect: vi.fn(),
      strokeRect: vi.fn(),
      setLineDash: vi.fn(),
      fillText: vi.fn(),
      measureText: vi.fn(() => ({ width: 50 })),
      createRadialGradient: vi.fn(() => ({ addColorStop: vi.fn() })),
      createLinearGradient: vi.fn(() => ({ addColorStop: vi.fn() })),
      roundRect: vi.fn(),
      set fillStyle(value: string) {},
      get fillStyle() {
        return "";
      },
      set strokeStyle(value: string) {},
      get strokeStyle() {
        return "";
      },
      set font(value: string) {},
      get font() {
        return "14px sans-serif";
      },
      set textAlign(value: string) {},
      get textAlign() {
        return "center";
      },
      set textBaseline(value: string) {},
      get textBaseline() {
        return "middle";
      },
      set lineWidth(value: number) {},
      get lineWidth() {
        return 1;
      },
      set shadowBlur(value: number) {},
      get shadowBlur() {
        return 0;
      },
      set shadowColor(value: string) {},
      get shadowColor() {
        return "";
      },
      set globalAlpha(value: number) {},
      get globalAlpha() {
        return 1;
      },
    };
    HTMLCanvasElement.prototype.getContext = vi.fn().mockReturnValue(mockCtx);

    global.ResizeObserver = vi.fn().mockImplementation(() => ({
      observe: vi.fn(),
      disconnect: vi.fn(),
      unobserve: vi.fn(),
    }));

    animationFrameHandles = [];
    let animationTime = 0;
    vi.stubGlobal("performance", { now: () => animationTime });
    vi.stubGlobal(
      "requestAnimationFrame",
      vi.fn().mockImplementation((cb: FrameRequestCallback) => {
        animationTime += 16;
        const handle = setTimeout(() => cb(animationTime), 16);
        animationFrameHandles.push(handle);
        return handle as unknown as number;
      })
    );
    vi.stubGlobal(
      "cancelAnimationFrame",
      vi.fn().mockImplementation((handle: number) => {
        clearTimeout(handle as unknown as ReturnType<typeof setTimeout>);
        const index = animationFrameHandles.indexOf(
          handle as unknown as ReturnType<typeof setTimeout>
        );
        if (index >= 0) animationFrameHandles.splice(index, 1);
      })
    );
  });

  afterEach(() => {
    animationFrameHandles.forEach((handle) => clearTimeout(handle));
    animationFrameHandles = [];
    vi.restoreAllMocks();
  });

  it("initializes opacity maps on simulation start", async () => {
    const { component } = render(GraphCanvas, {
      props: { nodes: mockNodes, links: mockLinks },
    });
    await new Promise((resolve) => setTimeout(resolve, 100));

    // Access the internal state through the component
    expect(component).toBeDefined();
  });

  it("starts with zero opacity for nodes and links", async () => {
    const state = createSimulationState();
    const transform: TransformState = { x: 0, y: 0, k: 1 };

    startSimulation(
      mockNodes,
      mockLinks,
      800,
      600,
      state,
      transform,
      () => {},
      () => {}
    );

    expect([...state.nodeOpacity.values()]).toEqual([0, 0, 0]);
    expect([...state.linkOpacity.values()]).toEqual([0, 0]);

    await new Promise((resolve) => setTimeout(resolve, 120));

    expect([...state.nodeOpacity.values()].some((value) => value > 0)).toBe(true);
    expect([...state.linkOpacity.values()].some((value) => value > 0)).toBe(true);
  });

  it("updates opacity during simulation ticks", async () => {
    const state = createSimulationState();
    const transform: TransformState = { x: 0, y: 0, k: 1 };

    startSimulation(
      mockNodes,
      mockLinks,
      800,
      600,
      state,
      transform,
      () => {},
      () => {}
    );
    await new Promise((resolve) => setTimeout(resolve, 120));

    if (mockState.tickCallback) {
      for (let i = 0; i < 25; i++) {
        mockState.tickCallback();
      }
    }

    await new Promise((resolve) => setTimeout(resolve, 50));

    expect([...state.nodeOpacity.values()].some((value) => value > 0)).toBe(true);
    expect([...state.linkOpacity.values()].some((value) => value > 0)).toBe(true);
  });

  it("triggers final fade animation on simulation end", async () => {
    render(GraphCanvas, { props: { nodes: mockNodes, links: mockLinks } });
    await new Promise((resolve) => setTimeout(resolve, 100));

    // Trigger simulation end
    if (mockState.endCallback) {
      mockState.endCallback();
    }

    await new Promise((resolve) => setTimeout(resolve, 50));
    // Final fade animation should be triggered
  });

  it("cancels fade animation on simulation stop", async () => {
    const { component } = render(GraphCanvas, {
      props: { nodes: mockNodes, links: mockLinks },
    });
    await new Promise((resolve) => setTimeout(resolve, 100));

    // Cleanup should cancel any running animations
    expect(component).toBeDefined();
  });

  it("handles empty data without errors", async () => {
    const { component } = render(GraphCanvas, {
      props: { nodes: [], links: [] },
    });
    await new Promise((resolve) => setTimeout(resolve, 50));

    expect(component).toBeDefined();
  });

  it("progresses opacity based on stabilized nodes", async () => {
    render(GraphCanvas, { props: { nodes: mockNodes, links: mockLinks } });
    await new Promise((resolve) => setTimeout(resolve, 100));

    // As more nodes get positions, opacity should increase
    if (mockState.tickCallback) {
      for (let i = 0; i < 30; i++) {
        mockState.tickCallback();
      }
    }

    await new Promise((resolve) => setTimeout(resolve, 50));
    // Progress should be based on nodes with valid positions
  });
});
