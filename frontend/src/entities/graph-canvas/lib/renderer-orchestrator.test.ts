/**
 * Regression tests for the 2D graph rendering orchestrator.
 *
 * The core performance fix is: drawAllLinks must resolve link endpoints
 * through the node id Map, not by calling `nodes.find()` in a loop.
 */
import { describe, it, expect, vi } from "vitest";
import { drawAllLinks } from "./renderer";
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
});
