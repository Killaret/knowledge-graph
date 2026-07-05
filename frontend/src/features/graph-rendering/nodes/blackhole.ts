/**
 * Black hole node renderer
 */
import { getGlowIntensity } from '$shared/lib/graph/renderer/utils';

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
