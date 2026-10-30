/**
 * Canvas background rendering for GraphCanvas
 */
import { graphConfig2D } from "$shared/config";
import type { SimulationNode } from "./types";

/**
 * Draw background grid or nebula
 */
export function drawBackground(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  nodes: SimulationNode[],
  time: number
): void {
  // Draw subtle grid
  const gridSize = 100;
  ctx.strokeStyle = "rgba(255, 255, 255, 0.03)";
  ctx.lineWidth = 1;

  for (let x = 0; x <= width; x += gridSize) {
    ctx.beginPath();
    ctx.moveTo(x, 0);
    ctx.lineTo(x, height);
    ctx.stroke();
  }

  for (let y = 0; y <= height; y += gridSize) {
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(width, y);
    ctx.stroke();
  }

  // Draw nebula (large blurred ellipses)
  if (nodes.length <= (graphConfig2D.visual_fx_threshold ?? 500)) {
    const nebulaCount = 3;
    for (let i = 0; i < nebulaCount; i++) {
      const nebulaX = (width / nebulaCount) * (i + 0.5);
      const nebulaY = height / 2;
      const nebulaRadius = 200 + Math.sin(time / 5000 + i) * 50;

      const gradient = ctx.createRadialGradient(
        nebulaX,
        nebulaY,
        0,
        nebulaX,
        nebulaY,
        nebulaRadius
      );
      gradient.addColorStop(0, "rgba(139, 92, 246, 0.05)");
      gradient.addColorStop(0.5, "rgba(59, 130, 246, 0.03)");
      gradient.addColorStop(1, "rgba(0, 0, 0, 0)");

      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, width, height);
    }
  }
}
