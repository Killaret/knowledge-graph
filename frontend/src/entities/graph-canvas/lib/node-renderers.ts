/**
 * Canvas node renderers for GraphCanvas
 */
import { graphConfig2D } from "$shared/config";
import { getGlowIntensity } from "$shared/lib/graph/glow-intensity";
import { getNodeGradient } from "$shared/lib/graph/node-gradient";
import { getAnomalyParams } from "$shared/lib/graph/renderer/anomalies/helpers";
import type { AnomalyRenderer } from "$shared/lib/graph/renderer/anomalies/helpers";
import { drawRealityRift } from "$shared/lib/graph/renderer/anomalies/reality-rift";
import { drawChromaticMaw } from "$shared/lib/graph/renderer/anomalies/chromatic-maw";
import { drawVoidWhisper } from "$shared/lib/graph/renderer/anomalies/void-whisper";
import { drawCosmicAbomination } from "$shared/lib/graph/renderer/anomalies/cosmic-abomination";
import type { NodeVariation } from "$shared/utils/variation";
import { stringHash, seededRand } from "./renderer-utils";
import { hexToRgba } from "$shared/lib/graph/helpers";

/**
 * Draw a star node with glow, gradient and corona
 */
export function drawStar(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  variation?: NodeVariation,
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const points = 5;
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const outerRadius = r * sizeMultiplier;
  // A slightly fuller 5-point star; 0.45 gives a classic, less spiky silhouette.
  const innerRadius = r * 0.45 * sizeMultiplier;
  const nodePhase = variation?.phaseShift ?? 0;
  let rot = angle + nodePhase;
  const step = Math.PI / points;

  // Body colors: explicit variation, manual color, or canonical gold.
  const fillColor = variation?.color ?? "#ffcc00";
  const glowColor = variation?.glowColor ?? "#ffcc00";
  const strokeColor = variation?.strokeColor ?? "#cc9900";

  // Apply glow effect
  if (
    time &&
    nodeId &&
    nodeCount !== undefined &&
    nodeCount < (graphConfig2D.shadows_threshold ?? 100)
  ) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 20 * glowIntensity;
    ctx.shadowColor = glowColor;
  }

  // Round joins and caps keep the star and its corona rays smooth and
  // star-like instead of sharply mitered or blunt-ended.
  ctx.lineJoin = "round";
  ctx.lineCap = "round";

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
      // Use the node's glow color to tint the corona, preserving the original alpha curve.
      ctx.strokeStyle = hexToRgba(glowColor, 0.3 + 0.2 * Math.sin(localTime / 500 + i));
      ctx.lineWidth = 2;
      ctx.stroke();
    }
  }

  ctx.beginPath();
  for (let i = 0; i < points; i++) {
    const outerX = x + Math.cos(rot) * outerRadius;
    const outerY = y + Math.sin(rot) * outerRadius;
    if (i === 0) {
      ctx.moveTo(outerX, outerY);
    } else {
      ctx.lineTo(outerX, outerY);
    }
    rot += step;
    const innerX = x + Math.cos(rot) * innerRadius;
    const innerY = y + Math.sin(rot) * innerRadius;
    ctx.lineTo(innerX, innerY);
    rot += step;
  }
  ctx.closePath();

  // Use gradient instead of solid fill
  const gradient = getNodeGradient(ctx, x, y, outerRadius, "star", fillColor);
  ctx.fillStyle = gradient;
  ctx.strokeStyle = strokeColor;
  ctx.lineWidth = 2;
  ctx.fill();
  ctx.stroke();

  // Reset shadow and line caps/joins so they don't leak to the next node.
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
  ctx.lineJoin = "miter";
  ctx.lineCap = "butt";
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
  variation?: NodeVariation,
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const adjustedR = r * sizeMultiplier;
  const fillColor = variation?.color ?? "#d6aa5d";
  const glowColor = variation?.glowColor ?? fillColor;

  // Apply glow effect
  if (
    time &&
    nodeId &&
    nodeCount !== undefined &&
    nodeCount < (graphConfig2D.shadows_threshold ?? 100)
  ) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 15 * glowIntensity;
    ctx.shadowColor = glowColor;
  }

  ctx.beginPath();
  ctx.arc(x, y, adjustedR, 0, 2 * Math.PI);

  // Use gradient instead of solid fill
  const gradient = getNodeGradient(ctx, x, y, adjustedR, "planet", fillColor);
  ctx.fillStyle = gradient;
  ctx.fill();

  // Draw bands
  for (let i = -adjustedR / 2; i <= adjustedR / 2; i += adjustedR / 4) {
    ctx.beginPath();
    ctx.ellipse(x, y + i, adjustedR * 0.8, adjustedR * 0.15, angle, 0, 2 * Math.PI);
    // Bands use a muted, semi-transparent complement of the planet color.
    ctx.fillStyle = hexToRgba(glowColor, 0.25);
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
    ctx.strokeStyle = hexToRgba(glowColor, 0.4);
    ctx.lineWidth = 2;
    ctx.stroke();

    // Inner ring
    ctx.beginPath();
    ctx.ellipse(0, 0, adjustedR * 1.3, adjustedR * 0.2, 0, 0, 2 * Math.PI);
    ctx.strokeStyle = hexToRgba(glowColor, 0.3);
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
  variation?: NodeVariation,
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const adjustedR = r * sizeMultiplier;
  const cometColor = variation?.color ?? "#e879f9";
  const glowColor = variation?.glowColor ?? cometColor;
  const nodePhase = variation?.phaseShift ?? 0;

  // Apply glow effect
  if (
    time &&
    nodeId &&
    nodeCount !== undefined &&
    nodeCount < (graphConfig2D.shadows_threshold ?? 100)
  ) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 12 * glowIntensity;
    ctx.shadowColor = glowColor;
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
  // comet tails don’t all wag in perfect unison.
  ctx.beginPath();
  ctx.moveTo(x, y);
  const midX = x + Math.cos(tailAngle) * (tailLength * 0.5);
  const midY = y + Math.sin(tailAngle) * (tailLength * 0.5);
  const localTime = (time ?? 0) + nodePhase * 1000;
  const curveOffset = 15 * Math.sin(localTime / 500);
  ctx.quadraticCurveTo(midX + curveOffset, midY + curveOffset, tipX, tipY);
  ctx.lineWidth = 4 * sizeMultiplier;
  ctx.strokeStyle = hexToRgba(glowColor, 0.6);
  ctx.stroke();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = "transparent";
}

/**
 * Draw a galaxy node with glow, gradient and spiral arms
 */
export function drawGalaxy(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  variation?: NodeVariation,
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const adjustedR = r * sizeMultiplier;
  const fillColor = variation?.color ?? "#8b5cf6";
  const glowColor = variation?.glowColor ?? fillColor;

  // Apply glow effect
  if (
    time &&
    nodeId &&
    nodeCount !== undefined &&
    nodeCount < (graphConfig2D.shadows_threshold ?? 100)
  ) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 25 * glowIntensity;
    ctx.shadowColor = glowColor;
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

    ctx.strokeStyle = hexToRgba(fillColor, 0.6 - arm * 0.1);
    ctx.lineWidth = 3;
    ctx.stroke();
  }

  // Center gradient
  const gradient = getNodeGradient(ctx, x, y, adjustedR, "galaxy", fillColor);
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
  angle: number,
  variation?: NodeVariation
): void {
  const fillColor = variation?.color ?? "#2dd4bf";

  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(angle);
  // Nebula - diffuse ellipses with decreasing opacity.
  for (let i = 0; i < 4; i++) {
    ctx.beginPath();
    ctx.ellipse(0, 0, r * (1.2 - i * 0.2), r * 0.5, i * 0.3, 0, 2 * Math.PI);
    ctx.fillStyle = hexToRgba(fillColor, 0.25 - i * 0.05);
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
  variation?: NodeVariation,
  disableVariation: boolean = false,
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const adjustedR = r * sizeMultiplier;
  const asteroidColor = variation?.color ?? "#94a3b8";
  const glowColor = variation?.glowColor ?? asteroidColor;
  const strokeColor = variation?.strokeColor ?? "#64748b";
  // Stable seed — fallback to empty string so seededRand still works without nodeId
  const seed = nodeId ?? "asteroid";

  // Apply glow effect
  if (
    time &&
    nodeId &&
    nodeCount !== undefined &&
    nodeCount < (graphConfig2D.shadows_threshold ?? 100)
  ) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 10 * glowIntensity;
    ctx.shadowColor = glowColor;
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
  ctx.strokeStyle = strokeColor;
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
  nodeId?: string,
  variation?: NodeVariation
): void {
  const seed = nodeId ?? "debris";
  const debrisColor = variation?.color ?? "#969696";
  // Scattered small particles — deterministic positions per node
  ctx.fillStyle = hexToRgba(debrisColor, 0.6);
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
  nodeId?: string,
  variation?: NodeVariation
): void {
  const seed = nodeId ?? "dust";
  const dustColor = variation?.color ?? "#a0a0a0";
  ctx.fillStyle = hexToRgba(dustColor, 0.4);
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
  time?: number,
  variation?: NodeVariation
): void {
  const fillColor = variation?.color ?? "#000000";
  const glowColor = variation?.glowColor ?? "#ff6600";

  // Apply glow effect
  if (
    time &&
    nodeId &&
    nodeCount !== undefined &&
    nodeCount < (graphConfig2D.shadows_threshold ?? 100)
  ) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 30 * glowIntensity;
    ctx.shadowColor = glowColor;
  }

  // Event horizon (black circle)
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = fillColor;
  ctx.fill();

  // Accretion disk (glowing ring)
  ctx.beginPath();
  ctx.arc(x, y, r * 1.3, 0, 2 * Math.PI);
  ctx.strokeStyle = glowColor;
  ctx.lineWidth = 3;
  ctx.stroke();

  // Inner glow
  ctx.beginPath();
  ctx.arc(x, y, r * 1.1, 0, 2 * Math.PI);
  ctx.strokeStyle = hexToRgba(glowColor, 0.5);
  ctx.lineWidth = 2;
  ctx.stroke();

  // Reset shadow
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
  animationTime?: number,
  variation?: NodeVariation
): void {
  const pulse = animationTime ? 0.7 + 0.3 * Math.abs(Math.sin(animationTime / 800)) : 1;
  const radius = r * 1.2;
  const fillColor = variation?.color ?? "#8b5cf6";
  const glowColor = variation?.glowColor ?? fillColor;

  ctx.save();
  ctx.globalAlpha = 0.85;

  // Soft colored glow
  ctx.shadowBlur = 20 * pulse;
  ctx.shadowColor = glowColor;

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
  gradient.addColorStop(0, hexToRgba(glowColor, 0.4));
  gradient.addColorStop(1, hexToRgba(fillColor, 0.15));
  ctx.fillStyle = gradient;
  ctx.fill();

  // Border
  ctx.strokeStyle = hexToRgba(glowColor, pulse);
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
  _angle: number,
  variation?: NodeVariation
): void {
  const fillColor = variation?.color ?? "#cccccc";
  const strokeColor = variation?.strokeColor ?? "#999999";
  const craterColor = variation?.glowColor ?? "#aaaaaa";

  // Moon body (grey circle)
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = fillColor;
  ctx.fill();
  ctx.strokeStyle = strokeColor;
  ctx.lineWidth = 1;
  ctx.stroke();
  // Crater
  ctx.beginPath();
  ctx.arc(x - r * 0.3, y - r * 0.2, r * 0.25, 0, 2 * Math.PI);
  ctx.fillStyle = craterColor;
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
