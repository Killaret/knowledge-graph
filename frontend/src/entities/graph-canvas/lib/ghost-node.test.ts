import { describe, it, expect, vi } from "vitest";
import {
  createGhostNode,
  updateGhostNodePosition,
  updateGhostNodePulse,
  updateGhostNodeZoom,
  isPointOverGhostNode,
  drawGhostNode,
  drawGhostNodeScreen,
  drawGhostNodeTooltip,
  drawGhostNodeTooltipScreen,
  GHOST_NODE_RADIUS,
} from "./ghost-node";
import type { SimulationNode } from "./types";

const makeCtx = () => {
  const gradient = { addColorStop: vi.fn() };
  return {
    save: vi.fn(),
    restore: vi.fn(),
    createRadialGradient: vi.fn(() => gradient),
    beginPath: vi.fn(),
    arc: vi.fn(),
    fill: vi.fn(),
    stroke: vi.fn(),
    moveTo: vi.fn(),
    lineTo: vi.fn(),
    roundRect: vi.fn(),
    measureText: vi.fn(() => ({ width: 100 })),
    fillText: vi.fn(),
    shadowBlur: 0,
    shadowColor: "",
    fillStyle: "",
    strokeStyle: "",
    lineWidth: 0,
    textAlign: "",
    textBaseline: "",
  } as unknown as CanvasRenderingContext2D;
};

describe("ghost-node", () => {
  it("creates a ghost node in top-left when notes exist", () => {
    const nodes: SimulationNode[] = [{ id: "n1", title: "n1" }];
    const ghost = createGhostNode(800, 600, nodes);
    expect(ghost.x).toBe(80);
    expect(ghost.y).toBe(80);
    expect(ghost.active).toBe(true);
    expect(ghost.radius).toBeGreaterThan(0);
  });

  it("creates a ghost node centered when graph is empty", () => {
    const ghost = createGhostNode(800, 600, []);
    expect(ghost.x).toBe(400);
    expect(ghost.y).toBe(300);
  });

  it("updates position based on current nodes", () => {
    const ghost = createGhostNode(800, 600, []);
    updateGhostNodePosition(ghost, 800, 600, [{ id: "n1", title: "n1" }]);
    expect(ghost.x).toBe(80);
    expect(ghost.y).toBe(80);
  });

  it("scales the ghost node radius with zoom", () => {
    const ghost = createGhostNode(800, 600, []);
    expect(ghost.radius).toBe(GHOST_NODE_RADIUS);
    updateGhostNodeZoom(ghost, 2);
    expect(ghost.radius).toBe(GHOST_NODE_RADIUS * 2);
    updateGhostNodeZoom(ghost, 0.2);
    expect(ghost.radius).toBe(GHOST_NODE_RADIUS * 0.4);
    updateGhostNodeZoom(ghost, 5);
    expect(ghost.radius).toBe(GHOST_NODE_RADIUS * 3);
  });

  it("hit-tests using the zoomed ghost radius", () => {
    const ghost = createGhostNode(800, 600, []);
    updateGhostNodeZoom(ghost, 2);
    expect(isPointOverGhostNode(ghost.x + ghost.radius - 1, ghost.y, ghost)).toBe(true);
    expect(isPointOverGhostNode(ghost.x + ghost.radius + 1, ghost.y, ghost)).toBe(false);
  });

  it("updates pulse phase based on time", () => {
    const ghost = createGhostNode(800, 600, []);
    updateGhostNodePulse(ghost, 1000);
    expect(ghost.pulsePhase).toBeGreaterThanOrEqual(0);
    expect(ghost.pulsePhase).toBeLessThanOrEqual(1);
  });

  it("detects when a point is over the ghost node", () => {
    const ghost = {
      x: 100,
      y: 100,
      radius: 20,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    expect(isPointOverGhostNode(105, 105, ghost)).toBe(true);
    expect(isPointOverGhostNode(500, 500, ghost)).toBe(false);
  });

  it("hit-tests in screen space", () => {
    const ghost = {
      x: 100,
      y: 100,
      radius: 20,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    // The ghost is drawn in screen coordinates, so the supplied point
    // is already in screen space and no pan/zoom transform is needed.
    expect(isPointOverGhostNode(115, 100, ghost)).toBe(true);
    expect(isPointOverGhostNode(200, 200, ghost)).toBe(false);
  });

  it("draws the world-space ghost node", () => {
    const ctx = makeCtx();
    const ghost = createGhostNode(800, 600, []);
    drawGhostNode(ctx, ghost, 0);
    expect(ctx.save).toHaveBeenCalled();
    expect(ctx.createRadialGradient).toHaveBeenCalled();
    expect(ctx.arc).toHaveBeenCalled();
    expect(ctx.fill).toHaveBeenCalled();
    expect(ctx.stroke).toHaveBeenCalled();
    expect(ctx.restore).toHaveBeenCalled();
  });

  it("draws the screen-space ghost node", () => {
    const ctx = makeCtx();
    const ghost = createGhostNode(800, 600, []);
    drawGhostNodeScreen(ctx, ghost, 0);
    expect(ctx.save).toHaveBeenCalled();
    expect(ctx.arc).toHaveBeenCalled();
    expect(ctx.restore).toHaveBeenCalled();
  });

  it("draws the tooltip in world space", () => {
    const ctx = makeCtx();
    const ghost = createGhostNode(800, 600, []);
    drawGhostNodeTooltip(ctx, ghost, "Add note");
    expect(ctx.roundRect).toHaveBeenCalled();
    expect(ctx.fillText).toHaveBeenCalledWith("Add note", expect.any(Number), expect.any(Number));
    expect(ctx.restore).toHaveBeenCalled();
  });

  it("draws the tooltip in screen space", () => {
    const ctx = makeCtx();
    const ghost = createGhostNode(800, 600, []);
    drawGhostNodeTooltipScreen(ctx, ghost, "New star");
    expect(ctx.roundRect).toHaveBeenCalled();
    expect(ctx.fillText).toHaveBeenCalledWith("New star", expect.any(Number), expect.any(Number));
    expect(ctx.restore).toHaveBeenCalled();
  });
});
