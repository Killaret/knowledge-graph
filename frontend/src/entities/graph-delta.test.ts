import { describe, it, expect } from "vitest";
import { GraphDelta } from "./graph-delta";
import type { GraphDeltaData } from "$shared/api/graph";

describe("GraphDelta", () => {
  const fullData: GraphDeltaData = {
    added_nodes: [{ id: "1", title: "A" }],
    removed_nodes: ["2"],
    updated_nodes: [{ id: "3", title: "B" }],
    added_links: [{ source: "1", target: "3" }],
    removed_links: [{ source: "2", target: "3" }],
  };

  it("builds an empty delta", () => {
    const delta = GraphDelta.empty();
    expect(delta.isEmpty()).toBe(true);
    expect(delta.totalChanges).toBe(0);
    expect(delta.requiresFullRestart()).toBe(false);
  });

  it("maps raw API data into normalized arrays", () => {
    const delta = GraphDelta.fromAPI(fullData);
    expect(delta.addedNodes).toHaveLength(1);
    expect(delta.removedNodeIds).toHaveLength(1);
    expect(delta.updatedNodes).toHaveLength(1);
    expect(delta.addedLinks).toHaveLength(1);
    expect(delta.removedLinks).toHaveLength(1);
    expect(delta.totalChanges).toBe(5);
  });

  it("detects when a full restart is required", () => {
    const small = GraphDelta.fromAPI({
      added_nodes: [{ id: "1", title: "A" }],
    });
    expect(small.requiresFullRestart()).toBe(false);

    const large = GraphDelta.fromAPI({
      added_nodes: Array.from({ length: 11 }, (_, i) => ({
        id: String(i),
        title: "Node",
      })),
    });
    expect(large.requiresFullRestart()).toBe(true);
    expect(large.requiresFullRestart(20)).toBe(false);
  });

  it("merges two deltas", () => {
    const a = GraphDelta.fromAPI({
      added_nodes: [{ id: "1", title: "A" }],
      removed_nodes: ["2"],
    });
    const b = GraphDelta.fromAPI({
      added_nodes: [{ id: "3", title: "B" }],
      removed_nodes: ["2", "4"],
    });
    const merged = a.merge(b);
    expect(merged.addedNodes).toHaveLength(2);
    expect(merged.removedNodeIds).toHaveLength(2);
    expect(merged.removedNodeIds).toContain("2");
    expect(merged.removedNodeIds).toContain("4");
    expect(merged.version).toBe(a.version + 1);
  });
});
