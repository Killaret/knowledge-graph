/**
 * GRAPH-LIGHT-1, decision 81: recommendations drawn on hover, selection ring.
 */
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { createMockCanvasContext } from "../test-canvas-mock";
import { draw, drawAllNodes } from "../renderer";
import type { SimulationNode } from "../types";
import {
  beginLightFrame,
  getGraphStyle,
  setGraphStyle,
  setLightFocusMix,
  setLightRecommendations,
  setLightSelection,
} from "./style";
import {
  activeRecommendationIds,
  drawLightRecommendations,
  recommendationDash,
} from "./recommendations";

const initialStyle = getGraphStyle();
const REC_STROKE = "rgba(223,231,255,";

const nodes: SimulationNode[] = [
  { id: "h", title: "Hovered", type: "star", x: 0, y: 0 },
  { id: "near", title: "Linked neighbour", type: "planet", x: 60, y: 0 },
  { id: "rec", title: "Recommended", type: "planet", x: 0, y: 80 },
  { id: "far", title: "Unrelated", type: "moon", x: 200, y: 200 },
];
const nodeMap = new Map(nodes.map((n) => [n.id, n]));

beforeEach(() => {
  setGraphStyle("light");
  beginLightFrame(1, 0, false);
  setLightFocusMix(1);
  setLightRecommendations("h", [
    { id: "rec", score: 0.9 },
    { id: "near", score: 0.8 },
    { id: "gone", score: 0.7 },
  ]);
});

afterEach(() => {
  setLightRecommendations(null, []);
  setLightSelection(null);
  setGraphStyle(initialStyle);
});

describe("recommendations on hover", () => {
  it("draws only recommendations of the hovered note that are on the graph and not yet linked", () => {
    const ctx = createMockCanvasContext();
    const drawn = drawLightRecommendations(ctx, nodes[0], nodeMap, new Set(["near"]));
    expect(drawn).toBe(1);
    expect(ctx.getStrokeStyles().some((s) => String(s).startsWith(REC_STROKE))).toBe(true);
  });

  it("uses denser dashes for a closer match", () => {
    const [, strongGap] = recommendationDash(0.9);
    const [, weakGap] = recommendationDash(0.4);
    expect(strongGap).toBeLessThan(weakGap);
  });

  it("ignores recommendations loaded for another note", () => {
    setLightRecommendations("far", [{ id: "rec", score: 0.9 }]);
    expect(activeRecommendationIds("h").size).toBe(0);
    expect(drawLightRecommendations(createMockCanvasContext(), nodes[0], nodeMap)).toBe(0);
  });

  it("draws them while hovering and never at rest", () => {
    const hovered = createMockCanvasContext();
    draw(
      hovered,
      400,
      300,
      [],
      nodes,
      new Map(),
      { x: 0, y: 0, k: 1 },
      undefined,
      undefined,
      [],
      new Map(),
      false,
      0,
      "h"
    );
    expect(hovered.getStrokeStyles().some((s) => String(s).startsWith(REC_STROKE))).toBe(true);

    const rest = createMockCanvasContext();
    draw(rest, 400, 300, [], nodes, new Map(), { x: 0, y: 0, k: 1 });
    expect(rest.getStrokeStyles().some((s) => String(s).startsWith(REC_STROKE))).toBe(false);
  });

  it("keeps recommended notes lit while the rest dims", () => {
    const ctx = createMockCanvasContext();
    drawAllNodes(
      ctx,
      nodes,
      new Map(),
      false,
      undefined,
      false,
      0,
      "h",
      null,
      false,
      undefined,
      undefined,
      false,
      new Set(["near"])
    );
    const alphas = ctx.getGlobalAlphas();
    expect(alphas).toContain(0.8);
    expect(alphas).toContain(0.15);
  });
});

describe("captions of recommendations", () => {
  it("captions recommended notes even when the label rule would not", () => {
    const ctx = createMockCanvasContext();
    (ctx as unknown as { getTransform: () => DOMMatrix }).getTransform = () =>
      ({ a: 1, b: 0, c: 0, d: 1, e: 0, f: 0 }) as DOMMatrix;
    drawAllNodes(
      ctx,
      nodes,
      new Map(),
      false,
      undefined,
      false,
      0,
      "h",
      null,
      false,
      undefined,
      undefined,
      false,
      new Set(["near"]),
      new Set(["h"])
    );
    const texts = (ctx.fillText as unknown as { mock: { calls: unknown[][] } }).mock.calls.map(
      (c) => c[0]
    );
    expect(texts).toContain("Recommended");
    expect(texts).not.toContain("Unrelated");
  });
});

describe("selection ring", () => {
  it("rings the selected note", () => {
    setLightSelection("far");
    const ctx = createMockCanvasContext();
    drawAllNodes(ctx, nodes, new Map(), false);
    const ringStrokes = ctx
      .getStrokeStyles()
      .filter((s) => String(s).startsWith("rgba(220,227,240,"));
    expect(ringStrokes.length).toBeGreaterThanOrEqual(1);
  });
});
