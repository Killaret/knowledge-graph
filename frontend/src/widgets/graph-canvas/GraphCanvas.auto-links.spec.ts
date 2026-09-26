import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, fireEvent } from "@testing-library/svelte";

// Allow the real animation loop to run so onUpdate/doRedraw are called.
vi.unmock("$entities/graph-canvas/lib/animation.ts");

const mockState = {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  stopCallback: null as (() => void) | null,
};

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
      links: (l?: any[]) => {
        if (l) mockState.simulationLinks = l;
        return mockState.simulationLinks;
      },
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

import GraphCanvas from "$widgets/graph-canvas/GraphCanvas.svelte";
import { graphStore } from "$shared/stores/graph.svelte";

const createMockContext = () => ({
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
});

const mockNodes = [
  { id: "1", title: "Node 1", type: "star" },
  { id: "2", title: "Node 2", type: "planet" },
  { id: "3", title: "Node 3", type: "comet" },
];

const gammaLink = {
  source: "1",
  target: "2",
  weight: 0.8,
  link_type: "related",
  source_type: "gamma",
};
const manualLink = {
  source: "2",
  target: "3",
  weight: 0.8,
  link_type: "reference",
  source_type: "user",
};

describe("GraphCanvas — UI-GRAPH-1 auto-links toggle", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    mockState.simulationNodes = [];
    mockState.simulationLinks = [];
    mockState.tickCallback = null;
    mockState.stopCallback = null;
    graphStore.reset();

    HTMLCanvasElement.prototype.getContext = vi.fn().mockReturnValue(createMockContext());

    global.ResizeObserver = vi.fn().mockImplementation(function () {
      return { observe: vi.fn(), disconnect: vi.fn(), unobserve: vi.fn() };
    });

    vi.stubGlobal(
      "requestAnimationFrame",
      vi.fn().mockImplementation((cb: FrameRequestCallback) => {
        setTimeout(cb, 16);
        return 1;
      })
    );
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
  });

  afterEach(() => {
    vi.restoreAllMocks();
    graphStore.reset();
  });

  it("renders the toggle button reflecting the store state", async () => {
    const { getByTestId } = render(GraphCanvas, {
      props: { nodes: mockNodes, links: [gammaLink, manualLink] },
    });
    const toggle = getByTestId("auto-links-toggle");
    expect(toggle.getAttribute("aria-pressed")).toBe("true");
  });

  it("drops model-suggested links from the simulation when toggled off", async () => {
    const { getByTestId } = render(GraphCanvas, {
      props: { nodes: mockNodes, links: [gammaLink, manualLink] },
    });
    await new Promise((resolve) => setTimeout(resolve, 300));
    expect(mockState.simulationLinks.length).toBe(2);

    await fireEvent.click(getByTestId("auto-links-toggle"));
    await new Promise((resolve) => setTimeout(resolve, 300));

    const sources = mockState.simulationLinks.map((l) => l.source_type ?? "user");
    expect(sources).not.toContain("gamma");
    expect(mockState.simulationLinks.length).toBe(1);

    await fireEvent.click(getByTestId("auto-links-toggle"));
    await new Promise((resolve) => setTimeout(resolve, 300));
    expect(mockState.simulationLinks.length).toBe(2);
  });
});
