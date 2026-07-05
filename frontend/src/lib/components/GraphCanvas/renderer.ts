/**
 * Canvas rendering functions for GraphCanvas
 */
import { graphConfig2D, anomalyConfig } from '$lib/config';
import type { SimulationNode, SimulationLink } from './types';
import { getVariation, applyHueShift } from '$lib/utils/variation';
import {
  drawBlackHole,
  drawBlackHoleTooltip,
  type BlackHoleState
} from './black-hole';
import {
  drawGhostNode,
  drawGhostNodeTooltip,
  type GhostNodeState
} from './ghost-node';
import { drawDistortedBackgroundGrid, createGravitySystem } from './gravity-system';

export type { SimulationNode, SimulationLink };

// Performance thresholds
const PERFORMANCE_THRESHOLD_NODES = 100;
const PERFORMANCE_THRESHOLD_LINKS = 50;

/**
 * Get glow intensity based on time and node ID (pulsating effect)
 */
export function getGlowIntensity(nodeId: string, time: number, nodeCount: number): number {
  if (nodeCount > PERFORMANCE_THRESHOLD_NODES) {
    return 0.3; // Minimal glow for stars only
  }

  const hash = stringHash(nodeId);
  const phase = (hash % 1000) / 1000;
  const period = 2000 + (hash % 1000);
  const t = (time + phase * period) % period;
  const normalizedT = t / period;

  // Sine wave from 0.3 to 1.0 (absolute value to ensure positive)
  return 0.3 + 0.7 * Math.abs(Math.sin(normalizedT * Math.PI * 2));
}

/**
 * Get radial gradient for a node
 */
export function getNodeGradient(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  type: string,
  color: string
): CanvasGradient {
  const gradient = ctx.createRadialGradient(x - r * 0.3, y - r * 0.3, 0, x, y, r);

  switch (type) {
    case 'star':
      gradient.addColorStop(0, '#ffffcc');
      gradient.addColorStop(0.5, '#ffcc00');
      gradient.addColorStop(1, '#ff9900');
      break;
    case 'planet':
      gradient.addColorStop(0, lightenColor(color, 30));
      gradient.addColorStop(0.7, color);
      gradient.addColorStop(1, darkenColor(color, 20));
      break;
    case 'galaxy':
      gradient.addColorStop(0, '#8b5cf6');
      gradient.addColorStop(0.5, '#6d28d9');
      gradient.addColorStop(1, 'rgba(109, 40, 217, 0)');
      break;
    default:
      gradient.addColorStop(0, lightenColor(color, 20));
      gradient.addColorStop(1, color);
  }

  return gradient;
}

function lightenColor(hex: string, percent: number): string {
  const num = parseInt(hex.replace('#', ''), 16);
  const amt = Math.round(2.55 * percent);
  const R = Math.min(255, (num >> 16) + amt);
  const G = Math.min(255, ((num >> 8) & 0x00ff) + amt);
  const B = Math.min(255, (num & 0x0000ff) + amt);
  return `#${(1 << 24 | R << 16 | G << 8 | B).toString(16).slice(1)}`;
}

function darkenColor(hex: string, percent: number): string {
  const num = parseInt(hex.replace('#', ''), 16);
  const amt = Math.round(2.55 * percent);
  const R = Math.max(0, (num >> 16) - amt);
  const G = Math.max(0, ((num >> 8) & 0x00ff) - amt);
  const B = Math.max(0, (num & 0x0000ff) - amt);
  return `#${(1 << 24 | R << 16 | G << 8 | B).toString(16).slice(1)}`;
}

/**
 * Simple hash function for strings (local copy for anomaly generation)
 */
function stringHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return Math.abs(hash);
}

// Colors for different link types (per specification)
const linkTypeColors: Record<string, string> = {
  reference: '#3366ff', // Blue - default direct link
  dependency: '#ff6600', // Orange - dependency
  related: '#999999', // Gray - related topic (default)
  custom: '#ff66ff' // Pink - custom
};

/**
 * Get link color based on weight and type
 */
export function getLinkColor(weight: number, linkType?: string, fadeOpacity: number = 1): string {
  const effectiveType = linkType || 'related';
  const color = linkTypeColors[effectiveType] || linkTypeColors['related'];
  const baseOpacity = 0.4 + (weight ?? 0.5) * 0.4;
  const finalOpacity = baseOpacity * fadeOpacity;

  // Convert hex to rgba
  const r = parseInt(color.slice(1, 3), 16);
  const g = parseInt(color.slice(3, 5), 16);
  const b = parseInt(color.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${finalOpacity})`;
}

/**
 * Get line dash pattern based on link type and weight
 */
export function getLineDash(linkType?: string, weight?: number): number[] {
  const effectiveType = linkType || 'related';

  switch (effectiveType) {
    case 'reference':
      return []; // Solid
    case 'dependency':
      return [10, 3]; // Dash-dot
    case 'related':
      // Dash only for weak weight
      return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
    case 'custom':
      return [2, 6]; // Dotted
    default:
      return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
  }
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
  variation?: { sizeMultiplier: number; hueShift: number },
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const points = 5;
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const outerRadius = r * sizeMultiplier;
  const innerRadius = r * 0.4 * sizeMultiplier;
  let rot = angle;
  const step = Math.PI / points;

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 20 * glowIntensity;
    ctx.shadowColor = '#ffcc00';
  }

  // Draw corona (rays)
  if (time && nodeCount !== undefined && nodeCount <= 50) {
    const rayCount = 8;
    const rayLength = outerRadius * 0.5;
    for (let i = 0; i < rayCount; i++) {
      const rayAngle = (i / rayCount) * Math.PI * 2 + (time ? time / 1000 : 0);
      const startX = x + Math.cos(rayAngle) * outerRadius;
      const startY = y + Math.sin(rayAngle) * outerRadius;
      const endX = x + Math.cos(rayAngle) * (outerRadius + rayLength);
      const endY = y + Math.sin(rayAngle) * (outerRadius + rayLength);
      
      ctx.beginPath();
      ctx.moveTo(startX, startY);
      ctx.lineTo(endX, endY);
      ctx.strokeStyle = `rgba(255, 204, 0, ${0.3 + 0.2 * Math.sin(time / 500 + i)})`;
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
  const gradient = getNodeGradient(ctx, x, y, outerRadius, 'star', '#ffcc00');
  const hueShift = variation?.hueShift ?? 0;
  ctx.fillStyle = gradient;
  ctx.strokeStyle = applyHueShift('#cc9900', hueShift);
  ctx.lineWidth = 2;
  ctx.fill();
  ctx.stroke();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = 'transparent';
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
  const planetColor = color ? applyHueShift(color, hueShift) : applyHueShift('#d6aa5d', hueShift);

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 15 * glowIntensity;
    ctx.shadowColor = planetColor;
  }

  ctx.beginPath();
  ctx.arc(x, y, adjustedR, 0, 2 * Math.PI);
  
  // Use gradient instead of solid fill
  const gradient = getNodeGradient(ctx, x, y, adjustedR, 'planet', planetColor);
  ctx.fillStyle = gradient;
  ctx.fill();

  // Draw bands
  for (let i = -adjustedR / 2; i <= adjustedR / 2; i += adjustedR / 4) {
    ctx.beginPath();
    ctx.ellipse(x, y + i, adjustedR * 0.8, adjustedR * 0.15, angle, 0, 2 * Math.PI);
    ctx.fillStyle = color ? 'rgba(100,100,100,0.3)' : '#b07a3a';
    ctx.fill();
  }

  // Draw rings (Saturn-like)
  if (nodeCount !== undefined && nodeCount <= 50) {
    ctx.save();
    ctx.translate(x, y);
    ctx.rotate(angle + Math.PI / 6);
    
    // Outer ring
    ctx.beginPath();
    ctx.ellipse(0, 0, adjustedR * 1.6, adjustedR * 0.3, 0, 0, 2 * Math.PI);
    ctx.strokeStyle = 'rgba(200, 180, 150, 0.4)';
    ctx.lineWidth = 2;
    ctx.stroke();
    
    // Inner ring
    ctx.beginPath();
    ctx.ellipse(0, 0, adjustedR * 1.3, adjustedR * 0.2, 0, 0, 2 * Math.PI);
    ctx.strokeStyle = 'rgba(180, 160, 130, 0.3)';
    ctx.lineWidth = 1.5;
    ctx.stroke();
    
    ctx.restore();
  }

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = 'transparent';
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
  variation?: { sizeMultiplier: number; hueShift: number },
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const adjustedR = r * sizeMultiplier;
  const hueShift = variation?.hueShift ?? 0;
  const cometColor = applyHueShift('#e879f9', hueShift);

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined) {
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

  // Curved tail using quadratic curve
  ctx.beginPath();
  ctx.moveTo(x, y);
  const midX = x + Math.cos(tailAngle) * (tailLength * 0.5);
  const midY = y + Math.sin(tailAngle) * (tailLength * 0.5);
  const curveOffset = 15 * Math.sin(time ? time / 500 : 0);
  ctx.quadraticCurveTo(midX + curveOffset, midY + curveOffset, tipX, tipY);
  ctx.lineWidth = 4 * sizeMultiplier;
  ctx.strokeStyle = `rgba(${applyHueShiftToRGBA(232, 121, 249, hueShift)}, 0.6)`;
  ctx.stroke();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = 'transparent';
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
function hexToRgba(hex: string, alpha: number): string {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
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
  variation?: { sizeMultiplier: number; hueShift: number },
  nodeId?: string,
  nodeCount?: number,
  time?: number
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const hueShift = variation?.hueShift ?? 0;
  const adjustedR = r * sizeMultiplier;

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 25 * glowIntensity;
    ctx.shadowColor = '#8b5cf6';
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
  const gradient = getNodeGradient(ctx, x, y, adjustedR, 'galaxy', '#8b5cf6');
  ctx.beginPath();
  ctx.arc(0, 0, adjustedR * 0.3, 0, 2 * Math.PI);
  ctx.fillStyle = gradient;
  ctx.fill();

  ctx.restore();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = 'transparent';
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
  const asteroidColor = applyHueShift('#94a3b8', hueShift);

  // Apply glow effect
  if (time && nodeId && nodeCount !== undefined) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 10 * glowIntensity;
    ctx.shadowColor = asteroidColor;
  }

  // Irregular rocky shape
  ctx.beginPath();
  const points = 7;
  for (let i = 0; i < points; i++) {
    const theta = (i / points) * 2 * Math.PI;
    const radiusVariation = disableVariation ? 0.85 : 0.7 + Math.random() * 0.3;
    const px = x + Math.cos(theta) * adjustedR * radiusVariation;
    const py = y + Math.sin(theta) * adjustedR * radiusVariation;
    if (i === 0) ctx.moveTo(px, py);
    else ctx.lineTo(px, py);
  }
  ctx.closePath();

  ctx.fillStyle = asteroidColor;
  ctx.fill();
  ctx.strokeStyle = applyHueShift('#64748b', hueShift);
  ctx.lineWidth = 1;
  ctx.stroke();

  // Add craters (dark spots)
  const craterCount = 3 + Math.floor(Math.random() * 3);
  for (let i = 0; i < craterCount; i++) {
    const craterAngle = Math.random() * Math.PI * 2;
    const craterDist = Math.random() * adjustedR * 0.6;
    const craterX = x + Math.cos(craterAngle) * craterDist;
    const craterY = y + Math.sin(craterAngle) * craterDist;
    const craterR = adjustedR * 0.15 * (0.5 + Math.random() * 0.5);
    
    ctx.beginPath();
    ctx.arc(craterX, craterY, craterR, 0, 2 * Math.PI);
    ctx.fillStyle = 'rgba(0, 0, 0, 0.2)';
    ctx.fill();
  }

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = 'transparent';
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
  disableVariation: boolean = false
): void {
  // Scattered small particles
  ctx.fillStyle = 'rgba(150, 150, 150, 0.6)';
  for (let i = 0; i < 5; i++) {
    const offsetX = disableVariation ? (i - 2) * (r * 0.25) : (Math.random() - 0.5) * r * 2;
    const offsetY = disableVariation ? ((i % 2 === 0 ? -1 : 1) * r * 0.2) : (Math.random() - 0.5) * r * 2;
    ctx.beginPath();
    ctx.arc(x + offsetX, y + offsetY, r * 0.3, 0, 2 * Math.PI);
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
  if (time && nodeId && nodeCount !== undefined) {
    const glowIntensity = getGlowIntensity(nodeId, time, nodeCount);
    ctx.shadowBlur = 30 * glowIntensity;
    ctx.shadowColor = '#ff6600';
  }

  // Event horizon (black circle)
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = '#000000';
  ctx.fill();

  // Accretion disk (glowing ring)
  ctx.beginPath();
  ctx.arc(x, y, r * 1.3, 0, 2 * Math.PI);
  ctx.strokeStyle = '#ff6600';
  ctx.lineWidth = 3;
  ctx.stroke();

  // Inner glow
  ctx.beginPath();
  ctx.arc(x, y, r * 1.1, 0, 2 * Math.PI);
  ctx.strokeStyle = 'rgba(255, 102, 0, 0.5)';
  ctx.lineWidth = 2;
  ctx.stroke();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = 'transparent';
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
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.03)';
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
  if (nodes.length <= 50) {
    const nebulaCount = 3;
    for (let i = 0; i < nebulaCount; i++) {
      const nebulaX = (width / nebulaCount) * (i + 0.5);
      const nebulaY = height / 2;
      const nebulaRadius = 200 + Math.sin(time / 5000 + i) * 50;

      const gradient = ctx.createRadialGradient(nebulaX, nebulaY, 0, nebulaX, nebulaY, nebulaRadius);
      gradient.addColorStop(0, 'rgba(139, 92, 246, 0.05)');
      gradient.addColorStop(0.5, 'rgba(59, 130, 246, 0.03)');
      gradient.addColorStop(1, 'rgba(0, 0, 0, 0)');

      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, width, height);
    }
  }
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
  hoveredNodeId?: string | null
): void {
  const sourceId = typeof link.source === 'string' ? link.source : link.source.id;
  const targetId = typeof link.target === 'string' ? link.target : link.target.id;

  const source = nodes.get(sourceId);
  const target = nodes.get(targetId);
  if (!source || !target || source.x == null || source.y == null || target.x == null || target.y == null) {
    return;
  }

  if (linkCount > PERFORMANCE_THRESHOLD_LINKS) {
    // Fallback to static link for performance
    drawLink(ctx, link, source, target, 1, hoveredNodeId);
    return;
  }

  // Check if this link should be highlighted
  const isHighlighted = hoveredNodeId && (sourceId === hoveredNodeId || targetId === hoveredNodeId);
  const opacity = hoveredNodeId ? (isHighlighted ? 1 : 0.3) : 1;

  const color = getLinkColor(link.weight ?? 0.5, link.link_type, opacity);
  const dashArray = getLineDash(link.link_type, link.weight);

  // Draw base line
  ctx.beginPath();
  ctx.moveTo(source.x, source.y);
  ctx.lineTo(target.x, target.y);
  ctx.strokeStyle = color;
  ctx.lineWidth = 2;
  ctx.setLineDash(dashArray);
  ctx.stroke();

  // Animate dash offset for moving dots effect
  const speed = 0.5 + (link.weight ?? 0.5) * 0.5;
  const offset = (time * speed) % 20;
  ctx.lineDashOffset = -offset;
  ctx.stroke();

  // Reset line dash
  ctx.setLineDash([]);
  ctx.lineDashOffset = 0;
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
  isDuplicateHighlighted?: boolean
): void {
  // Check if this link should be highlighted
  const isHovered = hoveredNodeId && (link.source === hoveredNodeId || link.target === hoveredNodeId);
  const finalOpacity = hoveredNodeId ? (isHovered ? 1 : 0.3) : opacity;

  ctx.beginPath();
  ctx.moveTo(sourceNode.x!, sourceNode.y!);
  ctx.lineTo(targetNode.x!, targetNode.y!);

  const weight = link.weight ?? 0.5;
  const linkType = link.link_type;

  const lineWidth = Math.max(1, weight * 4) * (isDuplicateHighlighted ? 1.5 : 1);
  ctx.lineWidth = lineWidth;
  ctx.strokeStyle = isDuplicateHighlighted ? `rgba(255, 204, 0, ${finalOpacity})` : getLinkColor(weight, linkType, finalOpacity);

  if (isDuplicateHighlighted) {
    ctx.shadowBlur = 15;
    ctx.shadowColor = 'rgba(255, 204, 0, 0.8)';
  }

  const dash = getLineDash(linkType, weight);
  if (dash.length > 0) {
    ctx.setLineDash(dash);
  }
  ctx.stroke();
  ctx.setLineDash([]);
  ctx.shadowBlur = 0;
  ctx.shadowColor = 'transparent';
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
  ctx.shadowColor = 'rgba(138, 43, 226, 0.6)';

  // Semi-transparent sphere
  ctx.beginPath();
  ctx.arc(x, y, radius, 0, 2 * Math.PI);
  const gradient = ctx.createRadialGradient(x - radius * 0.3, y - radius * 0.3, radius * 0.1, x, y, radius);
  gradient.addColorStop(0, 'rgba(167, 139, 250, 0.4)');
  gradient.addColorStop(1, 'rgba(138, 43, 226, 0.15)');
  ctx.fillStyle = gradient;
  ctx.fill();

  // Border
  ctx.strokeStyle = `rgba(167, 139, 250, ${pulse})`;
  ctx.lineWidth = 2;
  ctx.stroke();

  // Question mark icon
  ctx.shadowBlur = 0;
  ctx.fillStyle = 'rgba(255, 255, 255, 0.9)';
  ctx.font = `${Math.floor(radius * 1.2)}px sans-serif`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText('?', x, y + radius * 0.05);

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
  ctx.fillStyle = '#cccccc';
  ctx.fill();
  ctx.strokeStyle = '#999999';
  ctx.lineWidth = 1;
  ctx.stroke();
  // Crater
  ctx.beginPath();
  ctx.arc(x - r * 0.3, y - r * 0.2, r * 0.25, 0, 2 * Math.PI);
  ctx.fillStyle = '#aaaaaa';
  ctx.fill();
}

/**
 * Seeded random number generator for deterministic anomaly parameters
 */
function seededRandom(seed: number): number {
  const x = Math.sin(seed) * 10000;
  return x - Math.floor(x);
}

/**
 * Generate deterministic parameters for anomaly visualization
 */
interface AnomalyParams {
  crackCount: number;
  tentacleCount: number;
  particleCount: number;
  colorShift1: number;
  colorShift2: number;
  deformAmount: number;
  rotationOffset: number;
  seedBase: number;
}

type AnomalyRenderer = (
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: AnomalyParams
) => void;

export function getAnomalyParams(nodeId: string): AnomalyParams {
  const hash = stringHash(nodeId);
  
  // Use different parts of hash for different parameters
  const hash1 = hash % 1000;
  const hash2 = (hash >> 10) % 1000;
  const hash3 = (hash >> 20) % 1000;
  const hash4 = (hash >> 25) % 1000;
  const hash5 = (hash >> 30) % 1000;
  
  const rr = anomalyConfig.reality_rift;
  const cm = anomalyConfig.chromatic_maw;
  const vw = anomalyConfig.void_whisper;
  const ca = anomalyConfig.cosmic_abomination;
  
  return {
    crackCount: ca.crack_count_min + Math.floor((hash1 / 1000) * (ca.crack_count_max - ca.crack_count_min)),
    tentacleCount: ca.tentacle_count_min + Math.floor((hash2 / 1000) * (ca.tentacle_count_max - ca.tentacle_count_min)),
    particleCount: ca.particle_count_min + Math.floor((hash3 / 1000) * (ca.particle_count_max - ca.particle_count_min)),
    colorShift1: (hash4 / 1000) * cm.hue_shift_range,
    colorShift2: (hash5 / 1000) * vw.hue_shift_range,
    deformAmount: rr.deform_amount_min + (hash1 / 1000) * (rr.deform_amount_max - rr.deform_amount_min),
    rotationOffset: (hash2 / 1000) * Math.PI * 2,
    seedBase: hash,
  };
}

/**
 * Draw a Reality Rift anomaly - dark core + jagged cracks + amoebic contour
 */
export function drawRealityRift(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: AnomalyParams
): void {
  const { crackCount, deformAmount, rotationOffset } = params;
  const rr = anomalyConfig.reality_rift;
  
  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(rotationOffset);
  
  // Dark core with purple glow
  ctx.beginPath();
  ctx.arc(0, 0, r * 0.3, 0, 2 * Math.PI);
  ctx.fillStyle = rr.core_color;
  ctx.shadowBlur = 20;
  ctx.shadowColor = rr.glow_color;
  ctx.fill();
  ctx.shadowBlur = 0;
  
  // Amoebic outer contour
  ctx.beginPath();
  const points = 24;
  for (let i = 0; i <= points; i++) {
    const angle = (i / points) * 2 * Math.PI;
    const deformation = 1 + deformAmount * Math.sin(angle * 5 + rotationOffset * 2);
    const radius = r * deformation;
    const px = Math.cos(angle) * radius;
    const py = Math.sin(angle) * radius;
    if (i === 0) ctx.moveTo(px, py);
    else ctx.lineTo(px, py);
  }
  ctx.closePath();
  ctx.strokeStyle = hexToRgba(rr.glow_color, 0.6);
  ctx.lineWidth = 2;
  ctx.stroke();
  ctx.fillStyle = hexToRgba(rr.core_color, 0.7);
  ctx.fill();
  
  const seedBase = params.seedBase;

  // Jagged cracks radiating from center
  for (let i = 0; i < crackCount; i++) {
    const angle = (i / crackCount) * 2 * Math.PI + rotationOffset;
    const crackLength = r * (0.5 + seededRandom(seedBase + i * 31 + 13) * 0.4);
    
    ctx.beginPath();
    ctx.moveTo(0, 0);
    
    // Create jagged path
    const segments = 3;
    for (let j = 0; j < segments; j++) {
      const segProgress = (j + 1) / segments;
      const segAngle = angle + (seededRandom(seedBase + i * 17 + j * 23 + 7) - 0.5) * 0.3;
      const segRadius = crackLength * segProgress;
      ctx.lineTo(Math.cos(segAngle) * segRadius, Math.sin(segAngle) * segRadius);
    }
    
    ctx.strokeStyle = hexToRgba(rr.glow_color, 0.4);
    ctx.lineWidth = 1;
    ctx.stroke();
  }
  
  ctx.restore();
}

/**
 * Draw a Chromatic Maw anomaly - tentacles + gradient core
 */
export function drawChromaticMaw(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: AnomalyParams
): void {
  const { tentacleCount, colorShift1, rotationOffset } = params;
  const cm = anomalyConfig.chromatic_maw;
  
  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(rotationOffset);
  
  // Gradient core
  const gradient = ctx.createRadialGradient(0, 0, 0, 0, 0, r * 0.4);
  gradient.addColorStop(0, `hsl(${cm.hue_shift_base + colorShift1}, 100%, 70%)`);
  gradient.addColorStop(0.5, `hsl(${cm.hue_shift_base - 100 + colorShift1}, 100%, 60%)`);
  gradient.addColorStop(1, `hsl(${cm.hue_shift_base - 280 + colorShift1}, 100%, 50%)`);
  
  ctx.beginPath();
  ctx.arc(0, 0, r * 0.4, 0, 2 * Math.PI);
  ctx.fillStyle = gradient;
  ctx.shadowBlur = 25;
  ctx.shadowColor = `hsl(${cm.hue_shift_base + colorShift1}, 100%, 50%)`;
  ctx.fill();
  ctx.shadowBlur = 0;
  
  const seedBase = params.seedBase;

  // Organic tentacles
  for (let i = 0; i < tentacleCount; i++) {
    const baseAngle = (i / tentacleCount) * 2 * Math.PI;
    const tentacleLength = r * (1.2 + seededRandom(seedBase + i * 29 + 11) * 0.4);
    const controlOffset = seededRandom(seedBase + i * 19 + 17) * r * 0.5;
    const endAngle = baseAngle + (seededRandom(seedBase + i * 23 + 31) - 0.5) * 0.5;
    
    ctx.beginPath();
    ctx.moveTo(0, 0);
    
    // Bezier curve for organic tentacle
    const cp1x = Math.cos(baseAngle) * r * 0.3;
    const cp1y = Math.sin(baseAngle) * r * 0.3;
    const cp2x = Math.cos(baseAngle + controlOffset) * r * 0.7;
    const cp2y = Math.sin(baseAngle + controlOffset) * r * 0.7;
    const endX = Math.cos(endAngle) * tentacleLength;
    const endY = Math.sin(endAngle) * tentacleLength;
    
    ctx.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, endX, endY);
    
    // Gradient along tentacle
    const tentacleGradient = ctx.createLinearGradient(0, 0, endX, endY);
    tentacleGradient.addColorStop(0, `hsla(${cm.hue_shift_base + colorShift1}, 100%, 60%, 0.8)`);
    tentacleGradient.addColorStop(0.5, `hsla(${cm.hue_shift_base - 100 + colorShift1}, 100%, 50%, 0.6)`);
    tentacleGradient.addColorStop(1, `hsla(${cm.hue_shift_base - 280 + colorShift1}, 100%, 40%, 0.3)`);
    
    ctx.strokeStyle = tentacleGradient;
    ctx.lineWidth = 3;
    ctx.lineCap = 'round';
    ctx.stroke();
  }
  
  ctx.restore();
}

/**
 * Draw a Void Whisper anomaly - particles + lines + snow effect
 */
export function drawVoidWhisper(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: AnomalyParams
): void {
  const { particleCount, colorShift2, rotationOffset } = params;
  const vw = anomalyConfig.void_whisper;
  
  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(rotationOffset);
  
  // Generate deterministic particle positions
  const particles: Array<{ x: number; y: number; opacity: number }> = [];
  const seedBase = params.seedBase;

  for (let i = 0; i < particleCount; i++) {
    const angle = seededRandom(seedBase + i * 13 + 3) * 2 * Math.PI;
    const distance = r * (0.5 + seededRandom(seedBase + i * 17 + 7) * 0.8);
    const px = Math.cos(angle) * distance;
    const py = Math.sin(angle) * distance;
    const opacity = 0.3 + seededRandom(seedBase + i * 19 + 11) * 0.5;
    particles.push({ x: px, y: py, opacity });
  }
  
  // Draw connections between nearby particles
  ctx.strokeStyle = `hsla(${vw.hue_shift_base + colorShift2}, 80%, 70%, 0.2)`;
  ctx.lineWidth = 0.5;
  
  for (let i = 0; i < particles.length; i++) {
    for (let j = i + 1; j < particles.length; j++) {
      const dx = particles[i].x - particles[j].x;
      const dy = particles[i].y - particles[j].y;
      const distance = Math.sqrt(dx * dx + dy * dy);
      
      if (distance < r * vw.connection_distance_threshold) {
        ctx.beginPath();
        ctx.moveTo(particles[i].x, particles[i].y);
        ctx.lineTo(particles[j].x, particles[j].y);
        ctx.stroke();
      }
    }
  }
  
  // Draw particles with twinkling effect
  for (let i = 0; i < particles.length; i++) {
    const p = particles[i];
    ctx.beginPath();
    ctx.arc(p.x, p.y, 1.5, 0, 2 * Math.PI);
    ctx.fillStyle = `hsla(${vw.hue_shift_base + colorShift2}, 80%, 80%, ${p.opacity})`;
    ctx.fill();
  }
  
  // Central faint glow
  const gradient = ctx.createRadialGradient(0, 0, 0, 0, 0, r);
  gradient.addColorStop(0, `hsla(${vw.hue_shift_base + colorShift2}, 80%, 60%, 0.3)`);
  gradient.addColorStop(1, 'transparent');
  ctx.beginPath();
  ctx.arc(0, 0, r, 0, 2 * Math.PI);
  ctx.fillStyle = gradient;
  ctx.fill();
  
  ctx.restore();
}

/**
 * Draw a Cosmic Abomination - combines all three anomaly types
 */
export function drawCosmicAbomination(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: AnomalyParams
): void {
  // Dark core from Reality Rift
  const { crackCount, tentacleCount, particleCount, deformAmount, rotationOffset, colorShift1, colorShift2 } = params;
  const rr = anomalyConfig.reality_rift;
  const cm = anomalyConfig.chromatic_maw;
  const vw = anomalyConfig.void_whisper;
  
  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(rotationOffset);
  
  // Dark core
  ctx.beginPath();
  ctx.arc(0, 0, r * 0.25, 0, 2 * Math.PI);
  ctx.fillStyle = rr.core_color;
  ctx.shadowBlur = 15;
  ctx.shadowColor = hexToRgba(rr.glow_color, 0.6);
  ctx.fill();
  ctx.shadowBlur = 0;
  
  // Amoebic contour (simplified)
  ctx.beginPath();
  const points = 20;
  for (let i = 0; i <= points; i++) {
    const angle = (i / points) * 2 * Math.PI;
    const deformation = 1 + deformAmount * 0.5 * Math.sin(angle * 4);
    const radius = r * 0.6 * deformation;
    const px = Math.cos(angle) * radius;
    const py = Math.sin(angle) * radius;
    if (i === 0) ctx.moveTo(px, py);
    else ctx.lineTo(px, py);
  }
  ctx.closePath();
  ctx.strokeStyle = `hsla(${cm.hue_shift_base + colorShift1}, 80%, 60%, 0.5)`;
  ctx.lineWidth = 2;
  ctx.stroke();
  
  const seedBase = params.seedBase;

  // Fewer tentacles (3-4)
  for (let i = 0; i < tentacleCount; i++) {
    const baseAngle = (i / tentacleCount) * 2 * Math.PI;
    const tentacleLength = r * (1.0 + seededRandom(seedBase + i * 23 + 5) * 0.3);
    const controlOffset = seededRandom(seedBase + i * 17 + 13) * r * 0.4;
    const endAngle = baseAngle + (seededRandom(seedBase + i * 19 + 29) - 0.5) * 0.4;
    
    ctx.beginPath();
    ctx.moveTo(0, 0);
    const cp1x = Math.cos(baseAngle) * r * 0.2;
    const cp1y = Math.sin(baseAngle) * r * 0.2;
    const cp2x = Math.cos(baseAngle + controlOffset) * r * 0.5;
    const cp2y = Math.sin(baseAngle + controlOffset) * r * 0.5;
    const endX = Math.cos(endAngle) * tentacleLength;
    const endY = Math.sin(endAngle) * tentacleLength;
    
    ctx.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, endX, endY);
    
    const tentacleGradient = ctx.createLinearGradient(0, 0, endX, endY);
    tentacleGradient.addColorStop(0, `hsla(${cm.hue_shift_base + colorShift1}, 100%, 50%, 0.7)`);
    tentacleGradient.addColorStop(1, `hsla(${cm.hue_shift_base - 280 + colorShift1}, 100%, 40%, 0.2)`);
    
    ctx.strokeStyle = tentacleGradient;
    ctx.lineWidth = 2;
    ctx.lineCap = 'round';
    ctx.stroke();
  }
  
  // Fewer particles (12-15)
  const particles: Array<{ x: number; y: number; opacity: number }> = [];
  for (let i = 0; i < particleCount; i++) {
    const angle = seededRandom(seedBase + i * 17 + 19) * 2 * Math.PI;
    const distance = r * (0.7 + seededRandom(seedBase + i * 23 + 7) * 0.5);
    const px = Math.cos(angle) * distance;
    const py = Math.sin(angle) * distance;
    const opacity = 0.4 + seededRandom(seedBase + i * 19 + 11) * 0.4;
    particles.push({ x: px, y: py, opacity });
  }
  
  // Draw particles
  for (let i = 0; i < particles.length; i++) {
    const p = particles[i];
    ctx.beginPath();
    ctx.arc(p.x, p.y, 1, 0, 2 * Math.PI);
    ctx.fillStyle = `hsla(${vw.hue_shift_base + colorShift2}, 80%, 70%, ${p.opacity})`;
    ctx.fill();
  }
  
  // Subtle cracks (2-3)
  for (let i = 0; i < crackCount; i++) {
    const angle = (i / crackCount) * Math.PI + rotationOffset;
    const crackLength = r * 0.4;
    
    ctx.beginPath();
    ctx.moveTo(0, 0);
    const segments = 2;
    for (let j = 0; j < segments; j++) {
      const segProgress = (j + 1) / segments;
      const segAngle = angle + (seededRandom(seedBase + i * 13 + j * 19 + 23) - 0.5) * 0.2;
      const segRadius = crackLength * segProgress;
      ctx.lineTo(Math.cos(segAngle) * segRadius, Math.sin(segAngle) * segRadius);
    }
    ctx.strokeStyle = `hsla(${cm.hue_shift_base + colorShift1}, 60%, 40%, 0.3)`;
    ctx.lineWidth = 1;
    ctx.stroke();
  }
  
  ctx.restore();
}

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
  const renderers = customRenderers ?? {
    0: drawRealityRift,
    1: drawChromaticMaw,
    2: drawVoidWhisper,
    3: drawCosmicAbomination,
  } as Record<number, AnomalyRenderer>;

  const rendererFn = renderers[anomalyType] ?? drawRealityRift;
  rendererFn(ctx, x, y, r, params);
}

/**
 * Resolve a link endpoint after d3-force: `source` / `target` may be id strings
 * or the same simulation node objects d3 mutates in place.
 */
function resolveLinkEndpoint(
  ref: string | SimulationNode,
  nodes: SimulationNode[]
): SimulationNode | undefined {
  if (typeof ref === 'object' && ref !== null) {
    return ref as SimulationNode;
  }
  return nodes.find((n) => String(n.id) === String(ref));
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
  highlightedLinkId?: string | null
): void {
  let drawnCount = 0;
  let skippedCount = 0;

  if (import.meta.env.DEV) {
    console.log(`[drawAllLinks] Called with ${simLinks.length} links and ${nodes.length} nodes`);
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
    const stableLinkId = `${link.source}-${link.target}-${link.link_type || 'related'}`;
    const isHighlighted = highlightedLinkId === stableLinkId;
    const pulseOpacity = isHighlighted ? 0.5 + 0.5 * Math.abs(Math.sin(animationTime / 150)) : 1;

    // Use animated link drawing
    drawLink(ctx, link, sourceNode, targetNode, opacity * pulseOpacity, hoveredNodeId, isHighlighted);
    drawnCount++;
  });

  if (import.meta.env.DEV && (drawnCount === 0 || skippedCount > 0)) {
    console.log(`[drawAllLinks] Total: ${simLinks.length}, Drawn: ${drawnCount}, Skipped: ${skippedCount}`);
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
  const type = node.type || 'unknown';

  // Get deterministic variation for this node (used for hue/size/phase).
  // We still apply variation in stable render mode so color/size remain deterministic,
  // but other random jitter/animation is suppressed via `disableVariation` flags.
  const variation = ['blackhole', 'debris', 'unknown'].includes(type) ? undefined : getVariation(node.id, type);

  // Apply micro-jitter to position (±1px) for "alive" feel unless stable rendering or focus mode is requested
  let x = node.x! + (disableVariation || focusMode ? 0 : (Math.random() - 0.5) * 2);
  let y = node.y! + (disableVariation || focusMode ? 0 : (Math.random() - 0.5) * 2);

  // For stable render mode snap to integer pixel positions to avoid
  // subpixel anti-aliasing differences between runs/environments.
  if (disableVariation || focusMode) {
    x = Math.round(x);
    y = Math.round(y);
  }

  const effectiveEnableShadows = enableShadows && !focusMode;

  switch (type) {
    case 'star':
      if (effectiveEnableShadows) {
        ctx.shadowBlur = 10;
        ctx.shadowColor = 'rgba(255, 200, 100, 0.8)';
      }
      drawStar(ctx, x, y, r, angle, variation, node.id, focusMode ? 0 : nodeCount, focusMode ? 0 : animationTime);
      break;
    case 'planet':
      drawPlanet(ctx, x, y, r, angle, undefined, variation, node.id, focusMode ? undefined : nodeCount, focusMode ? undefined : animationTime);
      break;
    case 'satellite':
      drawPlanet(ctx, x, y, r * 0.6, angle, '#aaaaaa', variation, node.id, focusMode ? undefined : nodeCount, focusMode ? undefined : animationTime);
      break;
    case 'comet':
      drawComet(ctx, x, y, r, angle, variation, node.id, focusMode ? undefined : nodeCount, focusMode ? undefined : animationTime);
      break;
    case 'galaxy':
      drawGalaxy(ctx, x, y, r, angle, variation, node.id, focusMode ? undefined : nodeCount, focusMode ? undefined : animationTime);
      break;
    case 'nebula':
      drawNebula(ctx, x, y, r * 1.5, angle);
      break;
    case 'asteroid':
      drawAsteroid(ctx, x, y, r, angle, variation, disableVariation || focusMode, node.id, focusMode ? undefined : nodeCount, focusMode ? undefined : animationTime);
      break;
    case 'debris':
      drawDebris(ctx, x, y, r, angle, disableVariation || focusMode);
      break;
    case 'blackhole':
      drawBlackhole(ctx, x, y, r, angle, node.id, focusMode ? undefined : nodeCount, focusMode ? undefined : animationTime);
      break;
    case 'technical':
      drawTechnicalNode(ctx, x, y, r, focusMode ? 0 : animationTime);
      break;
    case 'moon':
      drawMoon(ctx, x, y, r, angle);
      break;
    case 'unknown':
      drawUnknown(ctx, x, y, r, angle, node.id);
      break;
    default:
      drawUnknown(ctx, x, y, r, angle, node.id);
      break;
  }
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
  ctx.textAlign = 'center';
  ctx.textBaseline = 'top';
  let title = node.title || 'Untitled';
  if (title.length > 14) title = title.slice(0, 12) + '…';
  
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
export function getNodeColor(type: string | undefined): string {
  const colors: Record<string, string> = {
    star: '#ffcc00',
    planet: '#d6aa5d',
    comet: '#e879f9',
    galaxy: '#8b5cf6',
    asteroid: '#94a3b8',
    blackhole: '#000000',
    moon: '#cccccc',
    nebula: '#2dd4bf',
    dust: '#a1a1aa',
    inbox: '#fbbf24',
    unknown: '#94a3b8'
  };
  return colors[type || 'unknown'] || colors.unknown;
}

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
  particleSystem?: { initParticles: (id: string, x: number, y: number, color: string) => void; update: (id: string, x: number, y: number) => void; draw: (ctx: CanvasRenderingContext2D, id: string) => void; isEnabled: () => boolean } | null,
  focusMode: boolean = false,
  searchMatchIds?: string[]
): void {
  const r = 24;
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
      particleSystem.initParticles(node.id, node.x || 0, node.y || 0, getNodeColor(node.type));
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

    drawNode(ctx, node, r, angle, enableShadows, disableVariation, node.id, nodeCount, animationTime, focusMode);
    drawNodeTitle(ctx, node, r, finalOpacity, disableVariation);

    // Search match outline
    if (isSearchMatch && node.x != null && node.y != null) {
      ctx.save();
      ctx.beginPath();
      ctx.arc(node.x, node.y, r + 6, 0, 2 * Math.PI);
      ctx.strokeStyle = 'rgba(255, 204, 0, 0.9)';
      ctx.lineWidth = 3;
      ctx.shadowBlur = 15;
      ctx.shadowColor = 'rgba(255, 204, 0, 0.6)';
      ctx.stroke();
      ctx.restore();
    }

    // New note indicator (pulsing turquoise outline for 24 hours)
    if (!focusMode && isNewNode(node) && node.x != null && node.y != null) {
      const pulse = 0.5 + 0.5 * Math.abs(Math.sin(animationTime / 1000));
      ctx.save();
      ctx.beginPath();
      ctx.arc(node.x, node.y, r + 10, 0, 2 * Math.PI);
      ctx.strokeStyle = `rgba(45, 212, 191, ${pulse})`;
      ctx.lineWidth = 2;
      ctx.stroke();
      ctx.restore();
    }

    if (!focusMode && particleSystem?.isEnabled() && node.x && node.y) {
      particleSystem.update(node.id, node.x, node.y);
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
  disableVariation: boolean = false,
  animationTime: number = 0,
  hoveredNodeId: string | null = null,
  particleSystem?: { initParticles: (id: string, x: number, y: number, color: string) => void; update: (id: string, x: number, y: number) => void; draw: (ctx: CanvasRenderingContext2D, id: string) => void; isEnabled: () => boolean } | null,
  blackHole?: { x: number; y: number; radius: number; pulsePhase: number; hovered: boolean } | null,
  ghostNode?: { x: number; y: number; radius: number; hovered: boolean; pulsePhase: number; active: boolean } | null,
  gravitySystem?: { applyAttraction: (nodes: SimulationNode[]) => void; getDistortion: (x: number, y: number, nodes: SimulationNode[], maxDistance?: number) => { dx: number; dy: number }; isEnabled: (nodeCount: number) => boolean } | null,
  focusMode: boolean = false,
  searchMatchIds?: string[],
  highlightedLinkId?: string | null
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
    ctx.imageSmoothingQuality = 'low';
    ctx.shadowBlur = 0;
    ctx.shadowColor = 'transparent';
    ctx.lineJoin = 'round';
    ctx.lineCap = 'round';
    ctx.translate(0.5, 0.5);
  }
  ctx.translate(transform.x, transform.y);
  ctx.scale(transform.k, transform.k);

  // Draw links with animation
  drawAllLinks(ctx, simLinks, nodes, linkOpacity, animationTime, hoveredNodeId, highlightedLinkId);

  // Draw nodes with new effects
  const enableShadows = !focusMode && nodes.length < graphConfig2D.shadows_threshold;
  drawAllNodes(ctx, nodes, angles, enableShadows, nodeOpacity, disableVariation, animationTime, hoveredNodeId, particleSystem, focusMode, searchMatchIds);

  // Draw interactive UI elements in world coordinates (hidden in focus mode)
  if (!focusMode && blackHole) {
    drawBlackHole(ctx, blackHole, animationTime);
    if (blackHole.hovered) {
      drawBlackHoleTooltip(ctx, blackHole);
    }
  }

  if (!focusMode && ghostNode?.active) {
    drawGhostNode(ctx, ghostNode, animationTime);
    if (ghostNode.hovered) {
      drawGhostNodeTooltip(ctx, ghostNode);
    }
  }

  ctx.restore();
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
