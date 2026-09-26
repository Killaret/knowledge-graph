import { describe, it, expect } from "vitest";
import { computeLabeledNodeIds, FULL_LABEL_ZOOM, SMALL_GRAPH_ALL_LABELS } from "./labels";
import type { SimulationLink, SimulationNode } from "./types";

function node(id: string): SimulationNode {
  return { id, x: 0, y: 0 } as SimulationNode;
}

// Graph above the small-graph threshold: hub -> leaf1..leaf4 (degree 4),
// a chain a-b-c, and a tail of isolated nodes.
const LINKS = [
  { source: "hub", target: "leaf1" },
  { source: "hub", target: "leaf2" },
  { source: "hub", target: "leaf3" },
  { source: "hub", target: "leaf4" },
  { source: "a", target: "b" },
  { source: "b", target: "c" },
] as SimulationLink[];
const NODES: SimulationNode[] = [
  node("hub"),
  node("leaf1"),
  node("leaf2"),
  node("leaf3"),
  node("leaf4"),
  node("a"),
  node("b"),
  node("c"),
  ...Array.from({ length: SMALL_GRAPH_ALL_LABELS - 4 }, (_, i) => node(`iso${i}`)),
];

const ZOOMED_OUT = { zoomK: 0.5 };

describe("computeLabeledNodeIds (UI-GRAPH-1: selective labels)", () => {
  it("labels only hub nodes when zoomed out and nothing is highlighted", () => {
    const labeled = computeLabeledNodeIds(NODES, LINKS, ZOOMED_OUT);
    expect(labeled).toBeDefined();
    expect(labeled!.has("hub")).toBe(true);
    expect(labeled!.has("a")).toBe(false); // degree 1 — below the hub threshold
    expect(labeled!.has("leaf1")).toBe(false);
    expect(labeled!.has("iso0")).toBe(false);
  });

  it("labels the hovered node and its direct neighbors", () => {
    const labeled = computeLabeledNodeIds(NODES, LINKS, { ...ZOOMED_OUT, hoveredId: "leaf1" });
    expect(labeled!.has("leaf1")).toBe(true);
    expect(labeled!.has("hub")).toBe(true); // neighbor of leaf1
    expect(labeled!.has("leaf2")).toBe(false); // not a neighbor of leaf1
  });

  it("always labels the selected node even if it is isolated", () => {
    const labeled = computeLabeledNodeIds(NODES, LINKS, { ...ZOOMED_OUT, selectedId: "iso0" });
    expect(labeled!.has("iso0")).toBe(true);
  });

  it("always labels search matches", () => {
    const labeled = computeLabeledNodeIds(NODES, LINKS, {
      ...ZOOMED_OUT,
      searchMatchIds: new Set(["iso1", "leaf2"]),
    });
    expect(labeled!.has("iso1")).toBe(true);
    expect(labeled!.has("leaf2")).toBe(true);
  });

  it("returns undefined at high zoom — every node gets a label", () => {
    expect(computeLabeledNodeIds(NODES, LINKS, { zoomK: FULL_LABEL_ZOOM })).toBeUndefined();
    expect(computeLabeledNodeIds(NODES, LINKS, { zoomK: FULL_LABEL_ZOOM + 1 })).toBeUndefined();
  });

  it("returns undefined for small graphs — every node gets a label", () => {
    const few = [node("x"), node("y")];
    expect(computeLabeledNodeIds(few, LINKS, ZOOMED_OUT)).toBeUndefined();
  });
});
