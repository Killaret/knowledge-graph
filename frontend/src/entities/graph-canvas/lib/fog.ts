/**
 * 2D graph fog rendering and visibility culling.
 *
 * Used for an atmospheric edge-vignette and an adaptive fog-of-war.
 */
import { graphConfig2D } from "$shared/config";
import { BASE_NODE_RADIUS } from "./graph-constants";
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
 * Build the set of node ids that should be rendered.
 *
 * Applies two layers:
 * 1. **Viewport culling** — nodes outside the visible canvas (with a margin)
 *    are never drawn.
 * 2. **Fog culling** — in `atmospheric` or `adaptive` mode, nodes outside the
 *    fog clear radius are hidden, unless they are the hovered node or a
 *    direct neighbor of it.
 *
 * `first-person` mode returns all nodes without culling.
 */
export function createFogVisibilitySet(
  nodes: SimulationNode[],
  links: SimulationLink[],
  width: number,
  height: number,
  transform: { x: number; y: number; k: number },
  fog: FogRenderParams,
  hoveredNodeId: string | null = null,
  nodeMap?: Map<string, SimulationNode>
): Set<string> {
  if (fog.mode === "first-person") {
    return new Set(nodes.map((n) => n.id));
  }

  const resolvedNodeMap = nodeMap ?? new Map<string, SimulationNode>();
  if (!nodeMap) {
    for (const node of nodes) {
      if (node.id) {
        resolvedNodeMap.set(node.id, node);
      }
    }
  }

  // Pre-compute screen positions for all nodes with coordinates.
  const screenById = new Map<string, { x: number; y: number }>();
  for (const node of nodes) {
    if (node.x != null && node.y != null) {
      screenById.set(node.id, toScreen(node, transform));
    }
  }

  // Viewport bounds with a small margin so nodes at the edge are not clipped.
  const margin = BASE_NODE_RADIUS * 2 * transform.k;
  const viewportLeft = -margin;
  const viewportRight = width + margin;
  const viewportTop = -margin;
  const viewportBottom = height + margin;

  function isInViewport(nodeId: string): boolean {
    const s = screenById.get(nodeId);
    if (!s) return false;
    return (
      s.x >= viewportLeft && s.x <= viewportRight && s.y >= viewportTop && s.y <= viewportBottom
    );
  }

  function isInFog(nodeId: string): boolean {
    const s = screenById.get(nodeId);
    if (!s) return false;
    const dx = s.x - fog.centerX;
    const dy = s.y - fog.centerY;
    return dx * dx + dy * dy <= fog.radius * fog.radius;
  }

  const visible = new Set<string>();

  if (fog.mode === "off") {
    // Viewport culling only.
    for (const node of nodes) {
      if (isInViewport(node.id)) {
        visible.add(node.id);
      }
    }
    return visible;
  }

  // Hovered node and its direct neighbors are always revealed if in viewport.
  if (hoveredNodeId) {
    const hoveredNode = resolvedNodeMap.get(hoveredNodeId);
    if (hoveredNode && isInViewport(hoveredNodeId)) {
      visible.add(hoveredNodeId);
      for (const link of links) {
        const sourceId = getLinkEndpointId(link.source);
        const targetId = getLinkEndpointId(link.target);
        const neighborId = sourceId === hoveredNodeId ? targetId : sourceId;
        if (neighborId && isInViewport(neighborId)) {
          visible.add(neighborId);
        }
      }
    }
  }

  // Nodes within the fog clear radius are visible.
  for (const node of nodes) {
    if (isInViewport(node.id) && isInFog(node.id)) {
      visible.add(node.id);
    }
  }

  return visible;
}

/** Whether a node is currently hidden by the fog. */
export function isNodeHiddenByFog(
  fog: FogRenderParams,
  visibleSet: Set<string>,
  node: SimulationNode
): boolean {
  if (fog.mode === "off" || fog.mode === "first-person") {
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
