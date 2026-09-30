import { describe, it, expect } from "vitest";
import { computeLabeledNodeIds, FULL_LABEL_ZOOM, SMALL_GRAPH_ALL_LABELS } from "./labels";
import { graphLabelHubCount } from "$shared/config";
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
  it("labels top-degree hubs when zoomed out and nothing is highlighted", () => {
    const labeled = computeLabeledNodeIds(NODES, LINKS, ZOOMED_OUT);
    expect(labeled).toBeDefined();
    expect(labeled!.has("hub")).toBe(true);
    // All connected nodes fit under the hub cap here; isolates are never
    // labeled — they have degree 0.
    for (const id of ["leaf1", "leaf2", "leaf3", "leaf4", "a", "b", "c"]) {
      expect(labeled!.has(id)).toBe(true);
    }
    expect(labeled!.has("iso0")).toBe(false);
  });

  it("caps labels at the configured hub count on a dense graph", () => {
    // Dense fixture like the reviewer's seed: 100 nodes, ~250 links,
    // average degree ~5 — the old absolute threshold (degree >= 3) labeled
    // almost every node.
    const nodes: SimulationNode[] = Array.from({ length: 100 }, (_, i) => node(`n${i}`));
    const links: SimulationLink[] = [];
    // Ring gives every node degree 2; a random-ish web raises the average.
    for (let i = 0; i < 100; i++) {
      links.push({ source: `n${i}`, target: `n${(i + 1) % 100}` } as SimulationLink);
    }
    for (let i = 0; i < 75; i++) {
      links.push({
        source: `n${(i * 7) % 100}`,
        target: `n${(i * 13 + 3) % 100}`,
      } as SimulationLink);
    }
    // 175 links -> average degree 3.5; add another 75 for >= 5.
    for (let i = 0; i < 75; i++) {
      links.push({
        source: `n${(i * 11 + 5) % 100}`,
        target: `n${(i * 17 + 7) % 100}`,
      } as SimulationLink);
    }

    const degree = new Map<string, number>();
    for (const l of links) {
      for (const e of [l.source, l.target]) {
        const id = e as string;
        degree.set(id, (degree.get(id) ?? 0) + 1);
      }
    }
    const avg = [...degree.values()].reduce((a, b) => a + b, 0) / nodes.length;
    expect(avg).toBeGreaterThanOrEqual(5);

    const labeled = computeLabeledNodeIds(nodes, links, ZOOMED_OUT);
    expect(labeled).toBeDefined();
    expect(labeled!.size).toBeLessThanOrEqual(graphLabelHubCount);
    // The most connected node must be labeled.
    const top = [...degree.entries()].sort((a, b) => b[1] - a[1])[0][0];
    expect(labeled!.has(top)).toBe(true);
  });

  it("labels the hovered node and its direct neighbors", () => {
    const labeled = computeLabeledNodeIds(NODES, LINKS, { ...ZOOMED_OUT, hoveredId: "leaf1" });
    expect(labeled!.has("leaf1")).toBe(true);
    expect(labeled!.has("hub")).toBe(true); // neighbor of leaf1
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
