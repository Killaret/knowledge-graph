/**
 * Star node renderer
 */
import { getGlowIntensity } from '../utils';
import { getNodeGradient } from '../utils';
import { applyHueShift } from '$lib/utils/variation';

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
