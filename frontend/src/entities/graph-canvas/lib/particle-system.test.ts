import { describe, it, expect } from "vitest";
import { graphConfig2D } from "$shared/config";
import { ParticleSystem } from "./particle-system";

function createMockCtx(): CanvasRenderingContext2D {
  return {
    globalAlpha: 1,
    fillStyle: "",
    beginPath: () => undefined,
    arc: () => undefined,
    fill: () => undefined,
  } as unknown as CanvasRenderingContext2D;
}

describe("ParticleSystem", () => {
  it("is disabled above the visual-fx threshold and can be enabled via updateNodeCount", () => {
    const threshold = graphConfig2D.visual_fx_threshold ?? 500;
    const ps = new ParticleSystem(threshold + 1);
    expect(ps.isEnabled()).toBe(false);

    ps.updateNodeCount(3);
    expect(ps.isEnabled()).toBe(true);
  });

  it("is enabled at zero node count (no nodes yet; will init as data arrives)", () => {
    const ps = new ParticleSystem(0);
    expect(ps.isEnabled()).toBe(true);
  });

  it("applies per-particle alpha through ctx.globalAlpha", () => {
    const ps = new ParticleSystem(3);
    ps.initParticles("n1", 100, 100, "#ffcc00");

    const ctx = createMockCtx();
    const capturedAlphas: number[] = [];
    Object.defineProperty(ctx, "globalAlpha", {
      set(value: number) {
        capturedAlphas.push(value);
      },
      get() {
        return 1;
      },
      configurable: true,
    });

    ps.draw(ctx, "n1");

    // Each particle sets alpha, draws, and then the final restore is captured.
    expect(capturedAlphas.length).toBeGreaterThan(1);
    // First real alpha must be < 1 because particles have alpha in [0.25, 0.8].
    expect(capturedAlphas[0]).toBeLessThan(1);
    expect(capturedAlphas[0]).toBeGreaterThanOrEqual(0.25);
    // Last call should restore base alpha (1).
    expect(capturedAlphas[capturedAlphas.length - 1]).toBe(1);
  });

  it("does not init or draw when disabled", () => {
    const threshold = graphConfig2D.visual_fx_threshold ?? 500;
    const ps = new ParticleSystem(threshold + 1);
    const ctx = createMockCtx();
    ps.draw(ctx, "n1");
    expect(ps.isEnabled()).toBe(false);
  });
});
