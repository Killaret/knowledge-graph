import type { GraphNode, GraphLink } from "$shared/api/graph";

export interface SimulationNode extends GraphNode {
  x: number;
  y: number;
  z: number;
  vx?: number;
  vy?: number;
  vz?: number;
  fx?: number | null;
  fy?: number | null;
  fz?: number | null;
  index?: number;
}

export interface Graph3DCallbacks {
  onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
  onNodeDoubleClick?: (node: { id: string; title: string; type?: string }) => void;
  onReady?: () => void;
  onLoadingChange?: (loading: boolean) => void;
  onError?: (message: string) => void;
}

export type FogPresetName = "birth" | "nebula" | "deep-space";

export interface FogPreset {
  density_initial: number;
  density_final: number;
}

export interface Graph3DConfig {
  maxNodes: number;
  baseNodeScale: number;
  labelScale: number;
  linkOpacity: number;
  fogDensityInitial: number;
  fogDensityFinal: number;
  warmStartTicks: number;
  enableLabels: boolean;
  autoRotate: boolean;
  disableAnimation: boolean;
  defaultFogPreset: FogPresetName;
  fog_presets?: Partial<Record<FogPresetName, FogPreset>>;
}

export const DEFAULT_GRAPH3D_CONFIG: Graph3DConfig = {
  maxNodes: 500,
  baseNodeScale: 1.5,
  labelScale: 0.8,
  linkOpacity: 0.8,
  fogDensityInitial: 0.01,
  fogDensityFinal: 0.0006,
  warmStartTicks: 80,
  enableLabels: true,
  autoRotate: false,
  disableAnimation: false,
  defaultFogPreset: "birth",
};

export type { GraphNode, GraphLink };
