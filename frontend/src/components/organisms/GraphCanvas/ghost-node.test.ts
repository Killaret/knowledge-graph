/**
 * Unit tests for ghost-node module
 */
import { describe, it, expect } from "vitest";
import {
  createGhostNode,
  updateGhostNodePosition,
  updateGhostNodePulse,
  isPointOverGhostNode,
  GHOST_NODE_RADIUS,
} from "./ghost-node";
import type { SimulationNode } from "./types";

describe("ghost-node", () => {
  it("should create ghost node centered when no notes exist", () => {
    const ghostNode = createGhostNode(800, 600, []);
    expect(ghostNode.x).toBe(400);
    expect(ghostNode.y).toBe(300);
    expect(ghostNode.radius).toBe(GHOST_NODE_RADIUS);
    expect(ghostNode.active).toBe(true);
  });

  it("should create ghost node in top-left when notes exist", () => {
    const nodes: SimulationNode[] = [{ id: "1", title: "Note" }];
    const ghostNode = createGhostNode(800, 600, nodes);
    expect(ghostNode.x).toBe(60);
    expect(ghostNode.y).toBe(60);
  });

  it("should update position when nodes change", () => {
    const ghostNode = createGhostNode(800, 600, []);
    updateGhostNodePosition(ghostNode, 800, 600, [{ id: "1", title: "Note" }]);
    expect(ghostNode.x).toBe(60);
    expect(ghostNode.y).toBe(60);
  });

  it("should update pulse phase between 0 and 1", () => {
    const ghostNode = createGhostNode(800, 600, []);
    updateGhostNodePulse(ghostNode, 1000);
    expect(ghostNode.pulsePhase).toBeGreaterThanOrEqual(0);
    expect(ghostNode.pulsePhase).toBeLessThanOrEqual(1);
  });

  it("should detect point over ghost node", () => {
    const ghostNode = createGhostNode(800, 600, []);
    const transform = { x: 0, y: 0, k: 1 };
    expect(isPointOverGhostNode(400, 300, ghostNode, transform)).toBe(true);
  });

  it("should detect point over ghost node with zoom", () => {
    const ghostNode = createGhostNode(800, 600, []);
    const transform = { x: 0, y: 0, k: 2 };
    expect(isPointOverGhostNode(800, 600, ghostNode, transform)).toBe(true);
  });

  it("should not detect point outside ghost node", () => {
    const ghostNode = createGhostNode(800, 600, []);
    const transform = { x: 0, y: 0, k: 1 };
    expect(
      isPointOverGhostNode(
        400 + GHOST_NODE_RADIUS + 10,
        300,
        ghostNode,
        transform,
      ),
    ).toBe(false);
  });
});
