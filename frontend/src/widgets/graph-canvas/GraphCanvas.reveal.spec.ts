import { describe, it, expect, vi, beforeAll, beforeEach, afterEach } from "vitest";
import { render, cleanup } from "@testing-library/svelte";
import { tick } from "svelte";
import type { GraphNode, GraphLink } from "$shared/api/graph";

// Shared state for the d3-force mock
const mockState = {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
};
const forceSimulationCalls: number[] = [];

vi.unmock("$entities/graph-canvas/lib/animation.ts");

vi.mock("d3-force", () => {
  const createMockSimulation = () => {
    const sim: any = {
      nodes: vi.fn((n?: any[]) => {
        if (n) {
          mockState.simulationNodes = n.map((node, i) => ({
            ...node,
            x: node.x ?? 400 + i * 50,
            y: node.y ?? 300 + i * 30,
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
      stop: vi.fn(() => sim),
    };
    return sim;
  };

  return {
    forceSimulation: vi.fn((nodes?: any[]) => {
      const sim = createMockSimulation();
      if (nodes) {
        forceSimulationCalls.push(nodes.length);
        sim.nodes(nodes);
      }
      return sim;
    }),
    forceLink: vi.fn((links?: any[]) => {
      if (links) mockState.simulationLinks = links;
      const linkForce: any = {
        id: vi.fn(() => linkForce),
        distance: vi.fn(() => linkForce),
        strength: vi.fn(() => linkForce),
        links: vi.fn(() => mockState.simulationLinks),
      };
      return linkForce;
    }),
    forceManyBody: vi.fn(() => ({ strength: vi.fn(() => ({})) })),
    forceCenter: vi.fn(() => ({ strength: vi.fn(() => ({})) })),
    forceCollide: vi.fn(() => ({ radius: vi.fn(() => ({})) })),
  };
});

// 50 nodes: a small hub cluster + a long tail, enough to trigger the
// progressive path (threshold is 40).
const mockNodes: GraphNode[] = Array.from({ length: 50 }, (_, i) => ({
  id: `n${i}`,
  title: `Node ${i}`,
  type: "star",
})) as GraphNode[];

const mockLinks: GraphLink[] = [
  // dense core among the first 10 nodes → they sort first by degree
  ...Array.from({ length: 9 }, (_, i) => ({
    source: `n${i}`,
    target: `n${i + 1}`,
    link_type: "reference",
    weight: 0.9,
  })),
  // tail: each later node hangs off n0
  ...Array.from({ length: 40 }, (_, i) => ({
    source: `n${i + 10}`,
    target: "n0",
    link_type: "reference",
    weight: 0.9,
  })),
] as GraphLink[];

let GraphCanvas: any;

beforeAll(async () => {
  GraphCanvas = (await import("$widgets/graph-canvas/GraphCanvas.svelte")).default;
});

function mockCanvas() {
  const ctx: any = new Proxy(
    {},
    {
      get: (_t, prop) => {
        if (prop === "measureText") return () => ({ width: 100 });
        if (prop === "getImageData") return () => ({ data: new Uint8ClampedArray(4) });
        if (prop === "createLinearGradient" || prop === "createRadialGradient")
          return () => ({ addColorStop: vi.fn() });
        return vi.fn();
      },
      set: () => true,
    }
  );
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockImplementation((id) =>
    id === "2d" ? ctx : null
  );
  vi.spyOn(Element.prototype, "getBoundingClientRect").mockReturnValue({
    x: 0,
    y: 0,
    width: 800,
    height: 600,
    top: 0,
    left: 0,
    right: 800,
    bottom: 600,
    toJSON: () => ({}),
  } as any);
}

describe("GraphCanvas progressive reveal (UI-LOAD-1)", () => {
  beforeEach(() => {
    mockState.simulationNodes = [];
    mockState.simulationLinks = [];
    mockState.tickCallback = null;
    forceSimulationCalls.length = 0;
    mockCanvas();
    vi.stubGlobal(
      "ResizeObserver",
      class {
        observe() {}
        unobserve() {}
        disconnect() {}
      }
    );
    vi.stubGlobal(
      "requestAnimationFrame",
      vi.fn(() => 1)
    );
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
    vi.useFakeTimers();
  });

  afterEach(() => {
    cleanup();
    vi.clearAllTimers();
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it(
    "starts with a first batch, then grows without rebuilding the simulation",
    { timeout: 60000 },
    async () => {
      render(GraphCanvas, {
        props: { nodes: mockNodes, links: mockLinks, progressiveReveal: true },
      });
      await tick();
      vi.advanceTimersByTime(50);

      // Simulation was created exactly once, with the head batch only.
      expect(forceSimulationCalls.length).toBe(1);
      expect(forceSimulationCalls[0]).toBeLessThan(50);
      expect(mockState.simulationNodes.length).toBe(forceSimulationCalls[0]);

      // The "N of M" chip is visible while batches are still arriving.
      await tick();
      const chip = document.querySelector('[data-testid="reveal-progress"]');
      expect(chip).not.toBeNull();
      expect(chip!.textContent).toContain("50");

      // Let all batches arrive.
      vi.advanceTimersByTime(120 * 10);
      await tick();

      expect(mockState.simulationNodes.length).toBe(50);
      // Still one simulation instance — additions joined the live layout.
      expect(forceSimulationCalls.length).toBe(1);
      expect(document.querySelector('[data-testid="reveal-progress"]')).toBeNull();
    }
  );

  it("does not reveal progressively below the size threshold", async () => {
    const smallNodes = mockNodes.slice(0, 10);
    const smallLinks = mockLinks.filter(
      (l: any) =>
        (l.source === "n0" || Number(l.source.slice(1)) < 10) &&
        (l.target === "n0" || Number(l.target.slice(1)) < 10)
    );
    render(GraphCanvas, {
      props: { nodes: smallNodes, links: smallLinks, progressiveReveal: true },
    });
    await tick();
    vi.advanceTimersByTime(50);

    expect(forceSimulationCalls.length).toBe(1);
    expect(forceSimulationCalls[0]).toBe(10);
    expect(document.querySelector('[data-testid="reveal-progress"]')).toBeNull();
  });

  it("keeps legacy behavior when progressiveReveal is off", async () => {
    render(GraphCanvas, {
      props: { nodes: mockNodes, links: mockLinks, progressiveReveal: false },
    });
    await tick();
    vi.advanceTimersByTime(50);

    expect(forceSimulationCalls[0]).toBe(50);
    expect(document.querySelector('[data-testid="reveal-progress"]')).toBeNull();
  });
});
