import type { SimulationNode } from '$lib/components/GraphCanvas/types';
import { getGlowIntensity } from '$shared/lib/graph/animations';

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

  // Draw star shape
  ctx.beginPath();
  for (let i = 0; i < points * 2; i++) {
    const radius = i % 2 === 0 ? outerRadius : innerRadius;
    const px = x + Math.cos(rot) * radius;
    const py = y + Math.sin(rot) * radius;
    if (i === 0) ctx.moveTo(px, py);
    else ctx.lineTo(px, py);
    rot += step;
  }
  ctx.closePath();

  // Apply gradient
  const gradient = ctx.createRadialGradient(x, y, 0, x, y, outerRadius);
  gradient.addColorStop(0, '#ffffcc');
  gradient.addColorStop(0.5, '#ffcc00');
  gradient.addColorStop(1, '#ff9900');
  ctx.fillStyle = gradient;
  ctx.fill();

  // Reset shadow
  ctx.shadowBlur = 0;
  ctx.shadowColor = 'transparent';
}
