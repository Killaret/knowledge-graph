import * as THREE from "three";
import type { FogPresetName } from "../model/types";

export interface FogPresetConfig {
  density_initial: number;
  density_final: number;
}

export interface FogConfig {
  fogDensityInitial: number;
  fogDensityFinal: number;
  fog_presets?: Partial<Record<FogPresetName, FogPresetConfig>>;
  default_preset?: FogPresetName;
}

export interface AppliedFogPreset {
  /** Density at the start of a simulation / scene transition. */
  initial: number;
  /** Target density once the scene has stabilized. */
  final: number;
}

const PRESET_ORDER: FogPresetName[] = ["birth", "nebula", "deep-space"];

function resolvePreset(preset: FogPresetName, config: FogConfig): FogPresetConfig {
  const override = config.fog_presets?.[preset];
  if (override) {
    return override;
  }

  switch (preset) {
    case "birth":
      return {
        density_initial: config.fogDensityInitial,
        density_final: config.fogDensityFinal,
      };
    case "nebula":
      return {
        density_initial: (config.fogDensityInitial + config.fogDensityFinal) / 2,
        density_final: config.fogDensityFinal,
      };
    case "deep-space":
    default:
      return { density_initial: 0, density_final: 0 };
  }
}

/**
 * Apply a fog preset to a 3D scene.
 * Ensures scene.fog is a FogExp2 instance and sets its density to the preset's
 * initial value. Returns both initial and final densities so callers can drive
 * a smooth transition as the simulation stabilizes.
 */
export function applyFogPreset(
  scene: THREE.Scene,
  preset: FogPresetName,
  config: FogConfig
): AppliedFogPreset {
  const presetConfig = resolvePreset(preset, config);
  const backgroundColor =
    scene.background instanceof THREE.Color ? scene.background : new THREE.Color(0x050510);

  if (!scene.fog || !(scene.fog instanceof THREE.FogExp2)) {
    scene.fog = new THREE.FogExp2(backgroundColor, presetConfig.density_initial);
  }

  scene.fog.density = presetConfig.density_initial;
  return { initial: presetConfig.density_initial, final: presetConfig.density_final };
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
