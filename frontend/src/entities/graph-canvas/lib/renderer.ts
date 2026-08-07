/**
 * Canvas rendering functions for GraphCanvas
 */
import { graphConfig2D } from "$shared/config";
import {
  getLinkEndpointId,
  resolveLinkEndpoint,
  type SimulationNode,
  type SimulationLink,
} from "./types";
import { getVariation, applyHueShift } from "$shared/utils/variation";
import { BASE_NODE_RADIUS } from "./graph-constants";
import { drawBlackHole, drawBlackHoleTooltip } from "./black-hole";
import { drawGhostNodeScreen, drawGhostNodeTooltipScreen } from "./ghost-node";
import { drawDistortedBackgroundGrid } from "./gravity-system";
import type { BlackHoleState } from "./black-hole";
import type { GhostNodeState } from "./ghost-node";
import { getGlowIntensity } from "$shared/lib/graph/glow-intensity";
import { getAnomalyParams } from "$shared/lib/graph/renderer/anomalies/helpers";
import type { AnomalyRenderer } from "$shared/lib/graph/renderer/anomalies/helpers";

import { drawRealityRift } from "$shared/lib/graph/renderer/anomalies/reality-rift";
import { drawChromaticMaw } from "$shared/lib/graph/renderer/anomalies/chromatic-maw";
import { drawVoidWhisper } from "$shared/lib/graph/renderer/anomalies/void-whisper";
import { drawCosmicAbomination } from "$shared/lib/graph/renderer/anomalies/cosmic-abomination";
export { drawChromaticMaw, drawVoidWhisper, drawCosmicAbomination };
import { getNodeGradient } from "$shared/lib/graph/node-gradient";
import { CelestialBody, LinkType } from "$entities";

export type { SimulationNode, SimulationLink };
export type { BlackHoleState } from "./black-hole";
export type { GhostNodeState } from "./ghost-node";

// Performance thresholds — sourced from knowledge-graph.config.json → frontend.graph.2d
const PERFORMANCE_THRESHOLD_LINKS = graphConfig2D.animated_links_threshold;

/**
 * Simple hash function for strings (local copy for anomaly generation)
 */
function stringHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return Math.abs(hash);
}

/**
 * Deterministic pseudo-random float in [0,1) seeded by string + index.
 * Replaces Math.random() in per-frame drawing to eliminate flickering.
 */
function seededRand(seed: string, index: number): number {
  const h = stringHash(seed + ":" + index);
  // Use two large primes to spread bits; result in [0,1)
  return ((h * 1664525 + 1013904223) & 0x7fffffff) / 0x7fffffff;
}

/**
 * Draw a star node with glow, gradient and corona
 */
export function drawStar(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  variation?: { sizeMultiplier: number; hueShift: number; phaseShift?: number },
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const points = 5;
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const outerRadius = r * sizeMultiplier;
  const innerRadius = r * 0.4 * sizeMultiplier;
  const nodePhase = variation?.phaseShift ?? 0;
  let rot = angle + nodePhase;
  const step = Math.PI / points;

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined && nodeCount < (graphConfig2D.shadows_threshold ?? 100)) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 20 * glowIntensity;
    ctx.shadowColor = "#ffcc00";
  }

  // Draw corona (rays)
  if (time && nodeCount !== undefined && nodeCount <= (graphConfig2D.visual_fx_threshold ?? 500)) {
    const rayCount = 8;
    const rayLength = outerRadius * 0.5;
    const localTime = time + nodePhase * 1000;
    for (let i = 0; i < rayCount; i++) {
      const rayAngle = (i / rayCount) * Math.PI * 2 + localTime / 1000;
      const startX = x + Math.cos(rayAngle) * outerRadius;
      const startY = y + Math.sin(rayAngle) * outerRadius;
      const endX = x + Math.cos(rayAngle) * (outerRadius + rayLength);
      const endY = y + Math.sin(rayAngle) * (outerRadius + rayLength);

      ctx.beginPath();
      ctx.moveTo(startX, startY);
      ctx.lineTo(endX, endY);
      ctx.strokeStyle = `rgba(255, 204, 0, ${0.3 + 0.2 * Math.sin(localTime / 500 + i)})`;
      ctx.lineWidth = 2;
      ctx.stroke();
    }
  }

  ctx.beginPath();
  for (let i = 0; i < points; i++) {
    const x1 = x + Math.cos(rot) * outerRadius;
    const y1 = y + Math.sin(rot) * outerRadius;
    ctx.lineTo(x1, y1);
    rot += step;
    const x2 = x + Math.cos(rot) * innerRadius;
    const y2 = y + Math.sin(rot) * innerRadius;
    ctx.lineTo(x2, y2);
    rot += step;
  }
  ctx.closePath();

  // Use gradient instead of solid fill
  const gradient = getNodeGradient(ctx, x, y, outerRadius, "star", "#ffcc00");
  const hueShift = variation?.hueShift ?? 0;
  ctx.fillStyle = gradient;
  ctx.strokeStyle = applyHueShift("#cc9900", hueShift);
  ctx.lineWidth = 2;
  ctx.fill();
  ctx.stroke();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
}

/**
 * Draw a planet node with glow, gradient and rings
 */
export function drawPlanet(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  color?: string,
  variation?: { sizeMultiplier: number; hueShift: number },
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const adjustedR = r * sizeMultiplier;
  const hueShift = variation?.hueShift ?? 0;
  const planetColor = color ? applyHueShift(color, hueShift) : applyHueShift("#d6aa5d", hueShift);

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined && nodeCount < (graphConfig2D.shadows_threshold ?? 100)) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 15 * glowIntensity;
    ctx.shadowColor = planetColor;
  }

  ctx.beginPath();
  ctx.arc(x, y, adjustedR, 0, 2 * Math.PI);

  // Use gradient instead of solid fill
  const gradient = getNodeGradient(ctx, x, y, adjustedR, "planet", planetColor);
  ctx.fillStyle = gradient;
  ctx.fill();

  // Draw bands
  for (let i = -adjustedR / 2; i <= adjustedR / 2; i += adjustedR / 4) {
    ctx.beginPath();
    ctx.ellipse(x, y + i, adjustedR * 0.8, adjustedR * 0.15, angle, 0, 2 * Math.PI);
    ctx.fillStyle = color ? "rgba(100,100,100,0.3)" : "#b07a3a";
    ctx.fill();
  }

  // Draw rings (Saturn-like)
  if (nodeCount !== undefined && nodeCount <= (graphConfig2D.visual_fx_threshold ?? 500)) {
    ctx.save();
    ctx.translate(x, y);
    ctx.rotate(angle + Math.PI / 6);

    // Outer ring
    ctx.beginPath();
    ctx.ellipse(0, 0, adjustedR * 1.6, adjustedR * 0.3, 0, 0, 2 * Math.PI);
    ctx.strokeStyle = "rgba(200, 180, 150, 0.4)";
    ctx.lineWidth = 2;
    ctx.stroke();

    // Inner ring
    ctx.beginPath();
    ctx.ellipse(0, 0, adjustedR * 1.3, adjustedR * 0.2, 0, 0, 2 * Math.PI);
    ctx.strokeStyle = "rgba(180, 160, 130, 0.3)";
    ctx.lineWidth = 1.5;
    ctx.stroke();

    ctx.restore();
  }

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
}

/**
 * Draw a comet node
 */
/**
 * Draw a comet node with glow and gradient
 */
export function drawComet(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  variation?: { sizeMultiplier: number; hueShift: number; phaseShift?: number },
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const adjustedR = r * sizeMultiplier;
  const hueShift = variation?.hueShift ?? 0;
  const cometColor = applyHueShift("#e879f9", hueShift);
  const nodePhase = variation?.phaseShift ?? 0;

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined && nodeCount < (graphConfig2D.shadows_threshold ?? 100)) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 12 * glowIntensity;
    ctx.shadowColor = cometColor;
  }

  ctx.beginPath();
  ctx.arc(x, y, adjustedR, 0, 2 * Math.PI);
  ctx.fillStyle = cometColor;
  ctx.fill();

  // Longer tail (up to 60px)
  const tailLength = 60 * sizeMultiplier;
  const tailAngle = angle;
  const tipX = x + Math.cos(tailAngle) * tailLength;
  const tipY = y + Math.sin(tailAngle) * tailLength;

  // Curved tail using quadratic curve with a per-node phase offset so
  // comet tails don't all wag in perfect unison.
  ctx.beginPath();
  ctx.moveTo(x, y);
  const midX = x + Math.cos(tailAngle) * (tailLength * 0.5);
  const midY = y + Math.sin(tailAngle) * (tailLength * 0.5);
  const localTime = (time ?? 0) + nodePhase * 1000;
  const curveOffset = 15 * Math.sin(localTime / 500);
  ctx.quadraticCurveTo(midX + curveOffset, midY + curveOffset, tipX, tipY);
  ctx.lineWidth = 4 * sizeMultiplier;
  ctx.strokeStyle = `rgba(${applyHueShiftToRGBA(232, 121, 249, hueShift)}, 0.6)`;
  ctx.stroke();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
}

/**
 * Helper function to apply hue shift to RGBA values
 */
function applyHueShiftToRGBA(r: number, g: number, b: number, hueShift: number): string {
  const hex = `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)}`;
  const shifted = applyHueShift(hex, hueShift);
  const r2 = parseInt(shifted.slice(1, 3), 16);
  const g2 = parseInt(shifted.slice(3, 5), 16);
  const b2 = parseInt(shifted.slice(5, 7), 16);
  return `${r2}, ${g2}, ${b2}`;
}

/**
 * Convert hex color to rgba string
 */

/**
 * Draw a galaxy node with glow, gradient and spiral arms
 */
export function drawGalaxy(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  variation?: { sizeMultiplier: number; hueShift: number },
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const hueShift = variation?.hueShift ?? 0;
  const adjustedR = r * sizeMultiplier;

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined && nodeCount < (graphConfig2D.shadows_threshold ?? 100)) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 25 * glowIntensity;
    ctx.shadowColor = "#8b5cf6";
  }

  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(angle);

  // Spiral arms with gradient
  for (let arm = 0; arm < 4; arm++) {
    const armAngle = (arm * Math.PI) / 2;
    ctx.beginPath();
    ctx.moveTo(0, 0);

    for (let i = 0; i < 20; i++) {
      const t = i / 20;
      const spiralAngle = armAngle + t * Math.PI * 2;
      const radius = t * adjustedR;
      const px = Math.cos(spiralAngle) * radius;
      const py = Math.sin(spiralAngle) * radius;
      ctx.lineTo(px, py);
    }

    const baseColor = applyHueShiftToRGBA(139, 92, 246, hueShift);
    ctx.strokeStyle = `rgba(${baseColor}, ${0.6 - arm * 0.1})`;
    ctx.lineWidth = 3;
    ctx.stroke();
  }

  // Center gradient
  const gradient = getNodeGradient(ctx, x, y, adjustedR, "galaxy", "#8b5cf6");
  ctx.beginPath();
  ctx.arc(0, 0, adjustedR * 0.3, 0, 2 * Math.PI);
  ctx.fillStyle = gradient;
  ctx.fill();

  ctx.restore();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
}

/**
 * Draw a nebula node
 */
export function drawNebula(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number
): void {
  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(angle);
  // Nebula - more blurred and cyan
  for (let i = 0; i < 4; i++) {
    ctx.beginPath();
    ctx.ellipse(0, 0, r * (1.2 - i * 0.2), r * 0.5, i * 0.3, 0, 2 * Math.PI);
    ctx.fillStyle = `rgba(45, 212, 191, ${0.25 - i * 0.05})`;
    ctx.fill();
  }
  ctx.restore();
}

/**
 * Draw an asteroid node with glow and craters
 */
export function drawAsteroid(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number,
  variation?: { sizeMultiplier: number; hueShift: number },
  disableVariation: boolean = false,
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const hueShift = variation?.hueShift ?? 0;
  const adjustedR = r * sizeMultiplier;
  const asteroidColor = applyHueShift("#94a3b8", hueShift);
  // Stable seed — fallback to empty string so seededRand still works without nodeId
  const seed = nodeId ?? "asteroid";

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined && nodeCount < (graphConfig2D.shadows_threshold ?? 100)) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 10 * glowIntensity;
    ctx.shadowColor = asteroidColor;
  }

  // Irregular rocky shape — deterministic per node, no flickering
  ctx.beginPath();
  const points = 7;
  for (let i = 0; i < points; i++) {
    const theta = (i / points) * 2 * Math.PI;
    const radiusVariation = disableVariation ? 0.85 : 0.7 + seededRand(seed, i) * 0.3;
    const px = x + Math.cos(theta) * adjustedR * radiusVariation;
    const py = y + Math.sin(theta) * adjustedR * radiusVariation;
    if (i === 0) ctx.moveTo(px, py);
    else ctx.lineTo(px, py);
  }
  ctx.closePath();

  ctx.fillStyle = asteroidColor;
  ctx.fill();
  ctx.strokeStyle = applyHueShift("#64748b", hueShift);
  ctx.lineWidth = 1;
  ctx.stroke();

  // Add craters (dark spots) — deterministic count and positions
  const craterCount = 3 + Math.floor(seededRand(seed, 100) * 3);
  for (let i = 0; i < craterCount; i++) {
    const craterAngle = seededRand(seed, 200 + i) * Math.PI * 2;
    const craterDist = seededRand(seed, 300 + i) * adjustedR * 0.6;
    const craterX = x + Math.cos(craterAngle) * craterDist;
    const craterY = y + Math.sin(craterAngle) * craterDist;
    const craterR = adjustedR * 0.15 * (0.5 + seededRand(seed, 400 + i) * 0.5);

    ctx.beginPath();
    ctx.arc(craterX, craterY, craterR, 0, 2 * Math.PI);
    ctx.fillStyle = "rgba(0, 0, 0, 0.2)";
    ctx.fill();
  }

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
}

/**
 * Draw a debris node
 */
export function drawDebris(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number,
  disableVariation: boolean = false,
  nodeId?: string
): void {
  const seed = nodeId ?? "debris";
  // Scattered small particles — deterministic positions per node
  ctx.fillStyle = "rgba(150, 150, 150, 0.6)";
  for (let i = 0; i < 5; i++) {
    const offsetX = disableVariation ? (i - 2) * (r * 0.25) : (seededRand(seed, i) - 0.5) * r * 2;
    const offsetY = disableVariation
      ? (i % 2 === 0 ? -1 : 1) * r * 0.2
      : (seededRand(seed, 10 + i) - 0.5) * r * 2;
    ctx.beginPath();
    ctx.arc(x + offsetX, y + offsetY, r * 0.3, 0, 2 * Math.PI);
    ctx.fill();
  }
}

/**
 * Draw a dust node as a soft, diffuse particle cloud.
 */
export function drawDust(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number,
  disableVariation: boolean = false,
  nodeId?: string
): void {
  const seed = nodeId ?? "dust";
  ctx.fillStyle = "rgba(160, 160, 160, 0.4)";
  for (let i = 0; i < 7; i++) {
    const offsetX = disableVariation ? (i - 3) * (r * 0.2) : (seededRand(seed, i) - 0.5) * r * 2.5;
    const offsetY = disableVariation
      ? (i % 2 === 0 ? -1 : 1) * (r * 0.15)
      : (seededRand(seed, 20 + i) - 0.5) * r * 2.5;
    const radius = disableVariation ? r * 0.2 : r * (0.25 + seededRand(seed, 40 + i) * 0.2);
    ctx.beginPath();
    ctx.arc(x + offsetX, y + offsetY, radius, 0, 2 * Math.PI);
    ctx.fill();
  }
}

/**
 * Draw a black hole node with glow
 */
export function drawBlackhole(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number,
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined && nodeCount < (graphConfig2D.shadows_threshold ?? 100)) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 30 * glowIntensity;
    ctx.shadowColor = "#ff6600";
  }

  // Event horizon (black circle)
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = "#000000";
  ctx.fill();

  // Accretion disk (glowing ring)
  ctx.beginPath();
  ctx.arc(x, y, r * 1.3, 0, 2 * Math.PI);
  ctx.strokeStyle = "#ff6600";
  ctx.lineWidth = 3;
  ctx.stroke();

  // Inner glow
  ctx.beginPath();
  ctx.arc(x, y, r * 1.1, 0, 2 * Math.PI);
  ctx.strokeStyle = "rgba(255, 102, 0, 0.5)";
  ctx.lineWidth = 2;
  ctx.stroke();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
}

/**
 * Draw background grid or nebula
 */
export function drawBackground(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  nodes: SimulationNode[],
  time: number
): void {
  // Draw subtle grid
  const gridSize = 100;
  ctx.strokeStyle = "rgba(255, 255, 255, 0.03)";
  ctx.lineWidth = 1;

  for (let x = 0; x <= width; x += gridSize) {
    ctx.beginPath();
    ctx.moveTo(x, 0);
    ctx.lineTo(x, height);
    ctx.stroke();
  }

  for (let y = 0; y <= height; y += gridSize) {
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(width, y);
    ctx.stroke();
  }

  // Draw nebula (large blurred ellipses)
  if (nodes.length <= (graphConfig2D.visual_fx_threshold ?? 500)) {
    const nebulaCount = 3;
    for (let i = 0; i < nebulaCount; i++) {
      const nebulaX = (width / nebulaCount) * (i + 0.5);
      const nebulaY = height / 2;
      const nebulaRadius = 200 + Math.sin(time / 5000 + i) * 50;

      const gradient = ctx.createRadialGradient(
        nebulaX,
        nebulaY,
        0,
        nebulaX,
        nebulaY,
        nebulaRadius
      );
      gradient.addColorStop(0, "rgba(139, 92, 246, 0.05)");
      gradient.addColorStop(0.5, "rgba(59, 130, 246, 0.03)");
      gradient.addColorStop(1, "rgba(0, 0, 0, 0)");

      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, width, height);
    }
  }
}

const BIDIRECTIONAL_LINK_OFFSET = 24;

/**
 * Draw a quadratic bezier path between two nodes, optionally curving it
 * perpendicular to the straight segment by `curveOffset`.
 */
function drawCurvedLinkPath(
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
function buildBidirectionalPairSet(links: SimulationLink[]): Set<string> {
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
 * Draw the Knowledge Core technical node
 */
export function drawTechnicalNode(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  animationTime?: number
): void {
  const pulse = animationTime ? 0.7 + 0.3 * Math.abs(Math.sin(animationTime / 800)) : 1;
  const radius = r * 1.2;

  ctx.save();
  ctx.globalAlpha = 0.85;

  // Soft purple glow
  ctx.shadowBlur = 20 * pulse;
  ctx.shadowColor = "rgba(138, 43, 226, 0.6)";

  // Semi-transparent sphere
  ctx.beginPath();
  ctx.arc(x, y, radius, 0, 2 * Math.PI);
  const gradient = ctx.createRadialGradient(
    x - radius * 0.3,
    y - radius * 0.3,
    radius * 0.1,
    x,
    y,
    radius
  );
  gradient.addColorStop(0, "rgba(167, 139, 250, 0.4)");
  gradient.addColorStop(1, "rgba(138, 43, 226, 0.15)");
  ctx.fillStyle = gradient;
  ctx.fill();

  // Border
  ctx.strokeStyle = `rgba(167, 139, 250, ${pulse})`;
  ctx.lineWidth = 2;
  ctx.stroke();

  // Question mark icon
  ctx.shadowBlur = 0;
  ctx.fillStyle = "rgba(255, 255, 255, 0.9)";
  ctx.font = `${Math.floor(radius * 1.2)}px sans-serif`;
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText("?", x, y + radius * 0.05);

  ctx.restore();
}

/**
 * Draw a moon node with glow
 */
export function drawMoon(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number
): void {
  // Moon body (grey circle)
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = "#cccccc";
  ctx.fill();
  ctx.strokeStyle = "#999999";
  ctx.lineWidth = 1;
  ctx.stroke();
  // Crater
  ctx.beginPath();
  ctx.arc(x - r * 0.3, y - r * 0.2, r * 0.25, 0, 2 * Math.PI);
  ctx.fillStyle = "#aaaaaa";
  ctx.fill();
}

/**
 * Seeded random number generator for deterministic anomaly parameters
 */

/**
 * Generate deterministic parameters for anomaly visualization
 */

/**
 * Draw a Reality Rift anomaly - dark core + jagged cracks + amoebic contour
 */

/**
 * Draw a Chromatic Maw anomaly - tentacles + gradient core
 */

/**
 * Draw a Void Whisper anomaly - particles + lines + snow effect
 */

/**
 * Draw a Cosmic Abomination - combines all three anomaly types
 */

/**
 * Draw an unknown type node - dispatcher for anomaly types
 */
export function drawUnknown(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  nodeId: string,
  customRenderers?: Record<number, AnomalyRenderer>
): void {
  // Select anomaly type based on hash of nodeId (deterministic)
  const hash = stringHash(nodeId);
  const anomalyType = hash % 4;

  const params = getAnomalyParams(nodeId);
  const renderers =
    customRenderers ??
    ({
      0: drawRealityRift,
      1: drawChromaticMaw,
      2: drawVoidWhisper,
      3: drawCosmicAbomination,
    } as Record<number, AnomalyRenderer>);

  const rendererFn = renderers[anomalyType] ?? drawRealityRift;
  rendererFn(ctx, x, y, r, params);
}

/**
 * Wire each CelestialBody instance to its Canvas renderer.
 * Kept in the renderer module so the domain object stays free of Canvas imports.
 */
function registerCelestialBodyDrawers(): void {
  CelestialBody.STAR.drawFunction = (ctx, c) => {
    drawStar(ctx, c.x, c.y, c.r, c.angle, c.variation, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.PLANET.drawFunction = (ctx, c) => {
    drawPlanet(ctx, c.x, c.y, c.r, c.angle, undefined, c.variation, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.SATELLITE.drawFunction = (ctx, c) => {
    drawPlanet(
      ctx,
      c.x,
      c.y,
      c.r,
      c.angle,
      CelestialBody.SATELLITE.color,
      c.variation,
      c.nodeId,
      c.nodeCount,
      c.time
    );
  };

  CelestialBody.COMET.drawFunction = (ctx, c) => {
    drawComet(ctx, c.x, c.y, c.r, c.angle, c.variation, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.GALAXY.drawFunction = (ctx, c) => {
    drawGalaxy(ctx, c.x, c.y, c.r, c.angle, c.variation, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.NEBULA.drawFunction = (ctx, c) => {
    drawNebula(ctx, c.x, c.y, c.r, c.angle);
  };

  CelestialBody.ASTEROID.drawFunction = (ctx, c) => {
    drawAsteroid(
      ctx,
      c.x,
      c.y,
      c.r,
      c.angle,
      c.variation,
      c.disableVariation || c.focusMode,
      c.nodeId,
      c.nodeCount,
      c.time
    );
  };

  CelestialBody.DEBRIS.drawFunction = (ctx, c) => {
    drawDebris(ctx, c.x, c.y, c.r, c.angle, c.disableVariation || c.focusMode, c.nodeId);
  };

  CelestialBody.DUST.drawFunction = (ctx, c) => {
    drawDust(ctx, c.x, c.y, c.r, c.angle, c.disableVariation || c.focusMode, c.nodeId);
  };

  CelestialBody.BLACKHOLE.drawFunction = (ctx, c) => {
    drawBlackhole(ctx, c.x, c.y, c.r, c.angle, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.TECHNICAL.drawFunction = (ctx, c) => {
    drawTechnicalNode(ctx, c.x, c.y, c.r, c.time);
  };

  CelestialBody.MOON.drawFunction = (ctx, c) => {
    drawMoon(ctx, c.x, c.y, c.r, c.angle);
  };

  CelestialBody.UNKNOWN.drawFunction = (ctx, c) => {
    drawUnknown(ctx, c.x, c.y, c.r, c.angle, c.nodeId);
  };

  CelestialBody.REALITY_RIFT.drawFunction = (ctx, c) => {
    drawRealityRift(ctx, c.x, c.y, c.r, getAnomalyParams(c.nodeId));
  };

  CelestialBody.CHROMATIC_MAW.drawFunction = (ctx, c) => {
    drawChromaticMaw(ctx, c.x, c.y, c.r, getAnomalyParams(c.nodeId));
  };

  CelestialBody.VOID_WHISPER.drawFunction = (ctx, c) => {
    drawVoidWhisper(ctx, c.x, c.y, c.r, getAnomalyParams(c.nodeId));
  };

  CelestialBody.COSMIC_ABOMINATION.drawFunction = (ctx, c) => {
    drawCosmicAbomination(ctx, c.x, c.y, c.r, getAnomalyParams(c.nodeId));
  };
}

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
  dyingLinkOpacity: Map<string, number> = new Map()
): void {
  let drawnCount = 0;
  let skippedCount = 0;

  if (import.meta.env.DEV) {
    console.log(`[drawAllLinks] Called with ${simLinks.length} links and ${nodes.length} nodes`);
  }

  const bidirectionalPairs = buildBidirectionalPairSet(simLinks);
  const nodeMap = new Map<string, SimulationNode>();
  for (const node of nodes) {
    if (node.id) {
      nodeMap.set(node.id, node);
    }
  }

  simLinks.forEach((link, index) => {
    const sourceNode = resolveLinkEndpoint(link.source, nodes);
    const targetNode = resolveLinkEndpoint(link.target, nodes);

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
    const sourceId = getLinkEndpointId(link.source);
    const targetId = getLinkEndpointId(link.target);
    const [a, b] = sourceId < targetId ? [sourceId, targetId] : [targetId, sourceId];
    const isBidirectional = bidirectionalPairs.has(`${a}|${b}`);
    const curveOffset = isBidirectional
      ? (sourceId < targetId ? 1 : -1) * BIDIRECTIONAL_LINK_OFFSET
      : 0;

    // Use animated link drawing
    drawAnimatedLink(
      ctx,
      link,
      nodeMap,
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
    const sourceNode = resolveLinkEndpoint(link.source, nodes);
    const targetNode = resolveLinkEndpoint(link.target, nodes);
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
  focusMode: boolean = false
): void {
  const body = CelestialBody.fromString(node.type);

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
  searchMatchIds?: string[]
): void {
  const r = BASE_NODE_RADIUS;
  const nodeCount = nodes.length;

  if (disableVariation) {
    for (const node of nodes) {
      if (!angles.has(node.id)) {
        angles.set(node.id, 0);
      }
    }
  }

  if (particleSystem?.isEnabled()) {
    for (const node of nodes) {
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
    const angle = angles.get(node.id) || 0;
    const opacity = nodeOpacity?.get(node.id) ?? 1;
    const isHovered = hoveredNodeId === node.id;
    const isSearchMatch = searchMatchIds?.includes(node.id) ?? false;
    const finalOpacity = hoveredNodeId ? (isHovered ? 1 : 0.3) : opacity;

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
      focusMode
    );
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
      // Per-node phase so multiple new notes don't pulse in lockstep.
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

    ctx.globalAlpha = previousAlpha;
  });
}

/**
 * Check if a node was created within the last 24 hours.
 */
function isNewNode(node: SimulationNode): boolean {
  if (!node.createdAt) return false;
  const created = new Date(node.createdAt).getTime();
  if (Number.isNaN(created)) return false;
  return Date.now() - created < 24 * 60 * 60 * 1000;
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
      maxDistance?: number
    ) => { dx: number; dy: number };
    isEnabled: (nodeCount: number) => boolean;
  } | null,
  focusMode: boolean = false,
  searchMatchIds?: string[],
  highlightedLinkId?: string | null,
  linkPreviewTarget?: { sourceId: string; targetId: string } | null,
  linkPreviewMousePos?: { sourceId: string; x: number; y: number } | null
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
    dyingLinkOpacity
  );

  // Draw link preview if dragging for link creation
  if (linkPreviewTarget) {
    // Preview to a specific target node
    const sourceNode = nodes.find((n) => n.id === linkPreviewTarget.sourceId);
    const targetNode = nodes.find((n) => n.id === linkPreviewTarget.targetId);
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
    const sourceNode = nodes.find((n) => n.id === linkPreviewMousePos.sourceId);
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
    searchMatchIds
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
  transform.k = Math.min(scaleX, scaleY, 1); // Don't zoom beyond 1:1

  // Center
  transform.x = (width - graphWidth * transform.k) / 2 - minX * transform.k;
  transform.y = (height - graphHeight * transform.k) / 2 - minY * transform.k;
}
