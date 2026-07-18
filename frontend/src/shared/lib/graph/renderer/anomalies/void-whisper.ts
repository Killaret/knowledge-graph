import type { AnomalyParams } from "./helpers";

export function drawVoidWhisper(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _params: AnomalyParams,
): void {
  ctx.save();
  ctx.translate(x, y);
  ctx.fillStyle = "#110022";
  ctx.beginPath();
  ctx.arc(0, 0, r, 0, Math.PI * 2);
  ctx.fill();
  ctx.strokeStyle = "#8866ff";
  ctx.stroke();
  ctx.restore();
}
