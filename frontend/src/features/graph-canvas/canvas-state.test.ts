import { describe, it, expect } from "vitest";
import { isTechnicalNode, pinTechnicalNodes } from "./canvas-state.svelte";

describe("canvas-state helpers", () => {
  it("isTechnicalNode detects technical node types", () => {
    const nodes = [
      { id: "n1", type: "star" },
      { id: "n2", type: "technical" },
    ];
    expect(isTechnicalNode(nodes, "n1")).toBe(false);
    expect(isTechnicalNode(nodes, "n2")).toBe(true);
    expect(isTechnicalNode(nodes, "missing")).toBe(false);
  });

  it("pinTechnicalNodes fixes position for technical nodes", () => {
    const nodes = [
      { id: "n1", title: "Star", type: "star" },
      { id: "n2", title: "Tech", type: "technical" },
    ];
    const result = pinTechnicalNodes(nodes);
    expect(result[1].x).toBe(60);
    expect(result[1].y).toBe(60);
    expect(result[1].fx).toBe(60);
    expect(result[1].fy).toBe(60);

    expect(result[1].id).toBe("n2");
    expect(result[0].x).toBeUndefined();
  });
});
