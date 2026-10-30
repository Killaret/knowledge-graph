import type { GraphNode, GraphLink } from "$shared/api/graph";
import { graphConfig3D } from "$shared/config/config";
import type { SimulationNode } from "./model/types";

/**
 * 3D graph runtime configuration.
 *
 * These values are consumed by Graph3DEngine and Graph3DViewer. They live in the
 * feature layer because they describe the 3D rendering/physics scenario.
 */
export interface Graph3DRuntimeConfig {
  /** Whether to use the graph-service for initial 3D coordinates instead of random D3 seeding. */
  useGraphServiceLayout: boolean;
  /** Number of warm-up ticks for the D3 force simulation (0 = start from provided positions). */
  warmStartTicks: number;
  /** Initial layout provider: d3 (client-side force simulation) or graph-service. */
  layoutProvider: "d3" | "graph-service";
}

export const DEFAULT_RUNTIME_CONFIG: Graph3DRuntimeConfig = {
  useGraphServiceLayout: false,
  warmStartTicks: 80,
  layoutProvider: "d3",
};

/**
 * Build a runtime configuration from the centralized project config.
 * Falls back to sensible defaults when values are missing.
 */
export function toRuntimeConfig(): Graph3DRuntimeConfig {
  const provider = graphConfig3D.layout_provider;
  return {
    useGraphServiceLayout: provider === "graph-service",
    warmStartTicks: DEFAULT_RUNTIME_CONFIG.warmStartTicks,
    layoutProvider: provider === "graph-service" || provider === "d3" ? provider : "d3",
  };
}

/**
 * Convert plain API nodes into simulation-ready nodes. Preserves x/y/z from the
 * graph-service when they are present; otherwise assigns random values that the
 * D3 force simulation will refine.
 */
export function toSimulationNodes(nodes: GraphNode[], _links: GraphLink[]): SimulationNode[] {
  const hasPositions = nodes.every(
    (n) => typeof n.x === "number" && typeof n.y === "number" && typeof n.z === "number"
  );

  return nodes.map((n, i) => ({
    id: n.id,
    title: n.title,
    type: n.type ?? "unknown",
    x: n.x ?? (hasPositions ? 0 : (i % 10) * 20 - 100),
    y: n.y ?? (hasPositions ? 0 : Math.floor(i / 10) * 20 - 100),
    z: n.z ?? (hasPositions ? 0 : Math.sin(i) * 50),
    size: n.size ?? 1,
  }));
}
