import type { AnomalyParams } from './helpers';

export function drawChromaticMaw(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _params: AnomalyParams
): void {
  ctx.save();
  ctx.translate(x, y);
  const gradient = ctx.createRadialGradient(0, 0, 0, 0, 0, r * 2);
  gradient.addColorStop(0, '#ff00aa');
  gradient.addColorStop(1, '#220011');
  ctx.fillStyle = gradient;
  ctx.beginPath();
  ctx.arc(0, 0, r, 0, Math.PI * 2);
  ctx.fill();
  ctx.restore();
}
