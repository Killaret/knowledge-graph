/**
 * 2D graph fog rendering and visibility culling.
 *
 * Used for an atmospheric edge-vignette and an adaptive fog-of-war.
 */
import { graphConfig2D } from "$shared/config";
import type { SimulationNode, SimulationLink } from "./types";
import { getLinkEndpointId } from "./types";

const FOG_CONFIG = graphConfig2D.fog;

/** Renderable fog parameters, in screen coordinates. */
export interface FogRenderParams {
  /** Whether fog rendering is enabled at all. */
  enabled: boolean;
  /** Current fog mode. */
  mode: "off" | "atmospheric" | "adaptive" | "first-person";
  /** Screen X of the fog center / clear area. */
  centerX: number;
  /** Screen Y of the fog center / clear area. */
  centerY: number;
  /** Radius of the clear area, in screen pixels. */
  radius: number;
  /** Distance over which the fog fades from transparent to opaque. */
  feather: number;
  /** Fog color, including alpha. */
  color: string;
}

/** Compute the world -> screen mapping for a single node. */
function toScreen(
  node: SimulationNode,
  transform: { x: number; y: number; k: number }
): { x: number; y: number } {
  return {
    x: (node.x ?? 0) * transform.k + transform.x,
    y: (node.y ?? 0) * transform.k + transform.y,
  };
}

/**
 * Draw the fog overlay. When `mode` is `off` or `first-person`, nothing is drawn.
 * The gradient is transparent at the center (within `radius`) and fades to `color`
 * at the edges, creating a vignette or fog-of-war effect.
 */
export function drawFog(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  fog: FogRenderParams
): void {
  if (!fog.enabled || fog.mode === "off" || fog.mode === "first-person") {
    return;
  }

  const maxR = Math.max(width, height) * Math.SQRT2;
  if (fog.radius <= 0 || maxR <= 0) {
    return;
  }

  // Clamp the visual clear radius so the fog always acts as an edge vignette
  // when the culling radius is large (atmospheric mode), while still matching
  // the culling radius in adaptive mode.
  const visualInnerRadius = Math.min(fog.radius, maxR - fog.feather);
  const visualOuterRadius = Math.min(fog.radius + fog.feather, maxR);
  const innerStop = Math.max(0, visualInnerRadius / maxR);
  const outerStop = Math.min(1, visualOuterRadius / maxR);

  const gradient = ctx.createRadialGradient(
    fog.centerX,
    fog.centerY,
    0,
    fog.centerX,
    fog.centerY,
    maxR
  );
  gradient.addColorStop(0, "rgba(0, 0, 0, 0)");
  gradient.addColorStop(innerStop, "rgba(0, 0, 0, 0)");
  gradient.addColorStop(outerStop, fog.color);
  gradient.addColorStop(1, fog.color);

  ctx.save();
  ctx.globalCompositeOperation = "source-over";
  ctx.fillStyle = gradient;
  ctx.fillRect(0, 0, width, height);
  ctx.restore();
}

/**
 * Build the set of node ids that should be rendered despite the fog.
 * Includes:
 * - the hovered node,
 * - all nodes directly linked to the hovered node,
 * - every node within `fog.radius` of the fog center.
 */
export function createFogVisibilitySet(
  nodes: SimulationNode[],
  links: SimulationLink[],
  transform: { x: number; y: number; k: number },
  fog: FogRenderParams,
  hoveredNodeId: string | null = null
): Set<string> {
  if (!fog.enabled || fog.mode === "off" || fog.mode === "first-person") {
    return new Set(nodes.map((n) => n.id));
  }

  const visible = new Set<string>();

  // Hovered node and its direct neighbors are always revealed.
  if (hoveredNodeId) {
    visible.add(hoveredNodeId);
    for (const link of links) {
      const sourceId = getLinkEndpointId(link.source);
      const targetId = getLinkEndpointId(link.target);
      if (sourceId === hoveredNodeId) {
        visible.add(targetId);
      } else if (targetId === hoveredNodeId) {
        visible.add(sourceId);
      }
    }
  }

  // Nodes within the clear radius are visible.
  for (const node of nodes) {
    if (node.x == null || node.y == null) continue;
    const s = toScreen(node, transform);
    const dx = s.x - fog.centerX;
    const dy = s.y - fog.centerY;
    if (dx * dx + dy * dy <= fog.radius * fog.radius) {
      visible.add(node.id);
    }
  }

  return visible;
}

/** Whether a node is currently hidden by the fog. */
export function isNodeHiddenByFog(
  node: SimulationNode,
  transform: { x: number; y: number; k: number },
  fog: FogRenderParams,
  visibleSet: Set<string>
): boolean {
  if (!fog.enabled || fog.mode === "off" || fog.mode === "first-person") {
    return false;
  }
  return !visibleSet.has(node.id);
}

/** Default fog parameters: fully clear / off. */
export function defaultFogRenderParams(): FogRenderParams {
  return {
    enabled: FOG_CONFIG.enabled,
    mode: "off",
    centerX: 0,
    centerY: 0,
    radius: FOG_CONFIG.radius_max,
    feather: FOG_CONFIG.edge_feather,
    color: FOG_CONFIG.color,
  };
}
