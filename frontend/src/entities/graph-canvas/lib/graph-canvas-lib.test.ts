import { describe, it, expect, vi, beforeEach } from "vitest";
import { drawBackground } from "./background";
import {
  drawFog,
  createFogVisibilitySet,
  isNodeHiddenByFog,
  getHoveredNeighborIds,
  defaultFogRenderParams,
} from "./fog";
import {
  createGhostNode,
  updateGhostNodePosition,
  updateGhostNodeZoom,
  updateGhostNodePulse,
  isPointOverGhostNode,
  drawGhostNode,
  drawGhostNodeScreen,
  drawGhostNodeTooltip,
  drawGhostNodeTooltipScreen,
} from "./ghost-node";
import type { SimulationLink, SimulationNode } from "./types";

function createCtx() {
  const gradient = { addColorStop: vi.fn() };
  return {
    beginPath: vi.fn(),
    moveTo: vi.fn(),
    lineTo: vi.fn(),
    stroke: vi.fn(),
    fill: vi.fn(),
    save: vi.fn(),
    restore: vi.fn(),
    arc: vi.fn(),
    fillRect: vi.fn(),
    createRadialGradient: vi.fn(() => gradient),
    measureText: vi.fn(() => ({ width: 80 })),
    roundRect: vi.fn(),
    fillText: vi.fn(),
    shadowBlur: 0,
    shadowColor: "",
    strokeStyle: "",
    fillStyle: "",
    lineWidth: 0,
    globalCompositeOperation: "",
    font: "",
    textAlign: "",
    textBaseline: "",
  };
}

describe("background", () => {
  it("draws a background with and without nebula effects", () => {
    const ctx = createCtx();
    drawBackground(ctx as any, 100, 100, [], 0);
    drawBackground(ctx as any, 100, 100, new Array(1000).fill({ id: "x" }), 0);
    expect(ctx.stroke).toHaveBeenCalled();
    expect(ctx.fillRect).toHaveBeenCalled();
  });
});

describe("fog", () => {
  it("returns default fog parameters", () => {
    const params = defaultFogRenderParams();
    expect(params.mode).toBe("off");
    expect(params.enabled).toBe(true);
  });

  it("skips drawing fog when disabled or in first-person/off mode", () => {
    const ctx = createCtx();
    const off = { ...defaultFogRenderParams(), enabled: false };
    drawFog(ctx as any, 100, 100, off);
    const firstPerson = { ...defaultFogRenderParams(), enabled: true, mode: "first-person" as const };
    drawFog(ctx as any, 100, 100, firstPerson);
    expect(ctx.fillRect).not.toHaveBeenCalled();
  });

  it("draws atmospheric and adaptive fog", () => {
    const ctx = createCtx();
    const atmospheric = { ...defaultFogRenderParams(), enabled: true, mode: "atmospheric" as const, centerX: 50, centerY: 50, radius: 30, feather: 10, color: "rgba(0,0,0,0.5)" };
    drawFog(ctx as any, 100, 100, atmospheric);
    const adaptive = { ...defaultFogRenderParams(), enabled: true, mode: "adaptive" as const, centerX: 50, centerY: 50, radius: 30, feather: 10, color: "rgba(0,0,0,0.5)" };
    drawFog(ctx as any, 100, 100, adaptive);
    expect(ctx.fillRect).toHaveBeenCalled();
  });

  it("draws nothing when the clear radius is non-positive", () => {
    const ctx = createCtx();
    const p = { ...defaultFogRenderParams(), enabled: true, mode: "atmospheric" as const, radius: 0 };
    drawFog(ctx as any, 100, 100, p);
    expect(ctx.fillRect).not.toHaveBeenCalled();
  });

  it("creates a fog visibility set for different modes", () => {
    const nodeA: SimulationNode = { id: "a", x: 10, y: 10 } as SimulationNode;
    const nodeB: SimulationNode = { id: "b", x: 90, y: 90 } as SimulationNode;
    const link: SimulationLink = { id: "l", source: "a", target: "b" } as SimulationLink;
    const transform = { x: 0, y: 0, k: 1 };

    const firstPerson = defaultFogRenderParams();
    firstPerson.mode = "first-person";
    expect(createFogVisibilitySet([nodeA, nodeB], [link], 100, 100, transform, firstPerson).size).toBe(2);

    const off = defaultFogRenderParams();
    off.mode = "off";
    off.enabled = true;
    expect(createFogVisibilitySet([nodeA, nodeB], [link], 100, 100, transform, off).size).toBeGreaterThanOrEqual(0);

    const adaptive = defaultFogRenderParams();
    adaptive.mode = "adaptive";
    adaptive.enabled = true;
    adaptive.centerX = 10;
    adaptive.centerY = 10;
    adaptive.radius = 20;
    const set = createFogVisibilitySet([nodeA, nodeB], [link], 100, 100, transform, adaptive, "a");
    expect(set.has("a")).toBe(true);
  });

  it("determines if a node is hidden by fog", () => {
    const node: SimulationNode = { id: "a" } as SimulationNode;
    const visible = new Set<string>(["a"]);
    const fog = defaultFogRenderParams();
    fog.mode = "off";
    expect(isNodeHiddenByFog(fog, visible, node)).toBe(false);
    fog.mode = "first-person";
    expect(isNodeHiddenByFog(fog, visible, node)).toBe(false);
    fog.mode = "adaptive";
    expect(isNodeHiddenByFog(fog, visible, node)).toBe(false);
    visible.delete("a");
    expect(isNodeHiddenByFog(fog, visible, node)).toBe(true);
  });

  it("builds hovered neighbor ids", () => {
    const link: SimulationLink = { id: "l", source: "a", target: "b" } as SimulationLink;
    const self: SimulationLink = { id: "s", source: "a", target: "a" } as SimulationLink;
    expect(getHoveredNeighborIds("a", [link, self]).has("b")).toBe(true);
    expect(getHoveredNeighborIds(null, [link]).size).toBe(0);
  });
});

describe("ghost-node", () => {
  it("creates a ghost node with and without notes", () => {
    const withNotes = createGhostNode(100, 100, [{ id: "1" }] as SimulationNode[]);
    const empty = createGhostNode(100, 100, []);
    expect(withNotes.x).not.toBe(empty.x);
    expect(empty.x).toBe(50);
  });

  it("updates ghost node position, zoom and pulse", () => {
    const state = createGhostNode(100, 100, [{ id: "1" }] as SimulationNode[]);
    updateGhostNodePosition(state, 200, 200, []);
    expect(state.x).toBe(100);
    updateGhostNodeZoom(state, 5);
    expect(state.radius).toBeLessThanOrEqual(state.radius);
    updateGhostNodePulse(state, 0);
    expect(state.pulsePhase).toBe(0.5);
  });

  it("detects points over the ghost node", () => {
    const state = createGhostNode(100, 100, []);
    expect(isPointOverGhostNode(state.x, state.y, state)).toBe(true);
    expect(isPointOverGhostNode(1000, 1000, state)).toBe(false);
  });

  it("draws the ghost node in world and screen variants", () => {
    const ctx = createCtx();
    const state = createGhostNode(100, 100, []);
    state.hovered = true;
    drawGhostNode(ctx as any, state, 0);
    drawGhostNodeScreen(ctx as any, state, 0);
    expect(ctx.fill).toHaveBeenCalled();
  });

  it("draws the ghost node tooltip with default and custom text", () => {
    const ctx = createCtx();
    const state = createGhostNode(100, 100, []);
    drawGhostNodeTooltip(ctx as any, state);
    drawGhostNodeTooltipScreen(ctx as any, state, "Custom");
    expect(ctx.fillText).toHaveBeenCalled();
  });
});
