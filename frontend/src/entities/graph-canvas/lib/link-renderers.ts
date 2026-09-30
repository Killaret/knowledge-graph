/**
 * Canvas link renderers for GraphCanvas
 */
import { graphConfig2D } from "$shared/config";
import { LinkType, getAutoLinkColor } from "$entities";
import { getLinkEndpointId, type SimulationNode, type SimulationLink } from "./types";
import { chainDepthOpacity, dependencyLinkKey, type DependencyChain } from "./dependency-chain";

// Performance thresholds — sourced from knowledge-graph.config.json → frontend.graph.2d
const PERFORMANCE_THRESHOLD_LINKS = graphConfig2D.animated_links_threshold;

/** Pixel offset used to separate bidirectional links into curved pairs. */
export const BIDIRECTIONAL_LINK_OFFSET = 24;

/**
 * UI-GRAPH-1: model-suggested links (source_type "gamma") render thinner and
 * more transparent than user-created ones so the hand-made structure reads
 * first on a dense graph.
 */
export const AUTO_LINK_WIDTH_FACTOR = 0.55;
export const AUTO_LINK_OPACITY_FACTOR = 0.45;

/** Colour of dependency-cycle links — red so loops read as warnings. */
export const DEPENDENCY_CYCLE_COLOR = "#ef4444";
/** Opacity of everything outside the highlighted dependency chain. */
export const DEPENDENCY_CHAIN_DIMMED_OPACITY = 0.15;

export function isAutoLink(link: SimulationLink): boolean {
  return link.source_type === "gamma";
}

/** "#rrggbb" → rgba() string with the given opacity. */
export function hexToRgba(hex: string, opacity: number): string {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${opacity})`;
}

/**
 * LINK-TYPES-1: stroke colour for a link. Model-suggested (gamma) links use
 * the dedicated auto-link colour — never the manual type colour — while
 * manual links keep their type colour. Brightness follows the weight.
 */
export function linkStrokeColor(
  link: SimulationLink,
  linkType: LinkType,
  weight: number,
  opacity: number
): string {
  if (isAutoLink(link)) {
    return getAutoLinkColor(weight, opacity);
  }
  return linkType.getColor(weight, opacity);
}

/**
 * Resolve the opacity of a link under the dependency-chain highlight. When a
 * chain is active it overrides the regular neighbour-dimming: chain members
 * fade with their BFS distance from the hovered node, everything else dims.
 */
export function chainLinkOpacity(
  depChain: DependencyChain,
  sourceId: string,
  targetId: string
): number | null {
  const depth = depChain.linkDepth.get(dependencyLinkKey(sourceId, targetId));
  return depth === undefined ? DEPENDENCY_CHAIN_DIMMED_OPACITY : chainDepthOpacity(depth);
}

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
  isDuplicateHighlighted: boolean = false,
  hoveredNeighborIds?: Set<string>,
  depChain?: DependencyChain | null
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
      curveOffset,
      hoveredNeighborIds,
      depChain
    );
    return;
  }

  // Check if this link should be highlighted
  const isHovered = hoveredNodeId && (sourceId === hoveredNodeId || targetId === hoveredNodeId);
  const isNeighborLink =
    hoveredNodeId && hoveredNeighborIds
      ? (hoveredNeighborIds.has(sourceId) && hoveredNeighborIds.has(targetId)) ||
        (hoveredNeighborIds.has(sourceId) && sourceId === hoveredNodeId) ||
        (hoveredNeighborIds.has(targetId) && targetId === hoveredNodeId)
      : false;
  let opacity = hoveredNodeId ? (isHovered ? 1 : isNeighborLink ? 0.7 : 0.3) : baseOpacity;
  if (depChain) {
    // LINK-TYPES-1: a dependency chain is highlighted — its links fade with
    // the distance from the hovered node, everything else dims hard.
    opacity = chainLinkOpacity(depChain, sourceId, targetId) ?? opacity;
  }
  if (isAutoLink(link)) opacity *= AUTO_LINK_OPACITY_FACTOR;
  const weight = link.weight ?? 0.5;
  const linkType = LinkType.fromString(link.link_type);
  const dashArray = linkType.getLineDash(weight);
  const isCycleLink = depChain?.cycleLinks.has(dependencyLinkKey(sourceId, targetId)) ?? false;

  ctx.beginPath();
  drawCurvedLinkPath(ctx, source, target, curveOffset);

  const lineWidth =
    Math.max(1, weight * 4) *
    (isDuplicateHighlighted ? 1.5 : 1) *
    (isAutoLink(link) ? AUTO_LINK_WIDTH_FACTOR : 1);
  ctx.lineWidth = lineWidth;

  if (isDuplicateHighlighted) {
    const pulseOpacity = 0.5 + 0.5 * Math.abs(Math.sin(time / 150));
    opacity = opacity * pulseOpacity;
    ctx.shadowBlur = 15;
    ctx.shadowColor = "rgba(255, 204, 0, 0.8)";
    ctx.strokeStyle = `rgba(255, 204, 0, ${opacity})`;
  } else if (isCycleLink) {
    // Dependency cycle — the loop reads as a warning in red.
    ctx.strokeStyle = hexToRgba(DEPENDENCY_CYCLE_COLOR, opacity);
  } else {
    ctx.strokeStyle = linkStrokeColor(link, linkType, weight, opacity);
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
  curveOffset: number = 0,
  hoveredNeighborIds?: Set<string>,
  depChain?: DependencyChain | null
): void {
  // Check if this link should be highlighted
  const sourceId = getLinkEndpointId(link.source);
  const targetId = getLinkEndpointId(link.target);
  const isHovered = hoveredNodeId && (sourceId === hoveredNodeId || targetId === hoveredNodeId);
  const isNeighborLink =
    hoveredNodeId && hoveredNeighborIds
      ? hoveredNeighborIds.has(sourceId) && hoveredNeighborIds.has(targetId)
      : false;
  let finalOpacity = hoveredNodeId ? (isHovered ? 1 : isNeighborLink ? 0.7 : 0.3) : opacity;
  if (depChain) {
    finalOpacity = chainLinkOpacity(depChain, sourceId, targetId) ?? finalOpacity;
  }
  if (isAutoLink(link)) finalOpacity *= AUTO_LINK_OPACITY_FACTOR;

  ctx.beginPath();
  drawCurvedLinkPath(ctx, sourceNode, targetNode, curveOffset);

  const weight = link.weight ?? 0.5;
  const linkType = LinkType.fromString(link.link_type);
  const isCycleLink = depChain?.cycleLinks.has(dependencyLinkKey(sourceId, targetId)) ?? false;

  const lineWidth =
    Math.max(1, weight * 4) *
    (isDuplicateHighlighted ? 1.5 : 1) *
    (isAutoLink(link) ? AUTO_LINK_WIDTH_FACTOR : 1);
  ctx.lineWidth = lineWidth;
  ctx.strokeStyle = isDuplicateHighlighted
    ? `rgba(255, 204, 0, ${finalOpacity})`
    : isCycleLink
      ? hexToRgba(DEPENDENCY_CYCLE_COLOR, finalOpacity)
      : linkStrokeColor(link, linkType, weight, finalOpacity);

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
