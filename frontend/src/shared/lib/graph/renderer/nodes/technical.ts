/**
 * Technical node renderer (Knowledge Core)
 */

/**
 * Draw the Knowledge Core technical node
 */
export function drawTechnicalNode(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  animationTime?: number
): void {
  const pulse = animationTime ? 0.7 + 0.3 * Math.abs(Math.sin(animationTime / 800)) : 1;
  const radius = r * 1.2;

  ctx.save();
  ctx.globalAlpha = 0.85;

  // Soft purple glow
  ctx.shadowBlur = 20 * pulse;
  ctx.shadowColor = 'rgba(138, 43, 226, 0.6)';

  // Semi-transparent sphere
  ctx.beginPath();
  ctx.arc(x, y, radius, 0, 2 * Math.PI);
  const gradient = ctx.createRadialGradient(x - radius * 0.3, y - radius * 0.3, radius * 0.1, x, y, radius);
  gradient.addColorStop(0, 'rgba(167, 139, 250, 0.4)');
  gradient.addColorStop(1, 'rgba(138, 43, 226, 0.15)');
  ctx.fillStyle = gradient;
  ctx.fill();

  // Border
  ctx.strokeStyle = `rgba(167, 139, 250, ${pulse})`;
  ctx.lineWidth = 2;
  ctx.stroke();

  // Question mark icon
  ctx.shadowBlur = 0;
  ctx.fillStyle = 'rgba(255, 255, 255, 0.9)';
  ctx.font = `${Math.floor(radius * 1.2)}px sans-serif`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText('?', x, y + radius * 0.05);

  ctx.restore();
}
