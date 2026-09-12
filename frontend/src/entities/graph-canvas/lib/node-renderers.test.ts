import { describe, it, expect, vi, beforeEach } from "vitest";
import type { NodeVariation } from "$shared/utils/variation";
import {
  drawStar,
  drawPlanet,
  drawComet,
  drawGalaxy,
  drawNebula,
  drawAsteroid,
  drawDebris,
  drawDust,
  drawBlackhole,
  drawTechnicalNode,
  drawMoon,
  drawUnknown,
} from "./node-renderers";

function createCtx() {
  const gradient = { addColorStop: vi.fn() };
  return {
    beginPath: vi.fn(),
    moveTo: vi.fn(),
    lineTo: vi.fn(),
    closePath: vi.fn(),
    arc: vi.fn(),
    ellipse: vi.fn(),
    quadraticCurveTo: vi.fn(),
    stroke: vi.fn(),
    fill: vi.fn(),
    save: vi.fn(),
    restore: vi.fn(),
    translate: vi.fn(),
    rotate: vi.fn(),
    fillText: vi.fn(),
    createRadialGradient: vi.fn(() => gradient),
    shadowBlur: 0,
    shadowColor: "",
    lineJoin: "",
    lineCap: "",
    strokeStyle: "",
    fillStyle: "",
    lineWidth: 0,
    globalAlpha: 1,
    font: "",
    textAlign: "",
    textBaseline: "",
  };
}

function v(partial: Partial<NodeVariation>) {
  return { sizeMultiplier: 1, hueShift: 0, phaseShift: 0, color: "", glowColor: "", strokeColor: "", ...partial } as unknown as NodeVariation;
}

describe("node renderers", () => {
  let ctx: ReturnType<typeof createCtx>;

  beforeEach(() => {
    ctx = createCtx();
  });

  it("draws a star with all effects enabled", () => {
    drawStar(ctx as any, 10, 10, 5, 0, v({ color: "#ffcc00", glowColor: "#ffcc00", strokeColor: "#cc9900" }), "n1", 50, 1000);
    expect(ctx.beginPath).toHaveBeenCalled();
  });

  it("draws a star without effects when count is high", () => {
    drawStar(ctx as any, 10, 10, 5, 0, undefined, "n1", 5000);
    expect(ctx.beginPath).toHaveBeenCalled();
  });

  it("draws a planet with rings and glow", () => {
    drawPlanet(ctx as any, 10, 10, 5, 0.5, v({ color: "#d6aa5d", glowColor: "#d6aa5d" }), "n1", 50, 1000);
    expect(ctx.ellipse).toHaveBeenCalled();
  });

  it("draws a planet without glow above threshold", () => {
    drawPlanet(ctx as any, 10, 10, 5, 0, undefined, undefined, 500, 1000);
    expect(ctx.arc).toHaveBeenCalled();
  });

  it("draws a comet with a tail", () => {
    drawComet(ctx as any, 10, 10, 5, 0.5, v({ color: "#e879f9", glowColor: "#e879f9" }), "n1", 50, 1000);
    expect(ctx.quadraticCurveTo).toHaveBeenCalled();
  });

  it("draws a comet without time", () => {
    drawComet(ctx as any, 10, 10, 5, 0, undefined);
    expect(ctx.arc).toHaveBeenCalled();
  });

  it("draws a galaxy with spiral arms", () => {
    drawGalaxy(ctx as any, 10, 10, 5, 0.5, v({ color: "#8b5cf6", glowColor: "#8b5cf6" }), "n1", 50, 1000);
    expect(ctx.stroke).toHaveBeenCalled();
  });

  it("draws a galaxy above visual threshold", () => {
    drawGalaxy(ctx as any, 10, 10, 5, 0, undefined, "n1", 1000);
    expect(ctx.arc).toHaveBeenCalled();
  });

  it("draws a nebula with and without variation", () => {
    drawNebula(ctx as any, 10, 10, 5, 0);
    drawNebula(ctx as any, 10, 10, 5, 0, v({ color: "#2dd4bf" }));
    expect(ctx.ellipse).toHaveBeenCalled();
  });

  it("draws an asteroid with craters and variation", () => {
    drawAsteroid(ctx as any, 10, 10, 5, 0, v({ color: "#94a3b8", glowColor: "#94a3b8", strokeColor: "#64748b" }), false, "n1", 50, 1000);
    expect(ctx.arc).toHaveBeenCalled();
  });

  it("draws an asteroid with disabled variation", () => {
    drawAsteroid(ctx as any, 10, 10, 5, 0, undefined, true);
    expect(ctx.beginPath).toHaveBeenCalled();
  });

  it("draws debris with and without disabled variation", () => {
    drawDebris(ctx as any, 10, 10, 5, 0, false, "n1", v({ color: "#969696" }));
    drawDebris(ctx as any, 10, 10, 5, 0, true, undefined, v({ color: "#969696" }));
    expect(ctx.arc).toHaveBeenCalled();
  });

  it("draws dust with and without disabled variation", () => {
    drawDust(ctx as any, 10, 10, 5, 0, false, "n1", v({ color: "#a0a0a0" }));
    drawDust(ctx as any, 10, 10, 5, 0, true, undefined, v({ color: "#a0a0a0" }));
    expect(ctx.arc).toHaveBeenCalled();
  });

  it("draws a black hole with and without glow", () => {
    drawBlackhole(ctx as any, 10, 10, 5, 0, "n1", 50, 1000, v({ color: "#000000", glowColor: "#ff6600" }));
    drawBlackhole(ctx as any, 10, 10, 5, 0);
    expect(ctx.arc).toHaveBeenCalled();
  });

  it("draws a technical node with and without animation time", () => {
    drawTechnicalNode(ctx as any, 10, 10, 5, 1000, v({ color: "#8b5cf6", glowColor: "#8b5cf6" }));
    drawTechnicalNode(ctx as any, 10, 10, 5);
    expect(ctx.fillText).toHaveBeenCalled();
  });

  it("draws a moon with and without variation", () => {
    drawMoon(ctx as any, 10, 10, 5, 0);
    drawMoon(ctx as any, 10, 10, 5, 0, v({ color: "#cccccc", strokeColor: "#999999", glowColor: "#aaaaaa" }));
    expect(ctx.arc).toHaveBeenCalled();
  });

  it("draws an unknown node with custom and default renderers", () => {
    const customRenderer = vi.fn();
    drawUnknown(ctx as any, 10, 10, 5, 0, "n2", { 0: customRenderer });
    expect(customRenderer).toHaveBeenCalled();

    const wrongRenderer = vi.fn();
    // "n3" maps to type 1, the custom map only has 0, so it falls back to default drawRealityRift
    drawUnknown(ctx as any, 10, 10, 5, 0, "n3", { 0: wrongRenderer });
    expect(wrongRenderer).not.toHaveBeenCalled();

    // no custom renderers, uses the default map
    drawUnknown(ctx as any, 10, 10, 5, 0, "n4");
  });
});
