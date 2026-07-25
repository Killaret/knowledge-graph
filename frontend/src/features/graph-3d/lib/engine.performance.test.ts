import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { Graph3DEngine } from "./engine";
import type { GraphNode, GraphLink } from "../model/types";

const nodes: GraphNode[] = [
  { id: "1", title: "A", type: "star", x: 0, y: 0, z: 0 },
  { id: "2", title: "B", type: "planet", x: 1, y: 1, z: 1 },
];

const links: GraphLink[] = [{ source: "1", target: "2", weight: 0.5, link_type: "related" }];

function createContainer() {
  const container = document.createElement("div");
  container.style.width = "300px";
  container.style.height = "300px";
  return container;
}

function stubRenderer(engine: Graph3DEngine) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const bundle = (engine as any).sceneBundle;
  bundle.renderer.render = vi.fn();
  bundle.labelRenderer.render = vi.fn();
  bundle.controls.update = vi.fn();
  return bundle;
}

describe("Graph3DEngine performance adaptation", () => {
  let nowMock: ReturnType<typeof vi.spyOn>;
  let now = 0;

  beforeEach(() => {
    now = 0;
    nowMock = vi.spyOn(performance, "now").mockImplementation(() => now);
  });

  afterEach(() => {
    nowMock.mockRestore();
  });

  it("initializes with birth preset and autoRotate enabled when configured", () => {
    const container = createContainer();
    const engine = new Graph3DEngine(container, {}, { autoRotate: true });
    stubRenderer(engine);
    engine.setData(nodes, links);

    const bundle = (engine as any).sceneBundle;
    expect((bundle.scene.fog as { density: number }).density).toBeCloseTo(0.08, 5);
    expect(bundle.controls.autoRotate).toBe(true);
    engine.dispose();
  });

  it("downgrades fog preset, starfield count, and disables autoRotate on sustained low FPS", () => {
    const container = createContainer();
    const engine = new Graph3DEngine(container, {}, { autoRotate: true });
    const bundle = stubRenderer(engine);
    engine.setData(nodes, links);

    // Stop the real rAF loop so we can drive frames manually.
    (engine as any).stopLoop();

    // Feed enough low-FPS frames (100ms apart => 10 fps) to exceed low_fps_sample_count.
    for (let i = 0; i < 35; i++) {
      now += 100;
      (engine as any).frame();
      (engine as any).stopLoop();
    }

    const fog = bundle.scene.fog as { density: number };
    expect(fog.density).toBeCloseTo(0.04, 1); // nebula initial (drifts toward final while sim runs)
    expect((engine as any).currentFogPreset).toBe("nebula");
    expect((engine as any).currentPerformanceLevel).toBe("medium");
    expect(bundle.controls.autoRotate).toBe(false);

    const starfieldPositions = bundle.starfield.geometry.attributes.position;
    expect(starfieldPositions.count).toBe(800);

    engine.dispose();
  });

  it("continues degrading to deep-space and low starfield count with more low FPS frames", () => {
    const container = createContainer();
    const engine = new Graph3DEngine(container, {}, { autoRotate: true });
    const bundle = stubRenderer(engine);
    engine.setData(nodes, links);

    (engine as any).stopLoop();

    // First batch: high -> medium
    for (let i = 0; i < 35; i++) {
      now += 100;
      (engine as any).frame();
      (engine as any).stopLoop();
    }

    // Second batch: medium -> low
    for (let i = 0; i < 35; i++) {
      now += 100;
      (engine as any).frame();
      (engine as any).stopLoop();
    }

    const fog = bundle.scene.fog as { density: number };
    expect(fog.density).toBeCloseTo(0, 5); // deep-space initial
    expect((engine as any).currentFogPreset).toBe("deep-space");
    expect((engine as any).currentPerformanceLevel).toBe("low");

    const starfieldPositions = bundle.starfield.geometry.attributes.position;
    expect(starfieldPositions.count).toBe(300);

    engine.dispose();
  });

  it("restores quality on sustained high FPS", () => {
    const container = createContainer();
    const engine = new Graph3DEngine(container, {}, { autoRotate: true });
    const bundle = stubRenderer(engine);
    engine.setData(nodes, links);

    (engine as any).stopLoop();

    // Drop to low.
    for (let i = 0; i < 70; i++) {
      now += 100;
      (engine as any).frame();
      (engine as any).stopLoop();
    }
    expect((engine as any).currentFogPreset).toBe("deep-space");

    // Reset the monitor and streaks so the recovery test does not have to
    // flush the 30-sample sliding window of low-FPS samples.
    (engine as any).monitor.reset();
    (engine as any).lowFpsStreak = 0;
    (engine as any).highFpsStreak = 0;

    // Now feed high FPS frames (~16ms => 60 fps) to step up to nebula.
    for (let i = 0; i < 35; i++) {
      now += 16;
      (engine as any).frame();
      (engine as any).stopLoop();
    }

    const fog = bundle.scene.fog as { density: number };
    expect(fog.density).toBeCloseTo(0.04, 1); // nebula initial (drifts toward final while sim runs)
    expect((engine as any).currentFogPreset).toBe("nebula");
    expect((engine as any).currentPerformanceLevel).toBe("medium");

    (engine as any).monitor.reset();
    (engine as any).lowFpsStreak = 0;
    (engine as any).highFpsStreak = 0;

    // Another batch to return to birth.
    for (let i = 0; i < 35; i++) {
      now += 16;
      (engine as any).frame();
      (engine as any).stopLoop();
    }

    expect(fog.density).toBeCloseTo(0.08, 1);
    expect((engine as any).currentFogPreset).toBe("birth");
    expect((engine as any).currentPerformanceLevel).toBe("high");
    expect(bundle.controls.autoRotate).toBe(true);

    engine.dispose();
  });
});
