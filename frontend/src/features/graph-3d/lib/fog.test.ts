import { describe, it, expect } from "vitest";
import * as THREE from "three";
import { applyFogPreset, getLowerFogPreset, getHigherFogPreset } from "./fog";
import type { FogConfig } from "./fog";

const baseConfig: FogConfig = {
  fogDensityInitial: 0.08,
  fogDensityFinal: 0.005,
};

describe("applyFogPreset", () => {
  it("applies the birth preset using fogDensityInitial", () => {
    const scene = new THREE.Scene();
    const density = applyFogPreset(scene, "birth", baseConfig);

    expect(scene.fog).toBeInstanceOf(THREE.FogExp2);
    expect((scene.fog as THREE.FogExp2).density).toBeCloseTo(0.08, 5);
    expect(density).toBeCloseTo(0.08, 5);
  });

  it("applies the nebula preset with an intermediate density", () => {
    const scene = new THREE.Scene();
    const density = applyFogPreset(scene, "nebula", baseConfig);

    expect((scene.fog as THREE.FogExp2).density).toBeCloseTo(0.0425, 5);
    expect(density).toBeCloseTo(0.0425, 5);
  });

  it("applies the deep-space preset with density 0", () => {
    const scene = new THREE.Scene();
    const density = applyFogPreset(scene, "deep-space", baseConfig);

    expect((scene.fog as THREE.FogExp2).density).toBeCloseTo(0, 5);
    expect(density).toBeCloseTo(0, 5);
  });

  it("uses fog_presets overrides when provided", () => {
    const scene = new THREE.Scene();
    const config: FogConfig = {
      ...baseConfig,
      fog_presets: {
        birth: { density: 0.1 },
        nebula: { density: 0.03 },
        "deep-space": { density: 0.001 },
      },
    };

    expect(applyFogPreset(scene, "birth", config)).toBeCloseTo(0.1, 5);
    expect(applyFogPreset(scene, "nebula", config)).toBeCloseTo(0.03, 5);
    expect(applyFogPreset(scene, "deep-space", config)).toBeCloseTo(0.001, 5);
  });

  it("creates FogExp2 if scene has no fog", () => {
    const scene = new THREE.Scene();
    expect(scene.fog).toBeNull();
    applyFogPreset(scene, "birth", baseConfig);
    expect(scene.fog).toBeInstanceOf(THREE.FogExp2);
  });

  it("reuses an existing FogExp2 instance", () => {
    const scene = new THREE.Scene();
    scene.fog = new THREE.FogExp2(0xffffff, 0.5);
    const original = scene.fog;

    applyFogPreset(scene, "nebula", baseConfig);

    expect(scene.fog).toBe(original);
  });
});

describe("preset navigation helpers", () => {
  it("returns the next lower preset", () => {
    expect(getLowerFogPreset("birth")).toBe("nebula");
    expect(getLowerFogPreset("nebula")).toBe("deep-space");
    expect(getLowerFogPreset("deep-space")).toBeUndefined();
  });

  it("returns the next higher preset", () => {
    expect(getHigherFogPreset("deep-space")).toBe("nebula");
    expect(getHigherFogPreset("nebula")).toBe("birth");
    expect(getHigherFogPreset("birth")).toBeUndefined();
  });
});
