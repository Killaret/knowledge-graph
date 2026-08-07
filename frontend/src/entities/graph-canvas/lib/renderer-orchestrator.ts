/**
 * Rendering orchestration for GraphCanvas.
 *
 * Coordinates link, node, background, overlay and preview drawing.
 */
import { graphConfig2D } from "$shared/config";
import { CelestialBody } from "$entities";
import { getVariation } from "$shared/utils/variation";
import { BASE_NODE_RADIUS } from "./graph-constants";
import {
  getLinkEndpointId,
  resolveLinkEndpoint,
  type SimulationNode,
  type SimulationLink,
} from "./types";
import { createFogVisibilitySet, defaultFogRenderParams, type FogRenderParams } from "./fog";
import { drawBlackHole, drawBlackHoleTooltip } from "./black-hole";
import type { BlackHoleState } from "./black-hole";
import { drawGhostNodeScreen, drawGhostNodeTooltipScreen } from "./ghost-node";
import type { GhostNodeState } from "./ghost-node";
import { drawDistortedBackgroundGrid } from "./gravity-system";
import { drawBackground } from "./background";
import { registerCelestialBodyDrawers } from "./node-registration";
import {
  drawAnimatedLink,
  drawLink,
  drawPreviewLink,
  buildBidirectionalPairSet,
  BIDIRECTIONAL_LINK_OFFSET,
} from "./link-renderers";
import { isNewNode } from "./renderer-utils";

/**
 * Draw all links with animation and hover effects
 */
export function drawAllLinks(
  ctx: CanvasRenderingContext2D,
  simLinks: SimulationLink[],
  nodes: SimulationNode[],
  linkOpacity?: Map<string, number>,
  animationTime: number = 0,
  hoveredNodeId: string | null = null,
  highlightedLinkId?: string | null,
  dyingLinks: SimulationLink[] = [],
  dyingLinkOpacity: Map<string, number> = new Map(),
  nodeMap?: Map<string, SimulationNode>,
  visibleNodeIds?: Set<string>
): void {
  let drawnCount = 0;
  let skippedCount = 0;

  const bidirectionalPairs = buildBidirectionalPairSet(simLinks);
  const resolvedNodeMap = nodeMap ?? new Map<string, SimulationNode>();
  if (!nodeMap) {
    for (const node of nodes) {
      if (node.id) {
        resolvedNodeMap.set(node.id, node);
      }
    }
  }

  simLinks.forEach((link, index) => {
    // Skip links entirely outside the fog radius to avoid expensive resolution.
    const sourceId = getLinkEndpointId(link.source);
    const targetId = getLinkEndpointId(link.target);
    if (visibleNodeIds && !visibleNodeIds.has(sourceId) && !visibleNodeIds.has(targetId)) {
      skippedCount++;
      return;
    }

    const sourceNode = resolveLinkEndpoint(link.source, nodes, resolvedNodeMap);
    const targetNode = resolveLinkEndpoint(link.target, nodes, resolvedNodeMap);

    if (
      !sourceNode ||
      !targetNode ||
      sourceNode.x == null ||
      sourceNode.y == null ||
      targetNode.x == null ||
      targetNode.y == null
    ) {
      skippedCount++;
      return;
    }

    // Get opacity for this link
    const linkId = `${link.source}-${link.target}-${index}`;
    const opacity = linkOpacity?.get(linkId) ?? 1;

    // Highlight duplicate links with a yellow pulse
    const stableLinkId = `${link.source}-${link.target}-${link.link_type || "related"}`;
    const isHighlighted = highlightedLinkId === stableLinkId;

    // Apply bidirectional curve offset when a reverse link exists
    const [a, b] = sourceId < targetId ? [sourceId, targetId] : [targetId, sourceId];
    const isBidirectional = bidirectionalPairs.has(`${a}|${b}`);
    const curveOffset = isBidirectional
      ? (sourceId < targetId ? 1 : -1) * BIDIRECTIONAL_LINK_OFFSET
      : 0;

    // Use animated link drawing
    drawAnimatedLink(
      ctx,
      link,
      resolvedNodeMap,
      animationTime,
      simLinks.length,
      hoveredNodeId,
      curveOffset,
      opacity,
      isHighlighted
    );
    drawnCount++;
  });

  // Draw dying (removed) links fading out
  dyingLinks.forEach((link) => {
    const sourceId = getLinkEndpointId(link.source);
    const targetId = getLinkEndpointId(link.target);
    if (visibleNodeIds && !visibleNodeIds.has(sourceId) && !visibleNodeIds.has(targetId)) {
      return;
    }
    const sourceNode = resolveLinkEndpoint(link.source, nodes, resolvedNodeMap);
    const targetNode = resolveLinkEndpoint(link.target, nodes, resolvedNodeMap);
    if (
      !sourceNode ||
      !targetNode ||
      sourceNode.x == null ||
      sourceNode.y == null ||
      targetNode.x == null ||
      targetNode.y == null
    ) {
      return;
    }
    const linkId =
      link.id ??
      getLinkEndpointId(link.source) +
        "-" +
        getLinkEndpointId(link.target) +
        "-" +
        (link.link_type || "related");
    const opacity = dyingLinkOpacity.get(linkId) ?? 0;
    if (opacity > 0) {
      drawLink(ctx, link, sourceNode, targetNode, opacity, null, false, 0);
    }
  });

  if (import.meta.env.DEV && (drawnCount === 0 || skippedCount > 0)) {
    console.log(
      `[drawAllLinks] Total: ${simLinks.length}, Drawn: ${drawnCount}, Skipped: ${skippedCount}`
    );
  }
}

/**
 * Draw a single node based on its type
 */
export function drawNode(
  ctx: CanvasRenderingContext2D,
  node: SimulationNode,
  r: number,
  angle: number,
  enableShadows: boolean,
  disableVariation: boolean = false,
  nodeId?: string,
  nodeCount?: number,
  animationTime?: number,
  focusMode: boolean = false,
  simplified: boolean = false
): void {
  if (node.x == null || node.y == null) {
    return;
  }

  const body = CelestialBody.fromString(node.type);

  // Fast path: when zoomed out, draw a simple filled circle using the body's glow color.
  // This skips expensive per-type renderers, shadows, and animated effects.
  if (simplified && !focusMode && node.x != null && node.y != null) {
    ctx.beginPath();
    ctx.arc(node.x, node.y, r * body.baseRadius, 0, 2 * Math.PI);
    ctx.fillStyle = body.glowColor;
    ctx.fill();
    return;
  }

  // The renderer layer wires Canvas primitives to the domain object lazily.
  // This guard also makes unit tests that call vi.resetModules() more robust.
  if (!CelestialBody.STAR.drawFunction) {
    registerCelestialBodyDrawers();
  }

  // Get deterministic variation for this node (used for hue/size/phase).
  // We still apply variation in stable render mode so color/size remain deterministic,
  // but other random jitter/animation is suppressed via `disableVariation` flags.
  // Anomalies, black holes and debris have no per-node size/hue variation.
  const variation =
    body.isAnomaly || ["blackhole", "debris"].includes(body.type)
      ? undefined
      : getVariation(node.id, body.type, body.minRadius, body.maxRadius);

  // Use exact node position — random jitter caused visible flickering every frame
  let x = node.x!;
  let y = node.y!;

  // For stable render mode snap to integer pixel positions to avoid
  // subpixel anti-aliasing differences between runs/environments.
  if (disableVariation || focusMode) {
    x = Math.round(x);
    y = Math.round(y);
  }

  const effectiveEnableShadows = enableShadows && !focusMode;

  body.draw(ctx, {
    x,
    y,
    r: r * body.baseRadius,
    angle,
    nodeId: node.id,
    nodeCount: focusMode ? undefined : nodeCount,
    time: focusMode ? undefined : animationTime,
    variation,
    disableVariation,
    enableShadows: effectiveEnableShadows,
    focusMode,
  });
  ctx.shadowBlur = 0;
}

/**
 * Draw node title
 */
export function drawNodeTitle(
  ctx: CanvasRenderingContext2D,
  node: SimulationNode,
  r: number,
  opacity: number = 1,
  disableVariation: boolean = false
): void {
  // Round font size for deterministic rendering in stable mode
  const fontSize = disableVariation ? Math.round(Math.min(14, r * 0.65)) : Math.min(14, r * 0.65);
  ctx.font = `bold ${fontSize}px sans-serif`;
  ctx.textAlign = "center";
  ctx.textBaseline = "top";
  let title = node.title || "Untitled";
  if (title.length > 14) title = title.slice(0, 12) + "…";

  // Disable text shadows in stable render mode for deterministic rendering
  if (!disableVariation) {
    ctx.shadowBlur = 2;
    ctx.shadowColor = `rgba(0,0,0,${0.5 * opacity})`;
  }

  ctx.fillStyle = `rgba(255,255,255,${opacity})`;
  const titleX = disableVariation ? Math.round(node.x!) : node.x!;
  const titleY = disableVariation ? Math.round(node.y! + r + 6) : node.y! + r + 6;
  ctx.fillText(title, titleX, titleY);
  ctx.shadowBlur = 0;
}

/**
 * Get color for a node type
 */

/**
 * Draw all nodes
 */
export function drawAllNodes(
  ctx: CanvasRenderingContext2D,
  nodes: SimulationNode[],
  angles: Map<string, number>,
  enableShadows: boolean,
  nodeOpacity?: Map<string, number>,
  disableVariation: boolean = false,
  animationTime: number = 0,
  hoveredNodeId: string | null = null,
  particleSystem?: {
    initParticles: (id: string, x: number, y: number, color: string) => void;
    update: (id: string, x: number, y: number) => void;
    draw: (ctx: CanvasRenderingContext2D, id: string) => void;
    isEnabled: () => boolean;
  } | null,
  focusMode: boolean = false,
  searchMatchIds?: Set<string>,
  visibleNodeIds?: Set<string>,
  simplified: boolean = false
): void {
  const r = BASE_NODE_RADIUS;
  const nodeCount = nodes.length;

  if (disableVariation) {
    for (const node of nodes) {
      if (visibleNodeIds && !visibleNodeIds.has(node.id)) continue;
      if (!angles.has(node.id)) {
        angles.set(node.id, 0);
      }
    }
  }

  if (particleSystem?.isEnabled() && !simplified) {
    for (const node of nodes) {
      if (visibleNodeIds && !visibleNodeIds.has(node.id)) continue;
      // Use glowColor for orbit particles so they remain visible even for
      // dark body types (e.g. blackhole notes) and read as a subtle halo.
      particleSystem.initParticles(
        node.id,
        node.x || 0,
        node.y || 0,
        CelestialBody.fromString(node.type).glowColor
      );
    }
  }

  nodes.forEach((node) => {
    if (visibleNodeIds && !visibleNodeIds.has(node.id)) return;
    if (node.x == null || node.y == null) return;
    const angle = angles.get(node.id) || 0;
    const opacity = nodeOpacity?.get(node.id) ?? 1;
    const isHovered = hoveredNodeId === node.id;
    const isSearchMatch = searchMatchIds?.has(node.id) ?? false;
    const finalOpacity = hoveredNodeId ? (isHovered ? 1 : 0.3) : opacity;
    const nodeSimplified = simplified && !isHovered;

    const previousAlpha = ctx.globalAlpha;
    ctx.globalAlpha = finalOpacity;

    drawNode(
      ctx,
      node,
      r,
      angle,
      enableShadows,
      disableVariation,
      node.id,
      nodeCount,
      animationTime,
      focusMode,
      nodeSimplified
    );

    if (!nodeSimplified) {
      drawNodeTitle(ctx, node, r, finalOpacity, disableVariation);

      // Search match outline
      if (isSearchMatch && node.x != null && node.y != null) {
        ctx.save();
        ctx.beginPath();
        ctx.arc(node.x, node.y, r + 6, 0, 2 * Math.PI);
        ctx.strokeStyle = "rgba(255, 204, 0, 0.9)";
        ctx.lineWidth = 3;
        ctx.shadowBlur = 15;
        ctx.shadowColor = "rgba(255, 204, 0, 0.6)";
        ctx.stroke();
        ctx.restore();
      }

      // New note indicator (pulsing turquoise outline for 24 hours)
      // Disabled in stable render mode to keep screenshots deterministic
      if (!disableVariation && !focusMode && isNewNode(node) && node.x != null && node.y != null) {
        // Per-node phase so multiple new notes don’t pulse in lockstep.
        const nodePhase = getVariation(node.id, node.type ?? "star").phaseShift;
        const pulse = 0.5 + 0.5 * Math.abs(Math.sin((animationTime + nodePhase * 1000) / 1000));
        ctx.save();
        ctx.beginPath();
        ctx.arc(node.x, node.y, r + 10, 0, 2 * Math.PI);
        ctx.strokeStyle = `rgba(45, 212, 191, ${pulse})`;
        ctx.lineWidth = 2;
        ctx.stroke();
        ctx.restore();
      }

      if (!focusMode && particleSystem?.isEnabled() && node.x && node.y) {
        // Keep particles fixed in stable render mode for deterministic screenshots
        if (!disableVariation) {
          particleSystem.update(node.id, node.x, node.y);
        }
        particleSystem.draw(ctx, node.id);
      }
    }

    ctx.globalAlpha = previousAlpha;
  });
}

/**
 * Main draw function
 */
export function draw(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  simLinks: SimulationLink[],
  nodes: SimulationNode[],
  angles: Map<string, number>,
  transform: { x: number; y: number; k: number },
  nodeOpacity?: Map<string, number>,
  linkOpacity?: Map<string, number>,
  dyingLinks?: SimulationLink[],
  dyingLinkOpacity?: Map<string, number>,
  disableVariation: boolean = false,
  animationTime: number = 0,
  hoveredNodeId: string | null = null,
  particleSystem?: {
    initParticles: (id: string, x: number, y: number, color: string) => void;
    update: (id: string, x: number, y: number) => void;
    draw: (ctx: CanvasRenderingContext2D, id: string) => void;
    isEnabled: () => boolean;
  } | null,
  blackHole?: BlackHoleState | null,
  ghostNode?: GhostNodeState | null,
  gravitySystem?: {
    applyAttraction: (nodes: SimulationNode[]) => void;
    getDistortion: (
      x: number,
      y: number,
      nodes: SimulationNode[],
      time: number,
      maxDistance?: number
    ) => { dx: number; dy: number };
    isEnabled: (nodeCount: number) => boolean;
  } | null,
  focusMode: boolean = false,
  searchMatchIds?: string[],
  highlightedLinkId?: string | null,
  linkPreviewTarget?: { sourceId: string; targetId: string } | null,
  linkPreviewMousePos?: { sourceId: string; x: number; y: number } | null,
  fog: FogRenderParams = defaultFogRenderParams()
): void {
  ctx.clearRect(0, 0, width, height);

  // Draw background with gravity lens distortion (skipped in focus mode)
  if (!focusMode) {
    drawBackground(ctx, width, height, nodes, animationTime);
    if (gravitySystem?.isEnabled(nodes.length)) {
      drawDistortedBackgroundGrid(ctx, width, height, nodes, animationTime);
    }
  }

  // Stable render mode: reduce anti-aliasing sources and align to pixel grid
  ctx.save();
  if (disableVariation || focusMode) {
    ctx.imageSmoothingEnabled = false;
    ctx.imageSmoothingQuality = "low";
    ctx.shadowBlur = 0;
    ctx.shadowColor = "transparent";
    ctx.lineJoin = "round";
    ctx.lineCap = "round";
    ctx.translate(0.5, 0.5);
  }
  ctx.translate(transform.x, transform.y);
  ctx.scale(transform.k, transform.k);

  // Build a node id map once per frame to avoid O(N*L) lookups
  const nodeMap = new Map<string, SimulationNode>();
  for (const node of nodes) {
    if (node.id) {
      nodeMap.set(node.id, node);
    }
  }

  // Determine which nodes are visible through viewport and fog culling.
  // Focus mode shows the selected node and its neighborhood, so culling is skipped.
  const visibleNodeIds = focusMode
    ? undefined
    : createFogVisibilitySet(
        nodes,
        simLinks,
        width,
        height,
        transform,
        fog,
        hoveredNodeId,
        nodeMap
      );

  // Draw links with animation
  drawAllLinks(
    ctx,
    simLinks,
    nodes,
    linkOpacity,
    animationTime,
    hoveredNodeId,
    highlightedLinkId,
    dyingLinks,
    dyingLinkOpacity,
    nodeMap,
    visibleNodeIds
  );

  // Draw link preview if dragging for link creation
  if (linkPreviewTarget) {
    // Preview to a specific target node
    const sourceNode = nodeMap.get(linkPreviewTarget.sourceId);
    const targetNode = nodeMap.get(linkPreviewTarget.targetId);
    if (
      sourceNode &&
      targetNode &&
      sourceNode.x != null &&
      sourceNode.y != null &&
      targetNode.x != null &&
      targetNode.y != null
    ) {
      drawPreviewLink(ctx, sourceNode.x, sourceNode.y, targetNode.x, targetNode.y, 0.6);
    }
  } else if (linkPreviewMousePos) {
    // No target node yet — draw line from dragged node to current mouse world position
    const sourceNode = nodeMap.get(linkPreviewMousePos.sourceId);
    if (sourceNode && sourceNode.x != null && sourceNode.y != null) {
      drawPreviewLink(
        ctx,
        sourceNode.x,
        sourceNode.y,
        linkPreviewMousePos.x,
        linkPreviewMousePos.y,
        0.35
      );
    }
  }

  // Draw nodes with CSS shadows only when node count is below the threshold (performance).
  // Threshold: frontend.graph.2d.shadows_threshold in knowledge-graph.config.json
  const enableShadows = !focusMode && nodes.length < graphConfig2D.shadows_threshold;
  const simplified = !focusMode && transform.k < graphConfig2D.lod_simplify_zoom;
  const searchMatchIdSet = searchMatchIds ? new Set(searchMatchIds) : undefined;
  drawAllNodes(
    ctx,
    nodes,
    angles,
    enableShadows,
    nodeOpacity,
    disableVariation,
    animationTime,
    hoveredNodeId,
    particleSystem,
    focusMode,
    searchMatchIdSet,
    visibleNodeIds,
    simplified
  );

  ctx.restore();

  // Draw black hole in SCREEN coordinates so it stays fixed regardless of pan/zoom
  if (!focusMode && blackHole) {
    drawBlackHole(ctx, blackHole, animationTime);
    if (blackHole.hovered) {
      drawBlackHoleTooltip(ctx, blackHole);
    }
  }

  // Draw ghost node in SCREEN coordinates so it stays fixed regardless of pan/zoom
  if (!focusMode && ghostNode?.active) {
    drawGhostNodeScreen(ctx, ghostNode, animationTime);
    if (ghostNode.hovered) {
      drawGhostNodeTooltipScreen(ctx, ghostNode);
    }
  }
}

/**
 * Reset view to center the graph
 */
export function resetView(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  nodes: SimulationNode[],
  transform: { x: number; y: number; k: number }
): void {
  if (nodes.length === 0) return;

  // Find graph bounds
  let minX = Infinity,
    minY = Infinity,
    maxX = -Infinity,
    maxY = -Infinity;
  for (const node of nodes) {
    if (node.x! < minX) minX = node.x!;
    if (node.x! > maxX) maxX = node.x!;
    if (node.y! < minY) minY = node.y!;
    if (node.y! > maxY) maxY = node.y!;
  }

  // Add padding
  const padding = 50;
  minX -= padding;
  minY -= padding;
  maxX += padding;
  maxY += padding;

  const graphWidth = maxX - minX;
  const graphHeight = maxY - minY;

  // Compute scale to fit entire graph
  const scaleX = width / graphWidth;
  const scaleY = height / graphHeight;
  transform.k = Math.min(scaleX, scaleY, 1); // Don’t zoom beyond 1:1

  // Center
  transform.x = (width - graphWidth * transform.k) / 2 - minX * transform.k;
  transform.y = (height - graphHeight * transform.k) / 2 - minY * transform.k;
}
