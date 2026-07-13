/**
 * Debris node renderer
 */

function seededRand(seed: string, index: number): number {
  let h = 0;
  const s = seed + ':' + index;
  for (let i = 0; i < s.length; i++) { const c = s.charCodeAt(i); h = ((h << 5) - h) + c; h = h & h; }
  return Math.abs((h * 1664525 + 1013904223) & 0x7fffffff) / 0x7fffffff;
}

/**
 * Draw a debris node
 */
export function drawDebris(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number,
  disableVariation: boolean = false,
  nodeId?: string
): void {
  const seed = nodeId ?? 'debris';
  // Scattered small particles — deterministic positions per node
  ctx.fillStyle = 'rgba(150, 150, 150, 0.6)';
  for (let i = 0; i < 5; i++) {
    const offsetX = disableVariation ? (i - 2) * (r * 0.25) : (seededRand(seed, i) - 0.5) * r * 2;
    const offsetY = disableVariation ? ((i % 2 === 0 ? -1 : 1) * r * 0.2) : (seededRand(seed, 10 + i) - 0.5) * r * 2;
    ctx.beginPath();
    ctx.arc(x + offsetX, y + offsetY, r * 0.3, 0, 2 * Math.PI);
    ctx.fill();
  }
}
