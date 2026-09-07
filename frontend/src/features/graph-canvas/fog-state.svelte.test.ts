import { describe, it, expect } from "vitest";
import { createFogState, type FogState, type FogMode } from "./fog-state.svelte";
import type { SimulationNode } from "$entities/graph-canvas/lib/types";
import { graphConfig2D } from "$shared/config";

const FOG = graphConfig2D.fog;
const PERFORMANCE_SAMPLE_COUNT = 30;

function simulateFps(state: FogState, fps: number, frames = PERFORMANCE_SAMPLE_COUNT + 5) {
  const frameMs = 1000 / fps;
  for (let i = 0; i < frames; i++) {
    state.tick(frameMs * i);
  }
}

function updateUntil(
  state: FogState,
  predicate: () => boolean,
  maxFrames = 200,
  hoveredNode: SimulationNode | null = null,
  focusMode = false
) {
  for (let i = 0; i < maxFrames; i++) {
    state.update(800, 600, { x: 0, y: 0, k: 1 }, hoveredNode, focusMode);
    if (predicate()) break;
  }
}

describe("createFogState", () => {
  it("starts with fog enabled and in atmospheric mode", () => {
    const state = createFogState();
    expect(state.enabled).toBe(true);
    expect(state.showWarning).toBe(false);
    expect(state.snapshot.mode).toBe("atmospheric" satisfies FogMode);
    expect(state.snapshot.radius).toBe(FOG.radius_max);
  });

  it("toggle disables fog and switches to off mode", () => {
    const state = createFogState();
    state.toggle();
    expect(state.enabled).toBe(false);
    expect(state.snapshot.mode).toBe("off" satisfies FogMode);
    expect(state.snapshot.radius).toBe(FOG.radius_max);
  });

  it("toggle re-enables fog and restores atmospheric mode", () => {
    const state = createFogState();
    state.toggle();
    state.toggle();
    expect(state.enabled).toBe(true);
    expect(state.snapshot.mode).toBe("atmospheric" satisfies FogMode);
  });

  it("focus mode expands the fog and enters first-person mode", () => {
    const state = createFogState();
    simulateFps(state, 15); // low fps
    updateUntil(
      state,
      () =>
        state.snapshot.mode === "adaptive" &&
        state.snapshot.radius <= FOG.radius_min + 50,
      200
    );
    updateUntil(
      state,
      () =>
        state.snapshot.mode === "first-person" &&
        state.snapshot.radius >= FOG.radius_max - 50,
      200,
      null,
      true
    );
    expect(state.snapshot.mode).toBe("first-person" satisfies FogMode);
    expect(state.snapshot.radius).toBeGreaterThanOrEqual(FOG.radius_max - 50);
    expect(state.showWarning).toBe(false);
  });

  it("drops into adaptive mode and shrinks radius when FPS is low", () => {
    const state = createFogState();
    simulateFps(state, 15);
    updateUntil(
      state,
      () =>
        state.snapshot.mode === "adaptive" &&
        state.snapshot.radius <= FOG.radius_min + 50,
      200
    );
    expect(state.snapshot.mode).toBe("adaptive" satisfies FogMode);
    expect(state.snapshot.radius).toBeLessThanOrEqual(FOG.radius_min + 50);
  });

  it("returns to atmospheric mode and full radius when FPS recovers", () => {
    const state = createFogState();
    simulateFps(state, 15);
    updateUntil(
      state,
      () =>
        state.snapshot.mode === "adaptive" &&
        state.snapshot.radius <= FOG.radius_min + 50,
      200
    );

    simulateFps(state, 60);
    updateUntil(
      state,
      () =>
        state.snapshot.mode === "atmospheric" &&
        state.snapshot.radius >= FOG.radius_max - 50,
      200
    );
    expect(state.snapshot.mode).toBe("atmospheric" satisfies FogMode);
    expect(state.snapshot.radius).toBeGreaterThanOrEqual(FOG.radius_max - 50);
  });

  it("centers the fog on a hovered node", () => {
    const state = createFogState();
    const hoveredNode = { id: "n1", title: "N", type: "star", x: 50, y: 75 };
    simulateFps(state, 60);

    // Single update starts the smooth approach.
    state.update(800, 600, { x: 0, y: 0, k: 1 }, hoveredNode, false);

    // Call update repeatedly until the center settles near the hovered screen position.
    const expectedX = (hoveredNode.x ?? 0) * 1 + 0;
    const expectedY = (hoveredNode.y ?? 0) * 1 + 0;
    updateUntil(
      state,
      () =>
        Math.abs(state.snapshot.centerX - expectedX) < 0.1 &&
        Math.abs(state.snapshot.centerY - expectedY) < 0.1,
      300,
      hoveredNode
    );

    expect(state.snapshot.centerX).toBeCloseTo(expectedX, 0);
    expect(state.snapshot.centerY).toBeCloseTo(expectedY, 0);
  });

  it("shows a warning only in adaptive mode with critically low FPS", () => {
    const state = createFogState();
    simulateFps(state, 10);
    updateUntil(
      state,
      () =>
        state.snapshot.mode === "adaptive" &&
        state.snapshot.radius <= FOG.radius_min + 50,
      200
    );
    expect(state.showWarning).toBe(true);
    expect(state.snapshot.showWarning).toBe(true);

    simulateFps(state, 60);
    updateUntil(
      state,
      () =>
        state.snapshot.mode === "atmospheric" &&
        state.snapshot.radius >= FOG.radius_max - 50,
      200
    );
    expect(state.showWarning).toBe(false);

    state.toggle();
    simulateFps(state, 10);
    state.update(800, 600, { x: 0, y: 0, k: 1 }, null, false);
    expect(state.showWarning).toBe(false);
  });

  it("expands the clear radius to include a hovered node and its direct neighbors", () => {
    const state = createFogState();
    simulateFps(state, 60);

    const nodes: SimulationNode[] = [
      { id: "n0", title: "H", type: "star", x: 0, y: 0 },
      { id: "n1", title: "A", type: "planet", x: 200, y: 0 },
      { id: "n2", title: "B", type: "comet", x: 0, y: 100 },
    ];
    const links = [
      { source: "n0", target: "n1", weight: 0.5, link_type: "related" as const },
      { source: "n0", target: "n2", weight: 0.5, link_type: "related" as const },
    ];
    const nodeMap = new Map(nodes.map((n) => [n.id, n]));

    const hovered = nodes[0];
    for (let i = 0; i < 100; i++) {
      state.update(800, 600, { x: 0, y: 0, k: 1 }, hovered, false, links, nodeMap);
    }

    // The farthest neighbor is 200 px away, so the radius should at least cover
    // that distance plus the node/border margin, but not exceed radius_max.
    expect(state.snapshot.radius).toBeGreaterThan(200);
    expect(state.snapshot.radius).toBeLessThanOrEqual(FOG.radius_max);
  });
});
