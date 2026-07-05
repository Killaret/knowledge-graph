/**
 * Comet node renderer
 */
import { getGlowIntensity } from '../utils';
import { applyHueShift } from '$shared/utils/variation';
import { applyHueShiftToRGBA } from '../utils/helpers';

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
