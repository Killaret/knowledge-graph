/**
 * Planet node renderer
 */
import { getGlowIntensity, getNodeGradient } from '$shared/lib/graph/renderer/utils';
import { applyHueShift } from '$lib/utils/variation';

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
