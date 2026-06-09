/**
 * Canvas rendering functions for GraphCanvas
 */
import { graphConfig2D, anomalyConfig } from '$lib/config';
import type { SimulationNode, SimulationLink } from './types';
import { getVariation, applyHueShift } from '$lib/utils/variation';

export type { SimulationNode, SimulationLink };

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

// Colors for different link types
const linkTypeColors: Record<string, string> = {
  reference: '#8b5cf6', // Purple - reference link
  dependency: '#ff3a2f', // Red - dependency
  related: '#6b7280', // Gray - related topic (default)
  custom: '#e879f9' // Bright purple - custom
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
 * Draw a star node
 */
export function drawStar(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  variation?: { sizeMultiplier: number; hueShift: number }
): void {
  const points = 5;
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const outerRadius = r * sizeMultiplier;
  const innerRadius = r * 0.4 * sizeMultiplier;
  let rot = angle;
  const step = Math.PI / points;

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
  
  const hueShift = variation?.hueShift ?? 0;
  ctx.fillStyle = applyHueShift('#ffcc00', hueShift);
  ctx.strokeStyle = applyHueShift('#cc9900', hueShift);
  ctx.lineWidth = 2;
  ctx.fill();
  ctx.stroke();
}

/**
 * Draw a planet node
 */
export function drawPlanet(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  color?: string,
  variation?: { sizeMultiplier: number; hueShift: number }
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const adjustedR = r * sizeMultiplier;
  const hueShift = variation?.hueShift ?? 0;
  
  ctx.beginPath();
  ctx.arc(x, y, adjustedR, 0, 2 * Math.PI);
  // Default planet color changed to warm gold to match visual baselines
  // Apply hueShift to both custom and default colors for visual variation
  ctx.fillStyle = color ? applyHueShift(color, hueShift) : applyHueShift('#d6aa5d', hueShift);
  ctx.fill();
  
  for (let i = -adjustedR / 2; i <= adjustedR / 2; i += adjustedR / 4) {
    ctx.beginPath();
    ctx.ellipse(x, y + i, adjustedR * 0.8, adjustedR * 0.15, angle, 0, 2 * Math.PI);
    // Inner band color: muted darker tone for default gold planet
    ctx.fillStyle = color ? 'rgba(100,100,100,0.3)' : '#b07a3a';
    ctx.fill();
  }
}

/**
 * Draw a comet node
 */
export function drawComet(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  variation?: { sizeMultiplier: number; hueShift: number }
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const adjustedR = r * sizeMultiplier;
  const hueShift = variation?.hueShift ?? 0;
  
  ctx.beginPath();
  ctx.arc(x, y, adjustedR, 0, 2 * Math.PI);
  ctx.fillStyle = applyHueShift('#e879f9', hueShift);
  ctx.fill();
  
  const tailLength = 40 * sizeMultiplier;
  const tailAngle = angle;
  const tipX = x + Math.cos(tailAngle) * tailLength;
  const tipY = y + Math.sin(tailAngle) * tailLength;
  
  ctx.beginPath();
  ctx.moveTo(x, y);
  ctx.lineTo(tipX, tipY);
  ctx.lineWidth = 4 * sizeMultiplier;
  ctx.strokeStyle = `rgba(${applyHueShiftToRGBA(232, 121, 249, hueShift)}, 0.6)`;
  ctx.stroke();
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
 * Draw a galaxy node
 */
export function drawGalaxy(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  variation?: { sizeMultiplier: number; hueShift: number }
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const hueShift = variation?.hueShift ?? 0;
  const adjustedR = r * sizeMultiplier;
  
  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(angle);
  
  for (let i = 0; i < 3; i++) {
    ctx.beginPath();
    ctx.ellipse(0, 0, adjustedR * (1 - i * 0.2), adjustedR * 0.4, 0, 0, 2 * Math.PI);
    const baseColor = applyHueShiftToRGBA(192, 132, 252, hueShift);
    ctx.fillStyle = `rgba(${baseColor}, ${0.3 - i * 0.1})`;
    ctx.fill();
  }
  
  ctx.restore();
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
 * Draw an asteroid node
 */
export function drawAsteroid(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number,
  variation?: { sizeMultiplier: number; hueShift: number },
  disableVariation: boolean = false
): void {
  const sizeMultiplier = variation?.sizeMultiplier ?? 1;
  const hueShift = variation?.hueShift ?? 0;
  const adjustedR = r * sizeMultiplier;
  
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
  
  ctx.fillStyle = applyHueShift('#94a3b8', hueShift);
  ctx.fill();
  ctx.strokeStyle = applyHueShift('#64748b', hueShift);
  ctx.lineWidth = 1;
  ctx.stroke();
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
 * Draw a black hole node
 */
export function drawBlackhole(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number
): void {
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
  ctx.arc(x, y, r * 0.8, 0, 2 * Math.PI);
  ctx.strokeStyle = '#ff3300';
  ctx.lineWidth = 2;
  ctx.stroke();
}

/**
 * Draw a moon node
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
 * Draw a link between two nodes
 */
export function drawLink(
  ctx: CanvasRenderingContext2D,
  link: SimulationLink,
  sourceNode: SimulationNode,
  targetNode: SimulationNode,
  opacity: number = 1
): void {
  ctx.beginPath();
  ctx.moveTo(sourceNode.x!, sourceNode.y!);
  ctx.lineTo(targetNode.x!, targetNode.y!);

  const weight = link.weight ?? 0.5;
  const linkType = link.link_type;

  // Line thickness depends on type and weight
  let lineWidth = Math.max(1, weight * 4);
  if (linkType === 'dependency') lineWidth *= 1.5;
  if (linkType === 'reference') lineWidth *= 0.8;

  ctx.lineWidth = lineWidth;
  ctx.strokeStyle = getLinkColor(weight, linkType, opacity);

  // Set dash pattern for dashed lines
  const dash = getLineDash(linkType, weight);
  if (dash.length > 0) {
    ctx.setLineDash(dash);
  } else {
    ctx.setLineDash([]);
  }

  ctx.stroke();
  ctx.setLineDash([]);
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
 * Draw all links
 */
export function drawAllLinks(
  ctx: CanvasRenderingContext2D,
  simLinks: SimulationLink[],
  nodes: SimulationNode[],
  linkOpacity?: Map<string, number>
): void {
  let drawnCount = 0;
  let skippedCount = 0;

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

    drawLink(ctx, link, sourceNode, targetNode, opacity);
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
  disableVariation: boolean = false
): void {
  const type = node.type || 'unknown';
  
  // Get deterministic variation for this node (used for hue/size/phase).
  // We still apply variation in stable render mode so color/size remain deterministic,
  // but other random jitter/animation is suppressed via `disableVariation` flags.
  const variation = ['blackhole', 'debris', 'unknown'].includes(type) ? undefined : getVariation(node.id, type);
  
  // Apply micro-jitter to position (±1px) for "alive" feel unless stable rendering is requested
  let x = node.x! + (disableVariation ? 0 : (Math.random() - 0.5) * 2);
  let y = node.y! + (disableVariation ? 0 : (Math.random() - 0.5) * 2);

  // For stable render mode snap to integer pixel positions to avoid
  // subpixel anti-aliasing differences between runs/environments.
  if (disableVariation) {
    x = Math.round(x);
    y = Math.round(y);
  }

  switch (type) {
    case 'star':
      if (enableShadows) {
        ctx.shadowBlur = 10;
        ctx.shadowColor = 'rgba(255, 200, 100, 0.8)';
      }
      drawStar(ctx, x, y, r, angle, variation);
      break;
    case 'planet':
      drawPlanet(ctx, x, y, r, angle, undefined, variation);
      break;
    case 'satellite':
      drawPlanet(ctx, x, y, r * 0.6, angle, '#aaaaaa', variation);
      break;
    case 'comet':
      drawComet(ctx, x, y, r, angle, variation);
      break;
    case 'galaxy':
      drawGalaxy(ctx, x, y, r, angle, variation);
      break;
    case 'nebula':
      drawNebula(ctx, x, y, r * 1.5, angle);
      break;
    case 'asteroid':
      drawAsteroid(ctx, x, y, r, angle, variation, disableVariation);
      break;
    case 'debris':
      drawDebris(ctx, x, y, r, angle, disableVariation);
      break;
    case 'blackhole':
      drawBlackhole(ctx, x, y, r, angle);
      break;
    case 'moon':
      drawMoon(ctx, x, y, r, angle);
      break;
    case 'unknown':
      drawUnknown(ctx, x, y, r, angle, node.id);
      break;
    default:
      if (enableShadows) {
        ctx.shadowBlur = 10;
        ctx.shadowColor = 'rgba(255, 200, 100, 0.8)';
      }
      drawStar(ctx, x, y, r, angle, variation);
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
 * Draw all nodes
 */
export function drawAllNodes(
  ctx: CanvasRenderingContext2D,
  nodes: SimulationNode[],
  angles: Map<string, number>,
  enableShadows: boolean,
  nodeOpacity?: Map<string, number>,
  disableVariation: boolean = false
): void {
  const r = 24; // radius increased for better readability

  // Ensure deterministic initial angles for stable render mode before any drawing
  if (disableVariation) {
    for (const node of nodes) {
      if (!angles.has(node.id)) {
        angles.set(node.id, 0);
      }
    }
  }

  nodes.forEach((node) => {
    const angle = angles.get(node.id) || 0;
    const opacity = nodeOpacity?.get(node.id) ?? 1;

    // Apply opacity using globalAlpha
    const previousAlpha = ctx.globalAlpha;
    ctx.globalAlpha = opacity;

    drawNode(ctx, node, r, angle, enableShadows, disableVariation);
    drawNodeTitle(ctx, node, r, opacity, disableVariation);

    // Restore previous alpha
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
  disableVariation: boolean = false
): void {
  ctx.clearRect(0, 0, width, height);

  // Stable render mode: reduce anti-aliasing sources and align to pixel grid
  ctx.save();
  if (disableVariation) {
    ctx.imageSmoothingEnabled = false;
    // some browsers support quality setting
    ctx.imageSmoothingQuality = 'low';
    ctx.shadowBlur = 0;
    ctx.shadowColor = 'transparent';
    ctx.lineJoin = 'round';
    ctx.lineCap = 'round';
    // small translate to reduce subpixel AA differences
    ctx.translate(0.5, 0.5);
  }
  ctx.translate(transform.x, transform.y);
  ctx.scale(transform.k, transform.k);

  // Draw links
  drawAllLinks(ctx, simLinks, nodes, linkOpacity);

  // Draw nodes
  const enableShadows = nodes.length < graphConfig2D.shadows_threshold;
  drawAllNodes(ctx, nodes, angles, enableShadows, nodeOpacity, disableVariation);

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
