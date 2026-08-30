/**
 * Fog-of-war state for the 2D graph canvas.
 *
 * Drives both an atmospheric edge vignette and an adaptive performance fog.
 * Tracks FPS and expands/contracts the visible radius to keep the renderer
 * comfortable. Exposes a warning flag that UI can display when the fog closes
 * due to critically low FPS.
 */
import { graphConfig2D } from "$shared/config";
import { createPerformanceMonitor } from "$shared/lib/performance-monitor";
import { getHoveredNeighborIds } from "$entities/graph-canvas/lib/fog";
import { BASE_NODE_RADIUS } from "$entities/graph-canvas/lib/graph-constants";
import type {
  SimulationLink,
  SimulationNode,
  TransformState,
} from "$entities/graph-canvas/lib/types";

const FOG = graphConfig2D.fog;

export type FogMode = "off" | "atmospheric" | "adaptive" | "first-person";

export interface FogRenderSnapshot {
  enabled: boolean;
  mode: FogMode;
  centerX: number;
  centerY: number;
  radius: number;
  feather: number;
  color: string;
  fps: number;
  showWarning: boolean;
}

export interface FogState {
  /** Current fog snapshot, derived from reactive state. */
  readonly snapshot: FogRenderSnapshot;
  /** Whether fog is enabled (user toggle). */
  readonly enabled: boolean;
  /** Show red performance warning banner. */
  readonly showWarning: boolean;
  /** Call once per animation frame with the rAF timestamp. */
  tick(timestamp: number): void;
  /** Update center/radius/mode based on frame state. */
  update(
    width: number,
    height: number,
    transform: TransformState,
    hoveredNode: SimulationNode | null,
    focusMode: boolean,
    links?: SimulationLink[],
    nodeMap?: Map<string, SimulationNode>
  ): void;
  /** Toggle fog on/off. */
  toggle(): void;
}

export function createFogState(): FogState {
  let enabled = $state(FOG.enabled);
  let mode = $state<FogMode>(FOG.enabled ? "atmospheric" : "off");
  let centerX = $state(0);
  let centerY = $state(0);
  let currentRadius = $state(FOG.radius_max);
  let targetRadius = $state(FOG.radius_max);
  let fps = $state(60);
  let showWarning = $state(false);

  const monitor = createPerformanceMonitor();

  function tick(timestamp: number) {
    monitor.tick(timestamp);
    fps = monitor.fps;
  }

  function computeCenter(
    width: number,
    height: number,
    transform: TransformState,
    hoveredNode: SimulationNode | null
  ) {
    let targetX = width / 2;
    let targetY = height / 2;

    if (hoveredNode && hoveredNode.x != null && hoveredNode.y != null) {
      targetX = hoveredNode.x * transform.k + transform.x;
      targetY = hoveredNode.y * transform.k + transform.y;
    }

    // Smooth center movement (~90% settling over 20 frames).
    centerX += (targetX - centerX) * 0.15;
    centerY += (targetY - centerY) * 0.15;
  }

  function computeHoverRadius(
    hoveredNode: SimulationNode,
    hoveredNeighborIds: Set<string>,
    nodeMap: Map<string, SimulationNode>,
    transform: TransformState
  ): number {
    const hoveredScreenX = (hoveredNode.x ?? 0) * transform.k + transform.x;
    const hoveredScreenY = (hoveredNode.y ?? 0) * transform.k + transform.y;

    let maxSq = 0;
    for (const neighborId of hoveredNeighborIds) {
      const neighbor = nodeMap.get(neighborId);
      if (!neighbor || neighbor.x == null || neighbor.y == null) continue;

      const sx = neighbor.x * transform.k + transform.x - hoveredScreenX;
      const sy = neighbor.y * transform.k + transform.y - hoveredScreenY;
      const d2 = sx * sx + sy * sy;
      if (d2 > maxSq) maxSq = d2;
    }

    // Make sure the hovered node and all of its direct neighbors fit inside the
    // clear area plus a small margin so the fog does not darken them.
    const neighborMargin = BASE_NODE_RADIUS * 2 * transform.k + FOG.edge_feather;
    return Math.sqrt(maxSq) + neighborMargin;
  }

  function adaptRadius(
    focusMode: boolean,
    hoveredNode: SimulationNode | null,
    hoveredNeighborIds: Set<string>,
    nodeMap: Map<string, SimulationNode>,
    transform: TransformState
  ) {
    if (!enabled || focusMode) {
      targetRadius = FOG.radius_max;
      if (focusMode) {
        mode = "first-person";
      } else {
        mode = "off";
      }
    } else if (fps <= FOG.fps_low) {
      mode = "adaptive";
      targetRadius = FOG.radius_min;
    } else if (fps >= FOG.fps_high) {
      targetRadius = FOG.radius_max;
      // Once fully opened and stable, treat as atmospheric visual.
      if (Math.abs(currentRadius - FOG.radius_max) < FOG.edge_feather) {
        mode = "atmospheric";
      }
    }
    // In the fps_low..fps_high band we keep the current target to avoid oscillation.

    // When a node is hovered, expand the clear radius so the hovered node and its
    // direct neighbors are not darkened by the fog overlay.
    if (enabled && !focusMode && hoveredNode && hoveredNeighborIds.size > 0) {
      const hoverRadius = computeHoverRadius(hoveredNode, hoveredNeighborIds, nodeMap, transform);
      targetRadius = Math.max(targetRadius, Math.min(hoverRadius, FOG.radius_max));
    }

    // Smooth radius transition based on transition_ms.
    // Approximate 1 - exp(-4 * dt / transition); dt ~ 16.67ms.
    const alpha = 1 - Math.exp(-4 * (16.67 / FOG.transition_ms));
    currentRadius += (targetRadius - currentRadius) * alpha;
    currentRadius = Math.max(FOG.radius_min, Math.min(FOG.radius_max, currentRadius));
  }

  function update(
    width: number,
    height: number,
    transform: TransformState,
    hoveredNode: SimulationNode | null,
    focusMode: boolean,
    links: SimulationLink[] = [],
    nodeMap: Map<string, SimulationNode> = new Map()
  ) {
    computeCenter(width, height, transform, hoveredNode);
    const hoveredNeighborIds = getHoveredNeighborIds(hoveredNode?.id ?? null, links);
    adaptRadius(focusMode, hoveredNode, hoveredNeighborIds, nodeMap, transform);

    // Warning shows only in adaptive mode with critically low FPS.
    showWarning = mode === "adaptive" && fps < FOG.warning_threshold;
  }

  function toggle() {
    enabled = !enabled;
    if (!enabled) {
      mode = "off";
      targetRadius = FOG.radius_max;
    } else {
      mode = "atmospheric";
      targetRadius = FOG.radius_max;
    }
  }

  return {
    get enabled() {
      return enabled;
    },
    get showWarning() {
      return showWarning;
    },
    get snapshot() {
      return {
        enabled,
        mode,
        centerX,
        centerY,
        radius: currentRadius,
        feather: FOG.edge_feather,
        color: FOG.color,
        fps,
        showWarning,
      };
    },
    tick,
    update,
    toggle,
  };
}
