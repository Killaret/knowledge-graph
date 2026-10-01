import { describe, it, expect, vi, beforeEach } from "vitest";
import { applyDelta } from "./delta";
import type { SimulationNode, SimulationLink, SimulationState, TransformState } from "./types";
import { resolveLinkEndpoint, getLinkEndpointId } from "./types";
import { resetMockState, createMockSimulation } from "../../../__mocks__/d3-force";

function createState(): SimulationState {
  return {
    simulation: null,
    simLinks: [],
    isRunning: false,
    stable: false,
    nodeOpacity: new Map(),
    linkOpacity: new Map(),
    dyingLinks: [],
    dyingLinkOpacity: new Map(),
    fadeAnimationId: null,
  };
}

function baseOptions(state: SimulationState) {
  return {
    nodes: [{ id: "n1", title: "One" } as SimulationNode],
    links: [] as SimulationLink[],
    width: 800,
    height: 600,
    state,
    transform: { x: 0, y: 0, k: 1 } as TransformState,
    onTick: vi.fn(),
    onResetView: vi.fn(),
  };
}

beforeEach(() => {
  resetMockState();
});

describe("delta", () => {
  it("applies an incremental update when changes are below the threshold", () => {
    const state = createState();
    const options = baseOptions(state);

    const restarted = applyDelta(
      {
        updated_nodes: [{ id: "n1", title: "Updated" }],
      },
      options
    );
    expect(restarted).toBe(false);
  });

  it("performs a full restart when the delta is large", () => {
    const state = createState();
    const options = baseOptions(state);

    const restarted = applyDelta(
      {
        added_nodes: Array.from({ length: 11 }, (_, i) => ({
          id: `n${i + 2}`,
          title: `Node ${i + 2}`,
        })),
      },
      options
    );
    expect(restarted).toBe(true);
    expect(options.onResetView).toHaveBeenCalled();
    expect(state.isRunning).toBe(true);
  });

  // SYNC-1 stage B: small deltas land in place — the simulation object and
  // the positions of untouched nodes survive; nothing re-layouts.
  it("removes nodes and links in place, keeping placed nodes put", () => {
    const state = createState();
    const sim = createMockSimulation();
    sim.nodes([
      { id: "n1", title: "One", x: 111, y: 222 },
      { id: "n2", title: "Two", x: 333, y: 444 },
    ]);
    state.simulation = sim as any;
    state.simLinks = [{ source: "n1", target: "n2", link_type: "related" } as SimulationLink];

    const options = baseOptions(state);
    options.nodes = [
      { id: "n1", title: "One" },
      { id: "n2", title: "Two" },
    ] as SimulationNode[];
    options.links = [{ source: "n1", target: "n2" }] as SimulationLink[];

    const restarted = applyDelta(
      {
        removed_nodes: ["n2"],
        removed_links: [{ source: "n1", target: "n2" } as any],
      },
      options
    );

    expect(restarted).toBe(true);
    expect(state.simulation).toBe(sim);
    expect(options.onResetView).not.toHaveBeenCalled();
    const remaining = sim.nodes();
    expect(remaining.map((n: SimulationNode) => n.id)).toEqual(["n1"]);
    expect(remaining[0].x).toBe(111);
    expect(remaining[0].y).toBe(222);
    expect(state.simLinks).toHaveLength(0);
  });

  it("adds nodes and links in place without a re-layout", () => {
    const state = createState();
    const sim = createMockSimulation();
    sim.nodes([{ id: "n1", title: "One", x: 111, y: 222 }]);
    state.simulation = sim as any;
    state.simLinks = [];

    const options = baseOptions(state);

    const restarted = applyDelta(
      {
        added_nodes: [{ id: "n3", title: "Three" }],
        added_links: [{ source: "n1", target: "n3", link_type: "related" } as any],
      },
      options
    );

    expect(restarted).toBe(true);
    expect(state.simulation).toBe(sim);
    expect(options.onResetView).not.toHaveBeenCalled();
    const ids = sim.nodes().map((n: SimulationNode) => n.id);
    expect(ids).toEqual(["n1", "n3"]);
    expect(sim.nodes()[0].x).toBe(111);
    // The new node spawns near its placed neighbour, not on a fresh ring.
    const n3 = sim.nodes()[1];
    expect(Math.hypot(n3.x - 111, n3.y - 222)).toBeLessThan(200);
    expect(state.simLinks).toHaveLength(1);
  });

  it("updates an existing link in place instead of duplicating it", () => {
    const state = createState();
    const sim = createMockSimulation();
    sim.nodes([
      { id: "n1", title: "One" },
      { id: "n2", title: "Two" },
    ]);
    state.simulation = sim as any;
    state.simLinks = [
      { source: "n1", target: "n2", link_type: "related", weight: 0.5 } as SimulationLink,
    ];

    const options = baseOptions(state);

    applyDelta(
      {
        added_links: [
          { id: "L1", source: "n1", target: "n2", link_type: "related", weight: 0.9 } as any,
        ],
      },
      options
    );

    expect(state.simLinks).toHaveLength(1);
    expect(state.simLinks[0].id).toBe("L1");
    expect(state.simLinks[0].weight).toBe(0.9);
  });

  it("preserves link identity and provenance when a delta lands in place", () => {
    const state = createState();
    const sim = createMockSimulation();
    sim.nodes([
      { id: "n1", title: "One" },
      { id: "n2", title: "Two" },
    ]);
    state.simulation = sim as any;
    state.simLinks = [];

    const options = baseOptions(state);
    options.nodes = [
      { id: "n1", title: "One" },
      { id: "n2", title: "Two" },
    ] as SimulationNode[];

    const restarted = applyDelta(
      {
        added_links: [
          {
            id: "promoted-link",
            source: "n1",
            target: "n2",
            source_type: "user",
            gamma_origin: true,
          },
        ],
      },
      options
    );

    expect(restarted).toBe(true);
    expect(state.simLinks).toEqual([
      expect.objectContaining({
        id: "promoted-link",
        source_type: "user",
        gamma_origin: true,
      }),
    ]);
  });

  it("updates existing simulation nodes during an incremental update", () => {
    const state = createState();
    const fakeNode = { id: "n1", title: "One", x: 10, y: 20 } as SimulationNode;
    state.simulation = {
      nodes: vi.fn().mockReturnValue([fakeNode]),
      alpha: vi.fn().mockReturnThis(),
      restart: vi.fn(),
      stop: vi.fn(),
    } as any;

    const options = baseOptions(state);
    options.nodes = [fakeNode];

    const restarted = applyDelta(
      {
        updated_nodes: [{ id: "n1", title: "Renamed", x: 30, y: 40 }],
      },
      options
    );

    expect(restarted).toBe(true);
    expect(fakeNode.title).toBe("Renamed");
    expect(fakeNode.x).toBe(30);
    expect(fakeNode.y).toBe(40);
  });

  it("does not crash when no simulation exists and only nodes are updated", () => {
    const state = createState();
    const options = baseOptions(state);
    const restarted = applyDelta(
      {
        updated_nodes: [{ id: "n1", title: "Updated" }],
      },
      options
    );
    expect(restarted).toBe(false);
  });

  it("filters updated nodes that are not present in the simulation", () => {
    const state = createState();
    const fakeNode = { id: "n1", title: "One" } as SimulationNode;
    state.simulation = {
      nodes: vi.fn().mockReturnValue([fakeNode]),
      alpha: vi.fn().mockReturnThis(),
      restart: vi.fn(),
      stop: vi.fn(),
    } as any;

    const options = baseOptions(state);
    options.nodes = [fakeNode];

    const restarted = applyDelta(
      {
        updated_nodes: [{ id: "n2", title: "Missing" }],
      },
      options
    );

    // Nothing landed, so the sim is not reheated either.
    expect(restarted).toBe(false);
    expect(fakeNode.title).toBe("One");
  });
});

describe("types link helpers", () => {
  const nodes: SimulationNode[] = [
    { id: "a", title: "A" },
    { id: "b", title: "B" },
  ];
  const map = new Map(nodes.map((n) => [n.id, n]));

  it("resolves an object reference directly", () => {
    const ref = nodes[0];
    expect(resolveLinkEndpoint(ref, nodes)).toBe(ref);
  });

  it("follows a moved copy from the node map instead of the linked object", () => {
    const moved = { ...nodes[0], x: 500, y: 500 };
    expect(resolveLinkEndpoint(nodes[0], nodes, new Map([["a", moved]]))).toBe(moved);
    expect(resolveLinkEndpoint(nodes[1], nodes, new Map([["a", moved]]))).toBe(nodes[1]);
  });

  it("resolves a numeric index", () => {
    expect(resolveLinkEndpoint(1, nodes)).toBe(nodes[1]);
  });

  it("resolves a string id using a node map", () => {
    expect(resolveLinkEndpoint("b", nodes, map)).toBe(nodes[1]);
  });

  it("falls back to a linear search when no map is provided", () => {
    expect(resolveLinkEndpoint("a", nodes)).toBe(nodes[0]);
  });

  it("returns undefined for an unknown id", () => {
    expect(resolveLinkEndpoint("z", nodes, map)).toBeUndefined();
    expect(resolveLinkEndpoint("z", nodes)).toBeUndefined();
  });

  it("extracts the id from various endpoint references", () => {
    expect(getLinkEndpointId("x")).toBe("x");
    expect(getLinkEndpointId(42)).toBe("42");
    expect(getLinkEndpointId({ id: "y" })).toBe("y");
  });
});
