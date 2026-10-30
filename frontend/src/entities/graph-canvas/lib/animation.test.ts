/**
 * Unit tests for animation module
 */
import { describe, it, expect, vi } from "vitest";

describe("animation", () => {
  async function loadAnimation() {
    vi.doUnmock("./animation");
    return import("./animation");
  }

  it("should return different base speeds per node type", async () => {
    const { getBaseSpeed } = await loadAnimation();
    expect(getBaseSpeed("star")).toBe(0.005);
    expect(getBaseSpeed("planet")).toBe(0.02);
    expect(getBaseSpeed("comet")).toBe(0.03);
    expect(getBaseSpeed("galaxy")).toBe(0.01);
    expect(getBaseSpeed("nebula")).toBe(0.008);
  });

  it("should update node angles based on type with per-node variation", async () => {
    const { updateNodeAngles } = await loadAnimation();
    const angles = new Map<string, number>();
    const speeds = new Map<string, number>();
    const nodes = [
      { id: "star1", type: "star" },
      { id: "planet1", type: "planet" },
    ];

    updateNodeAngles(nodes, angles, speeds);

    // Each node should receive a non-deterministic initial angle and speed,
    // so nodes of the same type do not rotate in lockstep.
    expect(angles.get("star1")).toBeGreaterThanOrEqual(0);
    expect(angles.get("planet1")).toBeGreaterThanOrEqual(0);
    expect(angles.get("star1")).not.toBe(angles.get("planet1"));
    expect(speeds.get("star1")).not.toBe(0);
    expect(speeds.get("planet1")).not.toBe(0);

    const prevStar = angles.get("star1")!;
    const prevPlanet = angles.get("planet1")!;
    updateNodeAngles(nodes, angles, speeds);

    expect(angles.get("star1")).toBe(prevStar + speeds.get("star1")!);
    expect(angles.get("planet1")).toBe(prevPlanet + speeds.get("planet1")!);
  });

  it("should keep angles fixed in disable variation mode", async () => {
    const { updateNodeAngles } = await loadAnimation();
    const angles = new Map<string, number>();
    const speeds = new Map<string, number>();
    const nodes = [{ id: "star1", type: "star" }];

    updateNodeAngles(nodes, angles, speeds, true);

    expect(angles.get("star1")).toBe(0);
    expect(speeds.get("star1")).toBe(0);
  });

  it("should reuse stored speeds", async () => {
    const { updateNodeAngles } = await loadAnimation();
    const angles = new Map<string, number>();
    const speeds = new Map<string, number>([["star1", 0.123]]);
    const nodes = [{ id: "star1", type: "star" }];

    updateNodeAngles(nodes, angles, speeds);

    expect(angles.get("star1")).toBe(0.123);
    expect(speeds.get("star1")).toBe(0.123);
  });

  it("should start and stop animation loop", async () => {
    const { startAnimationLoop } = await loadAnimation();
    const onUpdate = vi.fn();
    const loop = startAnimationLoop(onUpdate);

    expect(loop).toHaveProperty("stop");
    loop.stop();
  });

  it("should not crash when requestAnimationFrame is not available", async () => {
    const { startAnimationLoop } = await loadAnimation();
    const originalRaf = globalThis.requestAnimationFrame;
    const originalCaf = globalThis.cancelAnimationFrame;
    // @ts-expect-error test environment without RAF
    delete globalThis.requestAnimationFrame;
    // @ts-expect-error test environment without RAF
    delete globalThis.cancelAnimationFrame;

    const onUpdate = vi.fn();
    const loop = startAnimationLoop(onUpdate);

    expect(loop).toHaveProperty("stop");
    expect(() => loop.stop()).not.toThrow();

    globalThis.requestAnimationFrame = originalRaf;
    globalThis.cancelAnimationFrame = originalCaf;
  });

  it("should clear animation state", async () => {
    const { clearAnimationState } = await loadAnimation();
    const angles = new Map<string, number>([["a", 1]]);
    const speeds = new Map<string, number>([["a", 0.1]]);

    clearAnimationState(angles, speeds);

    expect(angles.size).toBe(0);
    expect(speeds.size).toBe(0);
  });
});
