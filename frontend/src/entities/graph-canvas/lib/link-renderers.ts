/**
 * Canvas link renderers for GraphCanvas
 */
import { graphConfig2D } from "$shared/config";
import { LinkType } from "$entities";
import { getLinkEndpointId, type SimulationNode, type SimulationLink } from "./types";

// Performance thresholds — sourced from knowledge-graph.config.json → frontend.graph.2d
const PERFORMANCE_THRESHOLD_LINKS = graphConfig2D.animated_links_threshold;

/** Pixel offset used to separate bidirectional links into curved pairs. */
export const BIDIRECTIONAL_LINK_OFFSET = 24;

/**
 * Draw a quadratic bezier path between two nodes, optionally curving it
 * perpendicular to the straight segment by `curveOffset`.
 */
export function drawCurvedLinkPath(
  ctx: CanvasRenderingContext2D,
  source: { x?: number; y?: number },
  target: { x?: number; y?: number },
  curveOffset: number
): void {
  const sx = source.x!;
  const sy = source.y!;
  const tx = target.x!;
  const ty = target.y!;

  const dx = tx - sx;
  const dy = ty - sy;
  const len = Math.sqrt(dx * dx + dy * dy) || 1;
  const midX = (sx + tx) / 2;
  const midY = (sy + ty) / 2;

  if (curveOffset === 0) {
    ctx.moveTo(sx, sy);
    ctx.lineTo(tx, ty);
    return;
  }

  // Perpendicular vector (rotated 90° counter-clockwise)
  const perpX = (-dy / len) * curveOffset;
  const perpY = (dx / len) * curveOffset;

  ctx.moveTo(sx, sy);
  ctx.quadraticCurveTo(midX + perpX, midY + perpY, tx, ty);
}

/**
 * Detect links that have a reverse counterpart and build a map
 * from pair key to the number of reverse links.
 */
export function buildBidirectionalPairSet(links: SimulationLink[]): Set<string> {
  const pairKeys = new Set<string>();
  const reverseKeys = new Set<string>();

  for (const link of links) {
    const sourceId = getLinkEndpointId(link.source);
    const targetId = getLinkEndpointId(link.target);
    const [a, b] = sourceId < targetId ? [sourceId, targetId] : [targetId, sourceId];
    const pairKey = `${a}|${b}`;
    const reverseKey = `${targetId}|${sourceId}`;

    if (pairKeys.has(reverseKey)) {
      reverseKeys.add(pairKey);
    }
    pairKeys.add(pairKey);
  }

  return reverseKeys;
}

/**
 * Draw animated link with moving dots
 */
export function drawAnimatedLink(
  ctx: CanvasRenderingContext2D,
  link: SimulationLink,
  nodes: Map<string, SimulationNode>,
  time: number,
  linkCount: number,
  hoveredNodeId?: string | null,
  curveOffset: number = 0,
  baseOpacity: number = 1,
  isDuplicateHighlighted: boolean = false
): void {
  const sourceId = getLinkEndpointId(link.source);
  const targetId = getLinkEndpointId(link.target);

  const source = nodes.get(sourceId);
  const target = nodes.get(targetId);
  if (
    !source ||
    !target ||
    source.x == null ||
    source.y == null ||
    target.x == null ||
    target.y == null
  ) {
    return;
  }

  if (linkCount > PERFORMANCE_THRESHOLD_LINKS) {
    // Fallback to static link for performance
    drawLink(
      ctx,
      link,
      source,
      target,
      baseOpacity,
      hoveredNodeId,
      isDuplicateHighlighted,
      curveOffset
    );
    return;
  }

  // Check if this link should be highlighted
  const isHovered = hoveredNodeId && (sourceId === hoveredNodeId || targetId === hoveredNodeId);
  let opacity = hoveredNodeId ? (isHovered ? 1 : 0.3) : baseOpacity;
  const weight = link.weight ?? 0.5;
  const linkType = LinkType.fromString(link.link_type);
  const dashArray = linkType.getLineDash(weight);

  ctx.beginPath();
  drawCurvedLinkPath(ctx, source, target, curveOffset);

  const lineWidth = Math.max(1, weight * 4) * (isDuplicateHighlighted ? 1.5 : 1);
  ctx.lineWidth = lineWidth;

  if (isDuplicateHighlighted) {
    const pulseOpacity = 0.5 + 0.5 * Math.abs(Math.sin(time / 150));
    opacity = opacity * pulseOpacity;
    ctx.shadowBlur = 15;
    ctx.shadowColor = "rgba(255, 204, 0, 0.8)";
    ctx.strokeStyle = `rgba(255, 204, 0, ${opacity})`;
  } else {
    ctx.strokeStyle = linkType.getColor(weight, opacity);
  }

  ctx.setLineDash(dashArray);
  ctx.stroke();

  // Animate dash offset for moving dots effect
  if (dashArray.length > 0) {
    const speed = 0.5 + weight * 0.5;
    const offset = (time * speed) % 20;
    ctx.lineDashOffset = -offset;
    ctx.stroke();
  }

  // Reset line dash
  ctx.setLineDash([]);
  ctx.lineDashOffset = 0;
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
}

/**
 * Draw a static link (fallback)
 */
export function drawLink(
  ctx: CanvasRenderingContext2D,
  link: SimulationLink,
  sourceNode: SimulationNode,
  targetNode: SimulationNode,
  opacity: number = 1,
  hoveredNodeId?: string | null,
  isDuplicateHighlighted?: boolean,
  curveOffset: number = 0
): void {
  // Check if this link should be highlighted
  const isHovered =
    hoveredNodeId && (link.source === hoveredNodeId || link.target === hoveredNodeId);
  const finalOpacity = hoveredNodeId ? (isHovered ? 1 : 0.3) : opacity;

  ctx.beginPath();
  drawCurvedLinkPath(ctx, sourceNode, targetNode, curveOffset);

  const weight = link.weight ?? 0.5;
  const linkType = LinkType.fromString(link.link_type);

  const lineWidth = Math.max(1, weight * 4) * (isDuplicateHighlighted ? 1.5 : 1);
  ctx.lineWidth = lineWidth;
  ctx.strokeStyle = isDuplicateHighlighted
    ? `rgba(255, 204, 0, ${finalOpacity})`
    : linkType.getColor(weight, finalOpacity);

  if (isDuplicateHighlighted) {
    ctx.shadowBlur = 15;
    ctx.shadowColor = "rgba(255, 204, 0, 0.8)";
  }

  const dash = linkType.getLineDash(weight);
  if (dash.length > 0) {
    ctx.setLineDash(dash);
  }
  ctx.stroke();
  ctx.setLineDash([]);
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
}

/**
 * Draw a preview link during drag-and-drop
 */
export function drawPreviewLink(
  ctx: CanvasRenderingContext2D,
  sourceX: number,
  sourceY: number,
  targetX: number,
  targetY: number,
  opacity: number = 0.6
): void {
  ctx.beginPath();
  ctx.moveTo(sourceX, sourceY);
  ctx.lineTo(targetX, targetY);

  ctx.lineWidth = 2;
  ctx.strokeStyle = `rgba(255, 204, 0, ${opacity})`;
  ctx.setLineDash([5, 5]);
  ctx.stroke();
  ctx.setLineDash([]);

  // Draw target indicator
  ctx.beginPath();
  ctx.arc(targetX, targetY, 15, 0, Math.PI * 2);
  ctx.strokeStyle = `rgba(255, 204, 0, ${opacity})`;
  ctx.lineWidth = 2;
  ctx.stroke();

  ctx.beginPath();
  ctx.arc(targetX, targetY, 8, 0, Math.PI * 2);
  ctx.fillStyle = `rgba(255, 204, 0, ${opacity})`;
  ctx.fill();
}
