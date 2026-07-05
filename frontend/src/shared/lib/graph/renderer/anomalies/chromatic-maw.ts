/**
 * Draw a Chromatic Maw anomaly - tentacles + gradient core
 */
import type { AnomalyParams } from './helpers';
import { seededRandom } from './helpers';
import { hexToRgba } from '../utils';

export function drawChromaticMaw(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: AnomalyParams
): void {
  const { tentacleCount, colorShift1, rotationOffset } = params;
  const cm = globalThis.anomalyConfig.chromatic_maw;

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
