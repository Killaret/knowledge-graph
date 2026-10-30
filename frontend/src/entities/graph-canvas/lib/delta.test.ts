import { describe, it, expect, vi, beforeEach } from "vitest";
import { applyDelta } from "./delta";
import type { SimulationNode, SimulationLink, SimulationState, TransformState } from "./types";
import { resolveLinkEndpoint, getLinkEndpointId } from "./types";
import { resetMockState } from "../../../__mocks__/d3-force";

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

  it("performs a full restart for added/removed nodes and links", () => {
    const state = createState();
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

    expect(restarted).toBe(true);
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
