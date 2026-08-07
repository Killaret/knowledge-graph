import { describe, it, expect } from "vitest";
import { drawFog, createFogVisibilitySet, defaultFogRenderParams, isNodeHiddenByFog } from "./fog";
import { createMockCanvasContext } from "./test-canvas-mock";
import type { SimulationNode, SimulationLink } from "./types";
import { getLinkEndpointId } from "./types";

function makeNodes(count: number, spread = 10): SimulationNode[] {
  return Array.from({ length: count }, (_, i) => ({
    id: `node-${i}`,
    title: `Note ${i}`,
    type: ["star", "planet"][i % 2],
    x: (i % spread) * 10,
    y: Math.floor(i / spread) * 10,
  }));
}

function makeLinks(nodes: SimulationNode[]): SimulationLink[] {
  return nodes.map((node, i) => ({
    source: node.id,
    target: nodes[(i + 1) % nodes.length].id,
    weight: 0.5,
    link_type: "related",
  }));
}

describe("2D fog rendering and culling", () => {
  it("drawFog creates a radial gradient on the canvas", () => {
    const ctx = createMockCanvasContext();
    const fog = { ...defaultFogRenderParams(), mode: "atmospheric" as const };
    drawFog(ctx, 800, 600, fog);
    expect(ctx.createRadialGradient).toHaveBeenCalled();
  });

  it("drawFog does nothing when fog mode is off", () => {
    const ctx = createMockCanvasContext();
    const fog = { ...defaultFogRenderParams(), mode: "off" as const };
    drawFog(ctx, 800, 600, fog);
    expect(ctx.createRadialGradient).not.toHaveBeenCalled();
  });

  it("createFogVisibilitySet keeps all nodes visible when fog is disabled", () => {
    const nodes = makeNodes(10);
    const links = makeLinks(nodes);
    const fog = { ...defaultFogRenderParams(), mode: "off" as const };
    const visible = createFogVisibilitySet(nodes, links, 800, 600, { x: 0, y: 0, k: 1 }, fog);
    expect(visible.size).toBe(nodes.length);
  });

  it("createFogVisibilitySet culls distant nodes when fog is active", () => {
    const nodes = makeNodes(20);
    const links = makeLinks(nodes);
    const fog = {
      ...defaultFogRenderParams(),
      mode: "adaptive" as const,
      centerX: 0,
      centerY: 0,
      radius: 50,
    };
    const visible = createFogVisibilitySet(nodes, links, 800, 600, { x: 0, y: 0, k: 1 }, fog);
    // Only nodes within 50 screen pixels from (0,0) are visible; the spread
    // is 10x10 world units, so at scale 1 only a few corner nodes qualify.
    expect(visible.size).toBeLessThan(nodes.length);
  });

  it("createFogVisibilitySet always reveals the hovered node and its neighbors", () => {
    const nodes = makeNodes(20);
    const links = makeLinks(nodes);
    const fog = {
      ...defaultFogRenderParams(),
      mode: "adaptive" as const,
      centerX: -1000,
      centerY: -1000,
      radius: 1,
    };
    const hoveredNodeId = nodes[5].id;
    const visible = createFogVisibilitySet(
      nodes,
      links,
      800,
      600,
      { x: 0, y: 0, k: 1 },
      fog,
      hoveredNodeId
    );
    // Hovered node + its linked neighbors should be in the set.
    expect(visible.has(hoveredNodeId)).toBe(true);
    const neighborIds = links
      .filter(
        (l) =>
          getLinkEndpointId(l.source) === hoveredNodeId ||
          getLinkEndpointId(l.target) === hoveredNodeId
      )
      .map((l) =>
        getLinkEndpointId(l.source) === hoveredNodeId
          ? getLinkEndpointId(l.target)
          : getLinkEndpointId(l.source)
      );
    for (const neighborId of neighborIds) {
      expect(visible.has(neighborId)).toBe(true);
    }
  });

  it("createFogVisibilitySet culls nodes outside the viewport", () => {
    const nodes = makeNodes(20, 1).map((n) => ({ ...n, x: (n.x ?? 0) * 5, y: (n.y ?? 0) * 5 }));
    // Place the first 10 nodes off-screen to the right.
    for (let i = 0; i < 10; i++) {
      nodes[i].x = 1000 + i * 50;
    }
    const links = makeLinks(nodes);
    const fog = { ...defaultFogRenderParams(), mode: "off" as const };
    const visible = createFogVisibilitySet(nodes, links, 800, 600, { x: 0, y: 0, k: 1 }, fog);
    for (let i = 0; i < 10; i++) {
      expect(visible.has(nodes[i].id)).toBe(false);
    }
    expect(visible.size).toBeLessThan(nodes.length);
  });

  it("createFogVisibilitySet returns all nodes in first-person mode", () => {
    const nodes = makeNodes(10);
    const links = makeLinks(nodes);
    const fog = { ...defaultFogRenderParams(), mode: "first-person" as const };
    const visible = createFogVisibilitySet(nodes, links, 800, 600, { x: 0, y: 0, k: 1 }, fog);
    expect(visible.size).toBe(nodes.length);
  });

  it("createFogVisibilitySet skips nodes without coordinates", () => {
    const nodes = makeNodes(5);
    nodes[0].x = undefined;
    nodes[0].y = undefined;
    const links = makeLinks(nodes);
    const fog = { ...defaultFogRenderParams(), mode: "off" as const };
    const visible = createFogVisibilitySet(nodes, links, 800, 600, { x: 0, y: 0, k: 1 }, fog);
    expect(visible.has(nodes[0].id)).toBe(false);
    expect(visible.size).toBe(nodes.length - 1);
  });

  it("createFogVisibilitySet uses provided nodeMap to look up the hovered node", () => {
    const nodes = makeNodes(3).map((n, i) => ({ ...n, x: i * 10, y: i * 10 }));
    const links = makeLinks(nodes);
    const hoveredId = nodes[0].id;

    // Build a custom nodeMap that does NOT include the hovered node.
    const nodeMap = new Map<string, SimulationNode>();
    nodeMap.set(nodes[1].id, nodes[1]);
    nodeMap.set(nodes[2].id, nodes[2]);

    // Place the fog center off-screen so no node is revealed by the fog radius itself.
    const fog = {
      ...defaultFogRenderParams(),
      mode: "adaptive" as const,
      radius: 50,
      centerX: 1000,
      centerY: 1000,
    };

    const visible = createFogVisibilitySet(
      nodes,
      links,
      800,
      600,
      { x: 0, y: 0, k: 1 },
      fog,
      hoveredId,
      nodeMap
    );

    // Because the hovered node is missing from the provided map, it cannot be revealed
    // along with its neighbors. The fog center is off-screen, so no node is visible.
    expect(visible.has(hoveredId)).toBe(false);
    expect(visible.has(nodes[1].id)).toBe(false);
    expect(visible.has(nodes[2].id)).toBe(false);
  });

  it("drawFog does nothing when fog is disabled", () => {
    const ctx = createMockCanvasContext();
    const fog = { ...defaultFogRenderParams(), enabled: false, mode: "atmospheric" as const };
    drawFog(ctx, 800, 600, fog);
    expect(ctx.createRadialGradient).not.toHaveBeenCalled();
  });

  it("drawFog does nothing when radius is zero or negative", () => {
    const ctx = createMockCanvasContext();
    const fog = { ...defaultFogRenderParams(), mode: "atmospheric" as const, radius: 0 };
    drawFog(ctx, 800, 600, fog);
    expect(ctx.createRadialGradient).not.toHaveBeenCalled();
  });

  it("drawFog does nothing in first-person mode", () => {
    const ctx = createMockCanvasContext();
    const fog = { ...defaultFogRenderParams(), mode: "first-person" as const };
    drawFog(ctx, 800, 600, fog);
    expect(ctx.createRadialGradient).not.toHaveBeenCalled();
  });

  it("drawFog fills the canvas in atmospheric mode", () => {
    const ctx = createMockCanvasContext();
    const fog = { ...defaultFogRenderParams(), mode: "atmospheric" as const };
    drawFog(ctx, 800, 600, fog);
    expect(ctx.createRadialGradient).toHaveBeenCalled();
    expect(ctx.fillRect).toHaveBeenCalledWith(0, 0, 800, 600);
  });

  it("isNodeHiddenByFog never hides nodes when fog is off", () => {
    const node = { id: "n1", title: "N", type: "star", x: 0, y: 0 };
    const fog = { ...defaultFogRenderParams(), mode: "off" as const };
    expect(isNodeHiddenByFog(fog, new Set(), node)).toBe(false);
    expect(isNodeHiddenByFog(fog, new Set([node.id]), node)).toBe(false);
  });

  it("isNodeHiddenByFog never hides nodes in first-person mode", () => {
    const node = { id: "n1", title: "N", type: "star", x: 0, y: 0 };
    const fog = { ...defaultFogRenderParams(), mode: "first-person" as const };
    expect(isNodeHiddenByFog(fog, new Set(), node)).toBe(false);
  });

  it("isNodeHiddenByFog returns true for nodes not in the visible set", () => {
    const node = { id: "n1", title: "N", type: "star", x: 0, y: 0 };
    const fog = { ...defaultFogRenderParams(), mode: "adaptive" as const };
    expect(isNodeHiddenByFog(fog, new Set(), node)).toBe(true);
    expect(isNodeHiddenByFog(fog, new Set(["n1"]), node)).toBe(false);
  });
});
