/**
 * Moon node renderer
 */

/**
 * Draw a moon node with glow
 */
export function drawMoon(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number
): void {
  // Moon body (grey circle)
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = '#cccccc';
  ctx.fill();
  ctx.strokeStyle = '#999999';
  ctx.lineWidth = 1;
  ctx.stroke();
  // Crater
  ctx.beginPath();
  ctx.arc(x - r * 0.3, y - r * 0.2, r * 0.25, 0, 2 * Math.PI);
  ctx.fillStyle = '#aaaaaa';
  ctx.fill();
}
