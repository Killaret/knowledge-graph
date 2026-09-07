import { describe, it, expect } from "vitest";
import { toSimulationNodes } from "./config";
import type { GraphNode, GraphLink } from "$shared/api/graph";

describe("toSimulationNodes", () => {
  it("preserves graph-service 3D coordinates when all nodes have x/y/z", () => {
    const nodes: GraphNode[] = [
      { id: "1", title: "A", type: "star", x: 10, y: 20, z: 30 },
      { id: "2", title: "B", type: "planet", x: -5, y: 0, z: 5 },
    ];
    const links: GraphLink[] = [];
    const result = toSimulationNodes(nodes, links);
    expect(result[0].x).toBe(10);
    expect(result[0].y).toBe(20);
    expect(result[0].z).toBe(30);
    expect(result[1].x).toBe(-5);
    expect(result[1].y).toBe(0);
    expect(result[1].z).toBe(5);
  });

  it("generates deterministic seed coordinates when positions are missing", () => {
    const nodes: GraphNode[] = [
      { id: "1", title: "A", type: "star" },
      { id: "2", title: "B", type: "planet" },
    ];
    const links: GraphLink[] = [];
    const result = toSimulationNodes(nodes, links);
    expect(typeof result[0].x).toBe("number");
    expect(typeof result[0].y).toBe("number");
    expect(typeof result[0].z).toBe("number");
    expect(result[0].x).not.toBe(0);
    expect(result[1].x).not.toBe(result[0].x);
  });
});
