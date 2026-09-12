import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  drawCurvedLinkPath,
  buildBidirectionalPairSet,
  drawAnimatedLink,
  drawLink,
  drawPreviewLink,
} from "./link-renderers";
import type { SimulationLink, SimulationNode } from "./types";

function createCtx() {
  return {
    beginPath: vi.fn(),
    moveTo: vi.fn(),
    lineTo: vi.fn(),
    quadraticCurveTo: vi.fn(),
    stroke: vi.fn(),
    arc: vi.fn(),
    fill: vi.fn(),
    setLineDash: vi.fn(),
    lineDashOffset: 0,
    lineWidth: 0,
    strokeStyle: "",
    fillStyle: "",
    shadowBlur: 0,
    shadowColor: "",
  };
}

function makeLink(type: string, source: string | object, target: string | object, weight = 0.5): SimulationLink {
  return { id: "l1", source, target, link_type: type, weight } as SimulationLink;
}

function makeNode(id: string, x: number, y: number): SimulationNode {
  return { id, x, y, vx: 0, vy: 0, index: 0 } as SimulationNode;
}

describe("link renderers", () => {
  let ctx: ReturnType<typeof createCtx>;

  beforeEach(() => {
    ctx = createCtx();
  });

  it("draws a straight curved link path", () => {
    drawCurvedLinkPath(ctx as any, { x: 0, y: 0 }, { x: 10, y: 10 }, 0);
    drawCurvedLinkPath(ctx as any, { x: 0, y: 0 }, { x: 10, y: 10 }, 5);
    expect(ctx.lineTo).toHaveBeenCalled();
    expect(ctx.quadraticCurveTo).toHaveBeenCalled();
  });

  it("builds bidirectional pair sets", () => {
    const links = [
      makeLink("association", "a", "b"),
      makeLink("association", "b", "a"),
      makeLink("association", "c", "d"),
    ];
    const pairs = buildBidirectionalPairSet(links);
    expect(pairs.has("a|b")).toBe(true);
    expect(pairs.has("c|d")).toBe(false);
  });

  it("skips animated link when nodes or coordinates are missing", () => {
    const link = makeLink("association", "a", "b");
    drawAnimatedLink(ctx as any, link, new Map(), 0, 10);
    expect(ctx.beginPath).not.toHaveBeenCalled();

    const nodes = new Map<string, SimulationNode>([["a", makeNode("a", 0, 0)]]);
    drawAnimatedLink(ctx as any, link, nodes, 0, 10);
    expect(ctx.beginPath).not.toHaveBeenCalled();
  });

  it("draws a static link fallback when link count is high", () => {
    const link = makeLink("association", "a", "b");
    const nodes = new Map<string, SimulationNode>([
      ["a", makeNode("a", 0, 0)],
      ["b", makeNode("b", 10, 10)],
    ]);
    drawAnimatedLink(ctx as any, link, nodes, 0, 999, null, 0, 1, false, new Set());
    expect(ctx.beginPath).toHaveBeenCalled();
  });

  it("draws an animated link with hover and duplicate highlight", () => {
    const link = makeLink("association", "a", "b");
    const nodes = new Map<string, SimulationNode>([
      ["a", makeNode("a", 0, 0)],
      ["b", makeNode("b", 10, 10)],
    ]);
    drawAnimatedLink(ctx as any, link, nodes, 0, 10, "a", 0, 1, true, new Set(["a", "b"]));
    expect(ctx.stroke).toHaveBeenCalled();
  });

  it("draws an animated link with hovered neighbor branch", () => {
    const link = makeLink("association", "c", "d");
    const nodes = new Map<string, SimulationNode>([
      ["c", makeNode("c", 0, 0)],
      ["d", makeNode("d", 10, 10)],
    ]);
    drawAnimatedLink(ctx as any, link, nodes, 0, 10, "c", 0, 1, false, new Set(["c", "d"]));
    expect(ctx.stroke).toHaveBeenCalled();
  });

  it("draws a static link with duplicate and hover", () => {
    const link = makeLink("association", "a", "b");
    drawLink(
      ctx as any,
      link,
      makeNode("a", 0, 0),
      makeNode("b", 10, 10),
      1,
      "a",
      true,
      0,
      new Set(["a", "b"])
    );
    expect(ctx.stroke).toHaveBeenCalled();
  });

  it("draws a static link without hover", () => {
    const link = makeLink("association", "a", "b", 0);
    drawLink(ctx as any, link, makeNode("a", 0, 0), makeNode("b", 10, 10), 0.5);
    expect(ctx.stroke).toHaveBeenCalled();
  });

  it("draws a preview link", () => {
    drawPreviewLink(ctx as any, 0, 0, 10, 10);
    drawPreviewLink(ctx as any, 0, 0, 10, 10, 0.3);
    expect(ctx.stroke).toHaveBeenCalled();
    expect(ctx.fill).toHaveBeenCalled();
  });
});
