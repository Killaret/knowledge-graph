import * as THREE from "three";
import type { FogPresetName } from "../model/types";

export interface FogPresetConfig {
  density: number;
}

export interface FogConfig {
  fogDensityInitial: number;
  fogDensityFinal: number;
  fog_presets?: Partial<Record<FogPresetName, FogPresetConfig>>;
}

const PRESET_ORDER: FogPresetName[] = ["birth", "nebula", "deep-space"];

function resolvePresetDensity(preset: FogPresetName, config: FogConfig): number {
  const override = config.fog_presets?.[preset];
  if (override) {
    return override.density;
  }

  switch (preset) {
    case "birth":
      return config.fogDensityInitial;
    case "nebula":
      return (config.fogDensityInitial + config.fogDensityFinal) / 2;
    case "deep-space":
    default:
      return 0;
  }
}

/**
 * Apply a fog preset to a 3D scene.
 * Ensures scene.fog is a FogExp2 instance and sets its density.
 */
export function applyFogPreset(
  scene: THREE.Scene,
  preset: FogPresetName,
  config: FogConfig
): number {
  const density = resolvePresetDensity(preset, config);
  const backgroundColor = scene.background instanceof THREE.Color ? scene.background : new THREE.Color(0x050510);

  if (!scene.fog || !(scene.fog instanceof THREE.FogExp2)) {
    scene.fog = new THREE.FogExp2(backgroundColor, density);
  }

  scene.fog.density = density;
  return density;
}

/** Get the next lower-quality preset, or undefined if already at the lowest. */
export function getLowerFogPreset(current: FogPresetName): FogPresetName | undefined {
  const index = PRESET_ORDER.indexOf(current);
  if (index >= PRESET_ORDER.length - 1) return undefined;
  return PRESET_ORDER[index + 1];
}

/** Get the next higher-quality preset, or undefined if already at the highest. */
export function getHigherFogPreset(current: FogPresetName): FogPresetName | undefined {
  const index = PRESET_ORDER.indexOf(current);
  if (index <= 0) return undefined;
  return PRESET_ORDER[index - 1];
}
