/**
 * Debris node renderer
 */

/**
 * Draw a debris node
 */
export function drawDebris(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number,
  disableVariation: boolean = false
): void {
  // Scattered small particles
  ctx.fillStyle = 'rgba(150, 150, 150, 0.6)';
  for (let i = 0; i < 5; i++) {
    const offsetX = disableVariation ? (i - 2) * (r * 0.25) : (Math.random() - 0.5) * r * 2;
    const offsetY = disableVariation ? ((i % 2 === 0 ? -1 : 1) * r * 0.2) : (Math.random() - 0.5) * r * 2;
    ctx.beginPath();
    ctx.arc(x + offsetX, y + offsetY, r * 0.3, 0, 2 * Math.PI);
    ctx.fill();
  }
}
