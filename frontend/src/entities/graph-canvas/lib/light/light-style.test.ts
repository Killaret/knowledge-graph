/**
 * GRAPH-LIGHT-1: the light style of the 2D graph (decision 83).
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { CelestialBody } from "$entities";
import { createMockCanvasContext } from "../test-canvas-mock";
import { drawAllNodes, draw } from "../renderer";
import { ensureCelestialBodyDrawers } from "../node-registration";
import type { SimulationLink, SimulationNode } from "../types";
import {
  beginLightFrame,
  getGraphStyle,
  lightAmbient,
  lightFrame,
  setGraphStyle,
  setLightAmbient,
  setLightFocusMix,
  setLightThreadFade,
} from "./style";
import { drawLightBackground, paintLightSky } from "./background";
import { spriteColor } from "./sprites";
import { hexToRgb, lightCoreRadius, lightTypeColor, mixRgb, rgba } from "./palette";
import { drawLightLink, lightThreadAlpha, LIGHT_DIMMED_THREAD_ALPHA } from "./threads";
import { captionPriority, drawLightCaptions } from "./labels";

const initialStyle = getGraphStyle();

/** Colours of the light stamped from sprites (halos, spheres), as rgba() strings. */
function stampedColors(ctx: ReturnType<typeof createMockCanvasContext>): string[] {
  const calls = (ctx.drawImage as unknown as ReturnType<typeof vi.fn>).mock.calls;
  return calls
    .map((call) => spriteColor(call[0]))
    .filter((c): c is [number, number, number] => !!c)
    .map((c) => rgba(c, 1));
}

function withIdentityTransform(ctx: ReturnType<typeof createMockCanvasContext>) {
  (ctx as unknown as { getTransform: () => DOMMatrix }).getTransform = () =>
    ({ a: 1, b: 0, c: 0, d: 1, e: 0, f: 0 }) as DOMMatrix;
  return ctx;
}

beforeEach(() => {
  setGraphStyle("light");
  beginLightFrame(1, 0, false);
});

afterEach(() => {
  setLightFocusMix(1);
  setLightThreadFade(null);
  setLightAmbient(true);
  setGraphStyle(initialStyle);
  ensureCelestialBodyDrawers();
});

describe("light palette", () => {
  it("converts, mixes and clamps colours", () => {
    expect(hexToRgb("#7fb2ff")).toEqual([127, 178, 255]);
    expect(mixRgb([0, 0, 0], [200, 100, 50], 0.5)).toEqual([100, 50, 25]);
    expect(rgba([10, 20, 30], 2)).toBe("rgba(10,20,30,1.000)");
  });

  it("gives each type family its own colour and falls back for unknown types", () => {
    expect(lightTypeColor("comet")).not.toEqual(lightTypeColor("nebula"));
    expect(lightTypeColor("no-such-type")).toEqual(lightTypeColor(undefined));
  });

  it("reads importance as size: galaxy › star › planet › moon", () => {
    const r = 16;
    const sizes = ["galaxy", "star", "planet", "moon"].map((t) => lightCoreRadius(t, r));
    expect([...sizes].sort((a, b) => b - a)).toEqual(sizes);
  });
});

describe("light glyphs", () => {
  it("re-registers drawers when the style changes", () => {
    setGraphStyle("classic");
    ensureCelestialBodyDrawers();
    const classicStar = CelestialBody.STAR.drawFunction;
    setGraphStyle("light");
    ensureCelestialBodyDrawers();
    expect(CelestialBody.STAR.drawFunction).not.toBe(classicStar);
  });

  it("draws a star as additive light and ignores random per-note colour", () => {
    ensureCelestialBodyDrawers();
    const ctx = createMockCanvasContext();
    const composites: string[] = [];
    Object.defineProperty(ctx, "globalCompositeOperation", {
      get: () => composites[composites.length - 1] ?? "source-over",
      set: (v: string) => composites.push(v),
      configurable: true,
    });
    CelestialBody.STAR.drawFunction!(ctx, {
      x: 0,
      y: 0,
      r: 16,
      angle: 0,
      nodeId: "n1",
      variation: {
        color: "#ff0000",
        glowColor: "#ff0000",
        strokeColor: "#ff0000",
        sizeMultiplier: 1,
        hueShift: 0,
        phaseShift: 0,
      } as never,
    });
    expect(composites).toContain("lighter");
    expect(ctx.drawImage).toHaveBeenCalled();
    const all = [...stampedColors(ctx), ...ctx.getFillStyles(), ...ctx.getStrokeStyles()];
    expect(stampedColors(ctx).length).toBeGreaterThan(0);
    expect(all.some((c) => String(c).startsWith("rgba(255,0,0"))).toBe(false);
  });

  it("draws a black hole with a dark core inside a ring", () => {
    ensureCelestialBodyDrawers();
    const ctx = createMockCanvasContext();
    CelestialBody.BLACKHOLE.drawFunction!(ctx, { x: 0, y: 0, r: 16, angle: 0, nodeId: "bh" });
    expect(ctx.getFillStyles()).toContain("rgba(3,4,9,0.950)");
    expect(ctx.stroke).toHaveBeenCalled();
  });

  it("draws a comet with a tail", () => {
    ensureCelestialBodyDrawers();
    const ctx = createMockCanvasContext();
    CelestialBody.COMET.drawFunction!(ctx, { x: 0, y: 0, r: 16, angle: 0, nodeId: "c" });
    expect(ctx.createLinearGradient).toHaveBeenCalled();
  });

  it("COMET-1: a settled comet has no tail, a live one does", () => {
    ensureCelestialBodyDrawers();
    const settled = createMockCanvasContext();
    CelestialBody.COMET.drawFunction!(settled, {
      x: 0,
      y: 0,
      r: 16,
      angle: 0,
      nodeId: "c",
      doneAt: new Date(Date.now() - 60_000).toISOString(),
    });
    expect(settled.createLinearGradient).not.toHaveBeenCalled();

    const live = createMockCanvasContext();
    CelestialBody.COMET.drawFunction!(live, {
      x: 0,
      y: 0,
      r: 16,
      angle: 0,
      nodeId: "c",
      dueAt: new Date(Date.now() + 3600_000).toISOString(),
    });
    expect(live.createLinearGradient).toHaveBeenCalled();
  });

  it("keeps the halo still in snapshot mode and lets it breathe otherwise", () => {
    ensureCelestialBodyDrawers();
    const drawAt = (time: number, stable: boolean) => {
      beginLightFrame(1, time, stable);
      const ctx = createMockCanvasContext();
      CelestialBody.STAR.drawFunction!(ctx, { x: 0, y: 0, r: 16, angle: 0, nodeId: "s" });
      return ctx.getGlobalAlphas()[0];
    };
    expect(drawAt(0, true)).toBeDefined();
    expect(drawAt(0, true)).toBe(drawAt(1700, true));
    expect(drawAt(0, false)).not.toBe(drawAt(1700, false));
  });

  it("stops breathing when background motion is off", () => {
    ensureCelestialBodyDrawers();
    setLightAmbient(false);
    const drawAt = (time: number) => {
      beginLightFrame(1, time, false);
      const ctx = createMockCanvasContext();
      CelestialBody.STAR.drawFunction!(ctx, { x: 0, y: 0, r: 16, angle: 0, nodeId: "s" });
      return ctx.getGlobalAlphas()[0];
    };
    expect(drawAt(0)).toBeDefined();
    expect(drawAt(0)).toBe(drawAt(1700));
  });

  it("stamps each halo from one sprite per colour", () => {
    ensureCelestialBodyDrawers();
    const ctx = createMockCanvasContext();
    for (const id of ["p1", "p2", "p3"]) {
      CelestialBody.PLANET.drawFunction!(ctx, { x: 0, y: 0, r: 16, angle: 0, nodeId: id });
    }
    const images = (ctx.drawImage as unknown as ReturnType<typeof vi.fn>).mock.calls.map(
      (call) => call[0]
    );
    expect(images.length).toBe(6);
    expect(new Set(images).size).toBe(2);
    expect(ctx.createRadialGradient).not.toHaveBeenCalled();
  });
});

describe("light threads", () => {
  const a: SimulationNode = { id: "a", title: "a", type: "star", x: 0, y: 0 };
  const b: SimulationNode = { id: "b", title: "b", type: "planet", x: 100, y: 0 };
  const manual: SimulationLink = { source: "a", target: "b", link_type: "related", weight: 0.5 };
  const auto: SimulationLink = {
    source: "a",
    target: "b",
    link_type: "related",
    source_type: "gamma",
    weight: 0.7,
  };
  const dependency: SimulationLink = { source: "a", target: "b", link_type: "dependency" };
  const unrelated: SimulationLink = { source: "b", target: "c", link_type: "related" };

  it("lights the hovered neighbourhood and almost hides the rest", () => {
    expect(lightThreadAlpha(manual, { fadeOpacity: 1 })).toBeCloseTo(0.28);
    expect(lightThreadAlpha(manual, { fadeOpacity: 1, hoveredNodeId: "a" })).toBeCloseTo(0.85);
    expect(lightThreadAlpha(unrelated, { fadeOpacity: 1, hoveredNodeId: "a" })).toBe(
      LIGHT_DIMMED_THREAD_ALPHA
    );
  });

  it("fades the focus in between plain and focused brightness", () => {
    setLightFocusMix(0.5);
    expect(lightThreadAlpha(manual, { fadeOpacity: 1, hoveredNodeId: "a" })).toBeCloseTo(
      (0.28 + 0.85) / 2
    );
  });

  it("fades a thread with whichever end has gone further toward the list", () => {
    setLightThreadFade(new Map([["a", 0.25]]));
    expect(lightThreadAlpha(manual, { fadeOpacity: 1 })).toBeCloseTo(0.28 * 0.25);
    setLightThreadFade(
      new Map([
        ["a", 0.6],
        ["b", 0.1],
      ])
    );
    expect(lightThreadAlpha(manual, { fadeOpacity: 1 })).toBeCloseTo(0.28 * 0.1);
  });

  it("draws threads as curves, model links as dots and dependencies in gold", () => {
    const ctx = createMockCanvasContext();
    drawLightLink(ctx, manual, a, b, { fadeOpacity: 1 });
    expect(ctx.quadraticCurveTo).toHaveBeenCalled();

    const dotted = createMockCanvasContext();
    drawLightLink(dotted, auto, a, b, { fadeOpacity: 1 });
    expect(dotted.setLineDash).toHaveBeenCalledWith([0.5, 4.2]);

    const gold = createMockCanvasContext();
    drawLightLink(gold, dependency, a, b, { fadeOpacity: 1 });
    expect(gold.getStrokeStyles().some((s) => String(s).startsWith("rgba(242,196,109"))).toBe(true);
  });

  it("runs dots along a highlighted chain, except in snapshot mode", () => {
    const chain = {
      nodeDepth: new Map([
        ["a", 0],
        ["b", 1],
      ]),
      linkDepth: new Map([["a|b", 0]]),
      cycleLinks: new Set<string>(),
      hasCycle: false,
    };
    const moving = createMockCanvasContext();
    drawLightLink(moving, dependency, a, b, { fadeOpacity: 1, depChain: chain as never });
    expect(moving.drawImage).toHaveBeenCalled();

    beginLightFrame(1, 0, true);
    const still = createMockCanvasContext();
    drawLightLink(still, dependency, a, b, { fadeOpacity: 1, depChain: chain as never });
    expect(still.drawImage).not.toHaveBeenCalled();
  });
});

describe("light captions", () => {
  const node = (id: string, title: string, x = 0, y = 0, type = "planet"): SimulationNode => ({
    id,
    title,
    type,
    x,
    y,
  });

  it("does not cut an ordinary title and shortens only very long ones", () => {
    const ctx = withIdentityTransform(createMockCanvasContext());
    const long = "Очень длинное название заметки о миграциях базы данных";
    drawLightCaptions(
      ctx,
      [
        { node: node("a", "Миграции без простоя в Postgres", 0, 0), opacity: 1, priority: 1 },
        { node: node("b", long, 0, 200), opacity: 1, priority: 1 },
      ],
      16
    );
    const texts = (ctx.fillText as ReturnType<typeof vi.fn>).mock.calls.map((call) => call[0]);
    expect(texts).toContain("Миграции без простоя в Postgres");
    expect(texts.find((t: string) => t.endsWith("…"))?.length).toBe(40);
  });

  it("places the more important caption first and skips one that would overlap", () => {
    const ctx = withIdentityTransform(createMockCanvasContext());
    drawLightCaptions(
      ctx,
      [
        { node: node("low", "Второстепенная", 0, 0), opacity: 1, priority: 1 },
        { node: node("mid", "Средняя", 0, 0), opacity: 1, priority: 2 },
        { node: node("top", "Главная", 0, 0), opacity: 1, priority: 3 },
      ],
      16
    );
    const texts = (ctx.fillText as ReturnType<typeof vi.fn>).mock.calls.map((call) => call[0]);
    expect(texts[0]).toBe("Главная");
    expect(texts).toContain("Средняя");
    expect(texts).not.toContain("Второстепенная");
  });

  it("ranks the hovered note above its neighbours and neighbours above the rest", () => {
    const n = node("n", "x");
    expect(captionPriority(n, "n")).toBeGreaterThan(captionPriority(n, "other", new Set(["n"])));
    expect(captionPriority(n, "other", new Set(["n"]))).toBeGreaterThan(captionPriority(n, null));
  });
});

describe("light style in the draw pipeline", () => {
  it("dims notes outside the hovered neighbourhood to 15 %", () => {
    const ctx = createMockCanvasContext();
    const nodes: SimulationNode[] = [
      { id: "h", title: "h", type: "star", x: 0, y: 0 },
      { id: "n", title: "n", type: "planet", x: 50, y: 0 },
      { id: "far", title: "far", type: "planet", x: 200, y: 0 },
    ];
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
      new Set(["n"])
    );
    const alphas = ctx.getGlobalAlphas();
    expect(alphas).toContain(0.15);
    expect(alphas.filter((v) => v === 1).length).toBeGreaterThanOrEqual(2);
  });

  it("dims the rest gradually while the focus fades in", () => {
    setLightFocusMix(0.5);
    const ctx = createMockCanvasContext();
    const nodes: SimulationNode[] = [
      { id: "h", title: "h", type: "star", x: 0, y: 0 },
      { id: "far", title: "far", type: "planet", x: 200, y: 0 },
    ];
    drawAllNodes(ctx, nodes, new Map(), false, undefined, false, 0, "h", null, false);
    expect(ctx.getGlobalAlphas().some((a) => Math.abs(a - 0.575) < 1e-9)).toBe(true);
  });

  it("paints the deep-ink sky instead of the classic background", () => {
    const ctx = createMockCanvasContext();
    draw(ctx, 400, 300, [], [], new Map(), { x: 0, y: 0, k: 2 });
    // The light sky's vignette is centred slightly above the middle.
    expect(ctx.createRadialGradient).toHaveBeenCalledWith(200, 135, 0, 200, 135, 300);
    expect(lightFrame.k).toBe(2);
  });

  it("keeps the sky in a layer while the camera stands still and paints it straight while it moves", () => {
    // Every DOM canvas here shares the mocked context of vitest-setup, so the
    // sky layer's paint calls land on one shared fillRect; the frame's own
    // paint calls land on the frame's fillRect.
    const layerPaint = (
      document.createElement("canvas").getContext("2d") as unknown as {
        fillRect: ReturnType<typeof vi.fn>;
      }
    ).fillRect;
    const ctx = createMockCanvasContext();
    const frameFill = ctx.fillRect as unknown as ReturnType<typeof vi.fn>;
    const direct = () => frameFill.mock.calls.length;
    const layered = () => layerPaint.mock.calls.length;
    const stamps = () => (ctx.drawImage as unknown as ReturnType<typeof vi.fn>).mock.calls.length;
    const sky = (x: number, time: number, twinkle: boolean) =>
      drawLightBackground(ctx, 333, 222, { x, y: 5 }, time, twinkle);

    sky(5, 0, false); // the camera has just arrived: straight into the frame
    const d1 = direct();
    const l1 = layered();
    expect(d1).toBeGreaterThan(0);
    expect(stamps()).toBe(0);

    sky(5, 900, false); // it stands still: kept in the layer
    const l2 = layered();
    expect(l2).toBeGreaterThan(l1);
    expect(direct()).toBe(d1);
    expect(stamps()).toBe(1);

    sky(5, 1800, false); // still: the layer is reused as is
    expect(layered()).toBe(l2);
    expect(direct()).toBe(d1);
    expect(stamps()).toBe(2);

    sky(60, 1800, false); // moving: straight into the frame, the layer untouched
    expect(direct()).toBeGreaterThan(d1);
    expect(layered()).toBe(l2);
    expect(stamps()).toBe(2);

    sky(60, 1000, true); // twinkling to a new step: straight into the frame
    sky(60, 1050, true); // same step: into the layer
    const l3 = layered();
    expect(l3).toBeGreaterThan(l2);
    sky(60, 1090, true); // same step again: reused
    expect(layered()).toBe(l3);
    expect(stamps()).toBe(4);
  });

  it("twinkles the stars only while background motion runs", () => {
    const starStyles = (time: number, twinkle: boolean) => {
      const sky = createMockCanvasContext();
      const styles: string[] = [];
      Object.defineProperty(sky, "fillStyle", {
        set: (v: string) => styles.push(v),
        get: () => styles[styles.length - 1] ?? "",
        configurable: true,
      });
      paintLightSky(sky, 400, 300, { x: 0, y: 0 }, time, twinkle);
      return styles.filter((v) => String(v).startsWith("rgba(210,222,255")).join("|");
    };
    expect(starStyles(0, false)).not.toBe("");
    expect(starStyles(0, false)).toBe(starStyles(1700, false));
    expect(starStyles(0, true)).not.toBe(starStyles(1700, true));
  });
});

describe("background motion switch", () => {
  const base = { reducedMotion: false, snapshot: false, nodeCount: 100, maxNodes: 400 };

  it("runs on an ordinary graph", () => {
    expect(lightAmbient(base)).toBe(true);
    expect(lightAmbient({ ...base, nodeCount: 400 })).toBe(true);
  });

  it("stops when the system asks to reduce motion, in snapshot mode and on large graphs", () => {
    expect(lightAmbient({ ...base, reducedMotion: true })).toBe(false);
    expect(lightAmbient({ ...base, snapshot: true })).toBe(false);
    expect(lightAmbient({ ...base, nodeCount: 401 })).toBe(false);
  });
});
