/**
 * Regression tests for the 2D graph rendering orchestrator.
 *
 * The core performance fix is: drawAllLinks must resolve link endpoints
 * through the node id Map, not by calling `nodes.find()` in a loop.
 */
import { describe, it, expect, vi } from "vitest";
import { drawAllLinks, drawAllNodes, draw, resetView } from "./renderer";
import { createMockCanvasContext } from "./test-canvas-mock";
import type { SimulationNode, SimulationLink } from "./types";

function makeNodes(count: number): SimulationNode[] {
  return Array.from({ length: count }, (_, i) => ({
    id: `node-${i}`,
    title: `Note ${i}`,
    type: ["star", "planet"][i % 2],
    x: (i % 100) * 8,
    y: Math.floor(i / 100) * 6,
  }));
}

function makeLinks(nodes: SimulationNode[]): SimulationLink[] {
  const links: SimulationLink[] = [];
  for (let i = 0; i < nodes.length; i++) {
    const target = (i + 1) % nodes.length;
    links.push({
      source: nodes[i].id,
      target: nodes[target].id,
      weight: 0.5,
      link_type: "related",
    });
    const cross = (i + Math.floor(nodes.length / 3)) % nodes.length;
    if (cross !== i && cross !== target) {
      links.push({
        source: nodes[i].id,
        target: nodes[cross].id,
        weight: 0.7,
        link_type: "related",
      });
    }
  }
  return links;
}

describe("renderer-orchestrator performance regressions", () => {
  it("drawAllLinks does not call nodes.find() for string endpoint resolution", () => {
    const nodes = makeNodes(500);
    const links = makeLinks(nodes);
    const findSpy = vi.spyOn(nodes, "find");
    const ctx = createMockCanvasContext();

    drawAllLinks(ctx, links, nodes);

    expect(findSpy).not.toHaveBeenCalled();
    findSpy.mockRestore();
  });

  it("drawAllLinks still draws all resolvable links", () => {
    const nodes = makeNodes(50);
    const links = makeLinks(nodes);
    const ctx = createMockCanvasContext();

    drawAllLinks(ctx, links, nodes);

    // Each resolvable link calls ctx.stroke at least once.
    // The number of strokes may exceed link count due to animated dash effects,
    // but it should be at least the number of links.
    const strokeMock = vi.mocked(ctx.stroke);
    expect(strokeMock.mock.calls.length).toBeGreaterThanOrEqual(links.length);
  });

  it("drawAllLinks culls links whose endpoints are outside the visible set", () => {
    const nodes = makeNodes(4).map((n, i) => ({ ...n, x: i * 100, y: i * 100 }));
    const links: SimulationLink[] = [
      { source: nodes[0].id, target: nodes[1].id, weight: 0.5, link_type: "related" },
      { source: nodes[2].id, target: nodes[3].id, weight: 0.5, link_type: "related" },
    ];
    const ctx = createMockCanvasContext();
    const visibleIds = new Set([nodes[0].id, nodes[1].id]);

    drawAllLinks(ctx, links, nodes, undefined, 0, null, null, [], new Map(), undefined, visibleIds);

    const beginPathCalls = (ctx.beginPath as any).mock.calls.length;
    // Only one link should have started a path
    expect(beginPathCalls).toBe(1);
  });

  it("drawAllLinks skips dying links when both endpoints are outside the visible set", () => {
    const nodes = makeNodes(2).map((n, i) => ({ ...n, x: i * 100, y: i * 100 }));
    const links: SimulationLink[] = [
      { source: nodes[0].id, target: nodes[1].id, weight: 0.5, link_type: "related" },
    ];
    const dyingLinks = [...links];
    const dyingLinkOpacity = new Map<string, number>([["node-0-node-1-related", 0.8]]);
    const ctx = createMockCanvasContext();
    const visibleIds = new Set<string>(); // both endpoints outside

    drawAllLinks(
      ctx,
      links,
      nodes,
      undefined,
      0,
      null,
      null,
      dyingLinks,
      dyingLinkOpacity,
      undefined,
      visibleIds
    );

    // No beginPath/stroke from the dying link
    const strokeMock = vi.mocked(ctx.stroke);
    expect(strokeMock.mock.calls.length).toBe(0);
  });

  it("drawAllNodes draws visible nodes and skips hidden ones", () => {
    const nodes = makeNodes(3).map((n, i) => ({ ...n, x: i * 50, y: i * 50 }));
    const ctx = createMockCanvasContext();
    const angles = new Map<string, number>();
    const visibleIds = new Set([nodes[0].id, nodes[2].id]);

    drawAllNodes(
      ctx,
      nodes,
      angles,
      false,
      undefined,
      false,
      0,
      null,
      null,
      false,
      undefined,
      visibleIds,
      false
    );

    const strokeCalls = (ctx.stroke as any).mock.calls;
    // node-0 and node-2 are stars and each calls ctx.stroke once.
    expect(strokeCalls.length).toBe(2);
  });

  it("drawAllNodes draws simplified circles when zoomed out", () => {
    const nodes = makeNodes(3).map((n, i) => ({ ...n, x: i * 50, y: i * 50 }));
    const ctx = createMockCanvasContext();
    const angles = new Map<string, number>();

    drawAllNodes(
      ctx,
      nodes,
      angles,
      false,
      undefined,
      false,
      0,
      null,
      null,
      false,
      undefined,
      undefined,
      true
    );

    const arcCalls = (ctx.arc as any).mock.calls;
    expect(arcCalls.length).toBe(nodes.length);
    // Titles should not be drawn in simplified mode
    expect(ctx.fillText).not.toHaveBeenCalled();
  });

  it("drawAllNodes highlights hovered node and direct neighbors at full/bright opacity", () => {
    const nodes = makeNodes(3).map((n, i) => ({ ...n, x: i * 50, y: i * 50 }));
    const links: SimulationLink[] = [
      { source: nodes[0].id, target: nodes[1].id, weight: 0.5, link_type: "related" },
    ];
    const ctx = createMockCanvasContext();
    const angles = new Map<string, number>();
    const hoveredNeighborIds = new Set([nodes[1].id]);

    drawAllNodes(
      ctx,
      nodes,
      angles,
      false,
      undefined,
      false,
      0,
      nodes[0].id,
      null,
      false,
      undefined,
      undefined,
      false,
      hoveredNeighborIds
    );

    const alphas = (ctx as any).getGlobalAlphas();
    // One globalAlpha per drawn node (set before drawNode, restored after).
    expect(alphas.length).toBeGreaterThanOrEqual(3);
    expect(alphas).toContain(1); // hovered
    expect(alphas).toContain(0.85); // neighbor
    expect(alphas).toContain(0.3); // unrelated
  });

  it("drawAllNodes draws search match outline", () => {
    const nodes = makeNodes(2).map((n, i) => ({ ...n, x: i * 50, y: i * 50 }));
    const ctx = createMockCanvasContext();
    const angles = new Map<string, number>();
    const searchMatchIds = new Set([nodes[1].id]);

    drawAllNodes(
      ctx,
      nodes,
      angles,
      false,
      undefined,
      false,
      0,
      null,
      null,
      false,
      searchMatchIds,
      undefined,
      false
    );

    const strokeStyles = ctx.getStrokeStyles();
    expect(strokeStyles).toContain("rgba(255, 204, 0, 0.9)");
  });

  it("draw clears the canvas, applies transform and draws nodes", () => {
    const nodes = makeNodes(2).map((n, i) => ({ ...n, x: i * 50, y: i * 50 }));
    const links: SimulationLink[] = [
      { source: nodes[0].id, target: nodes[1].id, weight: 0.5, link_type: "related" },
    ];
    const ctx = createMockCanvasContext();
    const angles = new Map<string, number>();
    const transform = { x: 10, y: 20, k: 1 };

    draw(ctx, 400, 300, links, nodes, angles, transform);

    expect(ctx.clearRect).toHaveBeenCalledWith(0, 0, 400, 300);
    expect(ctx.translate).toHaveBeenCalledWith(transform.x, transform.y);
    expect(ctx.scale).toHaveBeenCalledWith(transform.k, transform.k);
    // At least the two nodes should produce fill calls (star + planet)
    expect((ctx.fill as any).mock.calls.length).toBeGreaterThanOrEqual(nodes.length);
    // The star also strokes its outline.
    expect(ctx.stroke).toHaveBeenCalled();
  });

  it("draw uses simplified rendering below lod_simplify_zoom", () => {
    const nodes = makeNodes(3).map((n, i) => ({ ...n, x: i * 50, y: i * 50 }));
    const links: SimulationLink[] = [];
    const ctx = createMockCanvasContext();
    const angles = new Map<string, number>();
    const transform = { x: 0, y: 0, k: 0.2 };

    draw(ctx, 400, 300, links, nodes, angles, transform);

    // Simplified circles mean no title text
    expect(ctx.fillText).not.toHaveBeenCalled();
    // But nodes are still drawn via arc
    expect((ctx.arc as any).mock.calls.length).toBeGreaterThanOrEqual(nodes.length);
  });

  it("draw renders a link preview to a target node", () => {
    const nodes = makeNodes(2).map((n, i) => ({ ...n, x: i * 50, y: i * 50 }));
    const links: SimulationLink[] = [];
    const ctx = createMockCanvasContext();
    const angles = new Map<string, number>();
    const transform = { x: 0, y: 0, k: 1 };

    draw(
      ctx,
      400,
      300,
      links,
      nodes,
      angles,
      transform,
      undefined,
      undefined,
      undefined,
      undefined,
      false,
      undefined,
      null,
      null,
      null,
      null,
      null,
      false,
      undefined,
      undefined,
      {
        sourceId: nodes[0].id,
        targetId: nodes[1].id,
      }
    );

    expect(ctx.beginPath).toHaveBeenCalled();
    expect(ctx.stroke).toHaveBeenCalled();
    expect(ctx.moveTo).toHaveBeenCalledWith(nodes[0].x, nodes[0].y);
  });

  it("draw renders a link preview to the mouse position", () => {
    const nodes = makeNodes(1).map((n, i) => ({ ...n, x: i * 50, y: i * 50 }));
    const links: SimulationLink[] = [];
    const ctx = createMockCanvasContext();
    const angles = new Map<string, number>();
    const transform = { x: 0, y: 0, k: 1 };

    draw(
      ctx,
      400,
      300,
      links,
      nodes,
      angles,
      transform,
      undefined,
      undefined,
      undefined,
      undefined,
      false,
      undefined,
      null,
      null,
      null,
      null,
      null,
      false,
      undefined,
      undefined,
      null,
      {
        sourceId: nodes[0].id,
        x: 123,
        y: 456,
      }
    );

    expect(ctx.lineTo).toHaveBeenCalledWith(123, 456);
    expect(ctx.stroke).toHaveBeenCalled();
  });

  it("draw draws a dying link when both endpoints are visible", () => {
    const nodes = makeNodes(2).map((n, i) => ({ ...n, x: i * 50, y: i * 50 }));
    const dyingLink: SimulationLink = {
      source: nodes[0].id,
      target: nodes[1].id,
      weight: 0.5,
      link_type: "related",
    };
    const ctx = createMockCanvasContext();
    const angles = new Map<string, number>();
    const transform = { x: 0, y: 0, k: 1 };
    const dyingLinkOpacity = new Map<string, number>([["node-0-node-1-related", 0.8]]);

    draw(
      ctx,
      400,
      300,
      [],
      nodes,
      angles,
      transform,
      undefined,
      undefined,
      [dyingLink],
      dyingLinkOpacity
    );

    expect(ctx.beginPath).toHaveBeenCalled();
    expect(ctx.stroke).toHaveBeenCalled();
  });

  it("resetView centers the graph in the canvas", () => {
    const nodes = makeNodes(2).map((n, i) => ({ ...n, x: i * 100, y: i * 100 }));
    const ctx = createMockCanvasContext();
    const transform = { x: 0, y: 0, k: 1 };

    resetView(ctx, 400, 400, nodes, transform);

    expect(transform.k).toBeGreaterThan(0);
    expect(transform.x).toBeDefined();
    expect(transform.y).toBeDefined();
  });

  it("resetView does nothing when there are no nodes", () => {
    const ctx = createMockCanvasContext();
    const transform = { x: 0, y: 0, k: 1 };

    resetView(ctx, 400, 400, [], transform);

    expect(transform.k).toBe(1);
    expect(transform.x).toBe(0);
    expect(transform.y).toBe(0);
  });
});
