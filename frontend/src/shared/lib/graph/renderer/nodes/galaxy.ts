/**
 * Galaxy node renderer
 */
import { getGlowIntensity, getNodeGradient } from '../utils';
import { applyHueShift, applyHueShiftToRGBA } from '$lib/utils/variation';

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
