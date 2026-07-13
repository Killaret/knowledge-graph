/**
 * Asteroid node renderer
 */
import { getGlowIntensity } from '$shared/lib/graph/renderer/utils';
import { applyHueShift } from '$lib/utils/variation';

function seededRand(seed: string, index: number): number {
  let h = 0;
  const s = seed + ':' + index;
  for (let i = 0; i < s.length; i++) { const c = s.charCodeAt(i); h = ((h << 5) - h) + c; h = h & h; }
  return Math.abs((h * 1664525 + 1013904223) & 0x7fffffff) / 0x7fffffff;
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

  const seed = nodeId ?? 'asteroid';

  // Irregular rocky shape — deterministic per node, no per-frame flickering
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
  ctx.strokeStyle = applyHueShift('#64748b', hueShift);
  ctx.lineWidth = 1;
  ctx.stroke();

  // Add craters (dark spots) — deterministic
  const craterCount = 3 + Math.floor(seededRand(seed, 100) * 3);
  for (let i = 0; i < craterCount; i++) {
    const craterAngle = seededRand(seed, 200 + i) * Math.PI * 2;
    const craterDist = seededRand(seed, 300 + i) * adjustedR * 0.6;
    const craterX = x + Math.cos(craterAngle) * craterDist;
    const craterY = y + Math.sin(craterAngle) * craterDist;
    const craterR = adjustedR * 0.15 * (0.5 + seededRand(seed, 400 + i) * 0.5);
    
    ctx.beginPath();
    ctx.arc(craterX, craterY, craterR, 0, 2 * Math.PI);
    ctx.fillStyle = 'rgba(0, 0, 0, 0.2)';
    ctx.fill();
  }

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = 'transparent';
}
