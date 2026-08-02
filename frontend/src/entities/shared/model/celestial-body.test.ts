import { describe, it, expect, vi } from "vitest";
import { CelestialBody } from "./celestial-body";

describe("CelestialBody", () => {
  it("returns the correct body for known types (case-insensitive)", () => {
    expect(CelestialBody.fromString("STAR")).toBe(CelestialBody.STAR);
    expect(CelestialBody.fromString("Planet")).toBe(CelestialBody.PLANET);
    expect(CelestialBody.fromString("  BLACKHOLE ")).toBe(CelestialBody.BLACKHOLE);
  });

  it("falls back to UNKNOWN for unrecognized or missing types", () => {
    expect(CelestialBody.fromString("banana")).toBe(CelestialBody.UNKNOWN);
    expect(CelestialBody.fromString(undefined)).toBe(CelestialBody.UNKNOWN);
    expect(CelestialBody.fromString("")).toBe(CelestialBody.UNKNOWN);
  });

  it("exposes visual properties as getters", () => {
    const star = CelestialBody.STAR;
    expect(star.type).toBe("star");
    expect(star.label).toBe("Star");
    expect(star.description).toBe(
      "A central topic or pillar idea that anchors a cluster of notes."
    );
    expect(star.example).toBe("Culture, Anime, Backend");
    expect(star.emoji).toBe("⭐");
    expect(star.color).toBe("#ffcc00");
    expect(star.glowColor).toBe("#ffcc00");
    expect(star.baseRadius).toBe(1);
    expect(star.minRadius).toBe(0.8);
    expect(star.maxRadius).toBe(1.2);
    expect(star.baseSpeed).toBe(0.005);
    expect(star.gravityOffset).toBe(20);
    expect(star.cssVarName).toBe("--color-star");
    expect(star.isUi).toBe(true);
    expect(star.isAnomaly).toBe(false);
  });

  it("groups UI and anomaly types correctly", () => {
    const uiTypes = CelestialBody.UI_TYPES.map((b) => b.type);
    expect(uiTypes).toContain("star");
    expect(uiTypes).toContain("planet");
    expect(uiTypes).toContain("blackhole");
    expect(uiTypes).not.toContain("moon");
    expect(uiTypes).not.toContain("unknown");

    const anomalies = CelestialBody.ANOMALIES.map((b) => b.type);
    expect(anomalies).toContain("unknown");
    expect(anomalies).toContain("reality_rift");
    expect(anomalies).not.toContain("star");
  });

  it("produces a CSS color expression with fallback", () => {
    expect(CelestialBody.STAR.toCSSColor()).toBe("var(--color-star, #ffcc00)");
  });

  it("draw throws if no draw function is registered", () => {
    const body = new CelestialBody({
      type: "test",
      label: "Test",
      description: "celestialBody.type.test.description",
      example: "celestialBody.type.test.example",
      emoji: "✨",
      color: "#ffffff",
      glowColor: "#ffffff",
      baseRadius: 1,
      minRadius: 0.8,
      maxRadius: 1.2,
      gravityMass: 1,
      baseSpeed: 0,
      gravityOffset: 10,
    });

    const mockCtx = {} as CanvasRenderingContext2D;
    expect(() => body.draw(mockCtx, { x: 0, y: 0, r: 10, angle: 0, nodeId: "n1" })).toThrow(
      'Draw function not registered for celestial body type "test"'
    );
  });

  it("delegates to the registered draw function", () => {
    const drawFn = vi.fn();
    const body = new CelestialBody({
      type: "test",
      label: "Test",
      description: "celestialBody.type.test.description",
      example: "celestialBody.type.test.example",
      emoji: "✨",
      color: "#ffffff",
      glowColor: "#ffffff",
      baseRadius: 1,
      minRadius: 0.8,
      maxRadius: 1.2,
      gravityMass: 1,
      baseSpeed: 0,
      gravityOffset: 10,
      drawFunction: drawFn,
    });

    const mockCtx = {} as CanvasRenderingContext2D;
    const context = { x: 1, y: 2, r: 3, angle: 4, nodeId: "n1" };
    body.draw(mockCtx, context);
    expect(drawFn).toHaveBeenCalledWith(mockCtx, context);
  });
});
