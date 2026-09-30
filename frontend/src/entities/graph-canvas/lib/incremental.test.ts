import { describe, it, expect, vi, beforeEach } from "vitest";
import { addNodesToSimulation } from "./incremental";
import { startSimulation } from "./simulation";
import type { SimulationNode, SimulationLink, SimulationState, TransformState } from "./types";
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

function startRunning(state: SimulationState, nodes: SimulationNode[], links: SimulationLink[]) {
  startSimulation(
    nodes,
    links,
    800,
    600,
    state,
    { x: 0, y: 0, k: 1 } as TransformState,
    vi.fn(),
    vi.fn(),
    vi.fn()
  );
}

beforeEach(() => {
  resetMockState();
  // The fade loop uses requestAnimationFrame; keep it inert in this suite.
  vi.stubGlobal("requestAnimationFrame", () => 1);
  vi.stubGlobal("cancelAnimationFrame", vi.fn());
});

describe("addNodesToSimulation", () => {
  it("appends nodes and links to the live simulation without rebuilding it", () => {
    const state = createState();
    startRunning(
      state,
      [
        { id: "a", title: "A" },
        { id: "b", title: "B" },
      ],
      [{ source: "a", target: "b" }]
    );

    const sim = state.simulation!;
    const nodeCountBefore = sim.nodes().length;

    const added = addNodesToSimulation(
      state,
      [{ id: "c", title: "C" }],
      [{ source: "b", target: "c" }],
      800,
      600
    );

    expect(added).toBe(true);
    expect(state.simulation).toBe(sim); // same instance — no restart
    expect(sim.nodes().length).toBe(nodeCountBefore + 1);
    expect(sim.nodes().some((n) => n.id === "c")).toBe(true);
    expect(state.simLinks.some((l) => getSourceId(l) === "b" && getTargetId(l) === "c")).toBe(true);
    // link force received the new edge
    const linkForce = sim.force("link") as unknown as { links(): SimulationLink[] };
    expect(linkForce.links().length).toBe(2);
  });

  it("reheats instead of restarting: alpha is lowered, simulation kept running", () => {
    const state = createState();
    startRunning(state, [{ id: "a", title: "A" }], []);
    const sim = state.simulation!;
    const alphaSpy = vi.spyOn(sim, "alpha");
    const restartSpy = vi.spyOn(sim, "restart");

    addNodesToSimulation(state, [{ id: "c", title: "C" }], [], 800, 600);

    expect(alphaSpy).toHaveBeenCalledWith(0.4);
    expect(restartSpy).toHaveBeenCalled();
    expect(state.isRunning).toBe(true);
    expect(state.stable).toBe(false);
  });

  it("spawns a linked node next to its already-placed neighbor", () => {
    const state = createState();
    startRunning(state, [{ id: "a", title: "A" }], []);
    const anchor = state.simulation!.nodes()[0];

    addNodesToSimulation(
      state,
      [{ id: "c", title: "C" }],
      [{ source: "a", target: "c" }],
      800,
      600
    );

    const c = state.simulation!.nodes().find((n) => n.id === "c")!;
    expect(Math.hypot(c.x! - anchor.x!, c.y! - anchor.y!)).toBeLessThan(200);
  });

  it("marks new nodes/links transparent so the fade loop brings them in", () => {
    const state = createState();
    startRunning(state, [{ id: "a", title: "A" }], []);

    addNodesToSimulation(
      state,
      [{ id: "c", title: "C" }],
      [{ source: "a", target: "c" }],
      800,
      600
    );

    expect(state.nodeOpacity.get("c")).toBe(0);
    expect([...state.linkOpacity.values()]).toContain(0);
  });

  it("returns false when there is no live simulation", () => {
    const state = createState();
    expect(addNodesToSimulation(state, [{ id: "c", title: "C" }], [], 800, 600)).toBe(false);
  });
});

function getSourceId(l: SimulationLink): string {
  return typeof l.source === "object" ? (l.source as SimulationNode).id : String(l.source);
}
function getTargetId(l: SimulationLink): string {
  return typeof l.target === "object" ? (l.target as SimulationNode).id : String(l.target);
}
