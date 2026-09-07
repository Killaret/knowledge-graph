import type { AnomalyParams } from "./helpers";

export function drawCosmicAbomination(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _params: AnomalyParams
): void {
  ctx.save();
  ctx.translate(x, y);
  const gradient = ctx.createRadialGradient(0, 0, 0, 0, 0, r * 2);
  gradient.addColorStop(0, "#6600ff");
  gradient.addColorStop(1, "#000000");
  ctx.fillStyle = gradient;
  ctx.beginPath();
  ctx.arc(0, 0, r, 0, Math.PI * 2);
  ctx.fill();
  ctx.restore();
}
