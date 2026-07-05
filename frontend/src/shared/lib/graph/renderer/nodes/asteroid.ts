/**
 * Asteroid node renderer
 */
import { getGlowIntensity } from '../utils';
import { applyHueShift } from '$lib/utils/variation';

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
