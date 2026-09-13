import { describe, it, expect } from "vitest";
import { isInsideEdge, dragDistance, getPanelOffset } from "./panel-geometry";

const fullRect = { width: 100, height: 200 };

describe("panel-geometry", () => {
  describe("isInsideEdge", () => {
    it("detects pointer inside the top edge", () => {
      expect(isInsideEdge("top", { x: 50, y: 5 }, 10, fullRect)).toBe(true);
    });

    it("rejects pointer outside the top edge band", () => {
      expect(isInsideEdge("top", { x: 50, y: 20 }, 10, fullRect)).toBe(false);
    });

    it("rejects pointer outside the top edge horizontally", () => {
      expect(isInsideEdge("top", { x: -1, y: 5 }, 10, fullRect)).toBe(false);
      expect(isInsideEdge("top", { x: 101, y: 5 }, 10, fullRect)).toBe(false);
    });

    it("detects pointer inside the bottom edge", () => {
      expect(isInsideEdge("bottom", { x: 50, y: 195 }, 10, fullRect)).toBe(true);
    });

    it("rejects pointer above the bottom edge band", () => {
      expect(isInsideEdge("bottom", { x: 50, y: 180 }, 10, fullRect)).toBe(false);
    });

    it("detects pointer inside the left edge", () => {
      expect(isInsideEdge("left", { x: 5, y: 100 }, 10, fullRect)).toBe(true);
    });

    it("rejects pointer outside the left edge horizontally", () => {
      expect(isInsideEdge("left", { x: 20, y: 100 }, 10, fullRect)).toBe(false);
    });

    it("detects pointer inside the right edge", () => {
      expect(isInsideEdge("right", { x: 95, y: 100 }, 10, fullRect)).toBe(true);
    });

    it("rejects pointer left of the right edge band", () => {
      expect(isInsideEdge("right", { x: 80, y: 100 }, 10, fullRect)).toBe(false);
    });

    it("rejects an unknown panel position", () => {
      expect(isInsideEdge("unknown" as any, { x: 50, y: 50 }, 10, fullRect)).toBe(false);
    });

    it("handles zero and full rect edges", () => {
      const zeroRect = { width: 0, height: 0 };
      expect(isInsideEdge("top", { x: 0, y: 0 }, 0, zeroRect)).toBe(true);
      expect(isInsideEdge("right", { x: 0, y: 0 }, 10, zeroRect)).toBe(true);
    });
  });

  describe("dragDistance", () => {
    it("returns positive distance when dragging top panel down", () => {
      expect(dragDistance("top", { x: 0, y: 0 }, { x: 0, y: 30 }, fullRect)).toBe(30);
    });

    it("returns positive distance when dragging bottom panel up", () => {
      expect(dragDistance("bottom", { x: 0, y: 200 }, { x: 0, y: 170 }, fullRect)).toBe(30);
    });

    it("returns positive distance when dragging left panel right", () => {
      expect(dragDistance("left", { x: 0, y: 0 }, { x: 30, y: 0 }, fullRect)).toBe(30);
    });

    it("returns positive distance when dragging right panel left", () => {
      expect(dragDistance("right", { x: 100, y: 0 }, { x: 70, y: 0 }, fullRect)).toBe(30);
    });

    it("returns negative distance for closing drag", () => {
      expect(dragDistance("top", { x: 0, y: 30 }, { x: 0, y: 0 }, fullRect)).toBe(-30);
    });

    it("returns 0 for an unknown panel position", () => {
      expect(dragDistance("unknown" as any, { x: 0, y: 0 }, { x: 10, y: 10 }, fullRect)).toBe(0);
    });
  });

  describe("getPanelOffset", () => {
    it("returns no transform when the panel is open", () => {
      expect(getPanelOffset("left", 250, true)).toEqual({ transform: "translate(0, 0)" });
      expect(getPanelOffset("right", 250, true)).toEqual({ transform: "translate(0, 0)" });
    });

    it("offsets the top panel up when closed", () => {
      expect(getPanelOffset("top", 250, false)).toEqual({ transform: "translateY(-250px)" });
    });

    it("offsets the bottom panel down when closed", () => {
      expect(getPanelOffset("bottom", 250, false)).toEqual({ transform: "translateY(250px)" });
    });

    it("offsets the left panel left when closed", () => {
      expect(getPanelOffset("left", 250, false)).toEqual({ transform: "translateX(-250px)" });
    });

    it("offsets the right panel right when closed", () => {
      expect(getPanelOffset("right", 250, false)).toEqual({ transform: "translateX(250px)" });
    });
  });
});
