import { describe, it, expect, vi } from "vitest";
import { drawFog, createFogVisibilitySet, defaultFogRenderParams } from "./fog";
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
    const visible = createFogVisibilitySet(nodes, links, { x: 0, y: 0, k: 1 }, fog);
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
    const visible = createFogVisibilitySet(nodes, links, { x: 0, y: 0, k: 1 }, fog);
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
    const visible = createFogVisibilitySet(nodes, links, { x: 0, y: 0, k: 1 }, fog, hoveredNodeId);
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
});
