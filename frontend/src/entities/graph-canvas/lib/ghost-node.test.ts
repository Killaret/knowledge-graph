import { describe, it, expect, vi } from "vitest";
import {
  createGhostNode,
  updateGhostNodePosition,
  updateGhostNodePulse,
  isPointOverGhostNode,
  drawGhostNode,
  drawGhostNodeScreen,
  drawGhostNodeTooltip,
  drawGhostNodeTooltipScreen,
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
    expect(ghost.x).toBe(60);
    expect(ghost.y).toBe(60);
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
    expect(ghost.x).toBe(60);
    expect(ghost.y).toBe(60);
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
    const transform = { x: 0, y: 0, k: 1 };
    expect(isPointOverGhostNode(105, 105, ghost, transform)).toBe(true);
    expect(isPointOverGhostNode(500, 500, ghost, transform)).toBe(false);
  });

  it("accounts for pan/zoom when hit-testing", () => {
    const ghost = {
      x: 50,
      y: 50,
      radius: 20,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    const transform = { x: 10, y: 10, k: 2 };
    // world point (50,50) => screen point = 50*2 + 10 = 110
    expect(isPointOverGhostNode(110, 110, ghost, transform)).toBe(true);
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
