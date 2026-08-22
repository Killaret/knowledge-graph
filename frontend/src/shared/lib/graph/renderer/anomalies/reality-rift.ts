/**
 * Reality Rift anomaly renderer
 */
import { hexToRgba } from "../../helpers";
import { seededRandom } from "./helpers";

/**
 * Draw a Reality Rift anomaly - dark core + jagged cracks + amoebic contour
 */
export function drawRealityRift(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  params: import("./helpers").AnomalyParams
): void {
  const { crackCount, deformAmount, rotationOffset } = params;
  const rr = globalThis.anomalyConfig?.reality_rift || {
    core_color: "#6b21a8",
    glow_color: "#a855f7",
  };

  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(rotationOffset);

  // Dark core with purple glow
  ctx.beginPath();
  ctx.arc(0, 0, r * 0.3, 0, 2 * Math.PI);
  ctx.fillStyle = rr.core_color;
  ctx.shadowBlur = 20;
  ctx.shadowColor = rr.glow_color;
  ctx.fill();
  ctx.shadowBlur = 0;

  // Amoebic outer contour
  ctx.beginPath();
  const points = 24;
  for (let i = 0; i <= points; i++) {
    const angle = (i / points) * 2 * Math.PI;
    const deformation = 1 + deformAmount * Math.sin(angle * 5 + rotationOffset * 2);
    const radius = r * deformation;
    const px = Math.cos(angle) * radius;
    const py = Math.sin(angle) * radius;
    if (i === 0) ctx.moveTo(px, py);
    else ctx.lineTo(px, py);
  }
  ctx.closePath();
  ctx.strokeStyle = hexToRgba(rr.glow_color, 0.6);
  ctx.lineWidth = 2;
  ctx.stroke();
  ctx.fillStyle = hexToRgba(rr.core_color, 0.7);
  ctx.fill();

  const seedBase = params.seedBase;

  // Jagged cracks radiating from center
  for (let i = 0; i < crackCount; i++) {
    const angle = (i / crackCount) * 2 * Math.PI + rotationOffset;
    const crackLength = r * (0.5 + seededRandom(seedBase + i * 31 + 13) * 0.4);

    ctx.beginPath();
    ctx.moveTo(0, 0);

    // Create jagged path
    const segments = 3;
    for (let j = 0; j < segments; j++) {
      const segProgress = (j + 1) / segments;
      const segAngle = angle + (seededRandom(seedBase + i * 17 + j * 23 + 7) - 0.5) * 0.3;
      const segRadius = crackLength * segProgress;
      ctx.lineTo(Math.cos(segAngle) * segRadius, Math.sin(segAngle) * segRadius);
    }

    ctx.strokeStyle = hexToRgba(rr.glow_color, 0.4);
    ctx.lineWidth = 1;
    ctx.stroke();
  }

  ctx.restore();
}
