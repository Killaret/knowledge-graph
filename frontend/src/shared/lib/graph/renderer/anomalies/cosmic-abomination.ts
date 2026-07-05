/**
 * Draw a Cosmic Abomination - combines all three anomaly types
 */
import type { AnomalyParams } from './helpers';
import { seededRandom } from './helpers';
import { hexToRgba } from '../utils';

export function drawCosmicAbomination(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: AnomalyParams
): void {
  // Dark core from Reality Rift
  const { crackCount, tentacleCount, particleCount, deformAmount, rotationOffset, colorShift1, colorShift2 } = params;
  const rr = globalThis.anomalyConfig.reality_rift;
  const cm = globalThis.anomalyConfig.chromatic_maw;
  const vw = globalThis.anomalyConfig.void_whisper;

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
