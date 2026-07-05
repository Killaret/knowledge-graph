/**
 * Draw a Void Whisper anomaly - particles + lines + snow effect
 */
import type { AnomalyParams } from './helpers';
import { seededRandom } from './helpers';
import { hexToRgba } from '../utils';

export function drawVoidWhisper(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: AnomalyParams
): void {
  const { particleCount, colorShift2, rotationOffset } = params;
  const vw = globalThis.anomalyConfig.void_whisper;

  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(rotationOffset);

  // Generate deterministic particle positions
  const particles: Array<{ x: number; y: number; opacity: number }> = [];
  const seedBase = params.seedBase;

  for (let i = 0; i < particleCount; i++) {
    const angle = seededRandom(seedBase + i * 13 + 3) * 2 * Math.PI;
    const distance = r * (0.5 + seededRandom(seedBase + i * 17 + 7) * 0.8);
    const px = Math.cos(angle) * distance;
    const py = Math.sin(angle) * distance;
    const opacity = 0.3 + seededRandom(seedBase + i * 19 + 11) * 0.5;
    particles.push({ x: px, y: py, opacity });
  }

  // Draw connections between nearby particles
  ctx.strokeStyle = `hsla(${vw.hue_shift_base + colorShift2}, 80%, 70%, 0.2)`;
  ctx.lineWidth = 0.5;

  for (let i = 0; i < particles.length; i++) {
    for (let j = i + 1; j < particles.length; j++) {
      const dx = particles[i].x - particles[j].x;
      const dy = particles[i].y - particles[j].y;
      const distance = Math.sqrt(dx * dx + dy * dy);

      if (distance < r * vw.connection_distance_threshold) {
        ctx.beginPath();
        ctx.moveTo(particles[i].x, particles[i].y);
        ctx.lineTo(particles[j].x, particles[j].y);
        ctx.stroke();
      }
    }
  }

  // Draw particles with twinkling effect
  for (let i = 0; i < particles.length; i++) {
    const p = particles[i];
    ctx.beginPath();
    ctx.arc(p.x, p.y, 1.5, 0, 2 * Math.PI);
    ctx.fillStyle = `hsla(${vw.hue_shift_base + colorShift2}, 80%, 80%, ${p.opacity})`;
    ctx.fill();
  }

  // Central faint glow
  const gradient = ctx.createRadialGradient(0, 0, 0, 0, 0, r);
  gradient.addColorStop(0, `hsla(${vw.hue_shift_base + colorShift2}, 80%, 60%, 0.3)`);
  gradient.addColorStop(1, 'transparent');
  ctx.beginPath();
  ctx.arc(0, 0, r, 0, 2 * Math.PI);
  ctx.fillStyle = gradient;
  ctx.fill();

  ctx.restore();
}
