/**
 * Black hole component for GraphCanvas
 * Used for drag-and-drop deletion of nodes
 */

import type { SimulationNode } from "./types";

export interface BlackHoleState {
  x: number;
  y: number;
  radius: number;
  pulsePhase: number;
  hovered: boolean;
  label?: string;
}

export const BLACK_HOLE_RADIUS = 40;
export const BLACK_HOLE_CATCH_RADIUS = 60;

export function createBlackHole(width: number, height: number): BlackHoleState {
  return {
    x: width - 84,
    y: height - 84,
    radius: BLACK_HOLE_RADIUS,
    pulsePhase: 0,
    hovered: false,
    label: "",
  };
}

export function updateBlackHolePosition(
  state: BlackHoleState,
  width: number,
  height: number
): void {
  state.x = width - 84;
  state.y = height - 84;
}

export function updateBlackHolePulse(state: BlackHoleState, animationTime: number): void {
  const pulseSpeed = 0.003;
  state.pulsePhase = Math.sin(animationTime * pulseSpeed) * 0.5 + 0.5;
}

export function isNodeOverBlackHole(
  node: SimulationNode,
  blackHole: BlackHoleState,
  transform: { x: number; y: number; k: number }
): boolean {
  if (node.x == null || node.y == null) return false;
  const screenX = node.x * transform.k + transform.x;
  const screenY = node.y * transform.k + transform.y;
  const dx = screenX - blackHole.x;
  const dy = screenY - blackHole.y;
  const distance = Math.sqrt(dx * dx + dy * dy);
  return distance < BLACK_HOLE_CATCH_RADIUS;
}

export function isPointOverBlackHole(x: number, y: number, blackHole: BlackHoleState): boolean {
  const dx = x - blackHole.x;
  const dy = y - blackHole.y;
  const distance = Math.sqrt(dx * dx + dy * dy);
  return distance < BLACK_HOLE_CATCH_RADIUS;
}

export function drawBlackHole(
  ctx: CanvasRenderingContext2D,
  blackHole: BlackHoleState,
  _animationTime: number
): void {
  const { x, y, radius, pulsePhase, hovered } = blackHole;

  ctx.save();

  // Pulsating scale when hovered
  const scale = hovered ? 1 + pulsePhase * 0.2 : 1;
  const scaledRadius = radius * scale;

  // Event horizon glow
  ctx.shadowBlur = 20 + pulsePhase * 10;
  ctx.shadowColor = "rgba(138, 43, 226, 0.6)";

  // Radial gradient from black center to dark purple edge
  const gradient = ctx.createRadialGradient(x, y, 0, x, y, scaledRadius);
  gradient.addColorStop(0, "#000000");
  gradient.addColorStop(0.6, "#1a0033");
  gradient.addColorStop(1, "#4b0082");

  ctx.beginPath();
  ctx.arc(x, y, scaledRadius, 0, Math.PI * 2);
  ctx.fillStyle = gradient;
  ctx.fill();

  // Accretion disk ring
  ctx.beginPath();
  ctx.arc(x, y, scaledRadius * 1.2, 0, Math.PI * 2);
  ctx.strokeStyle = `rgba(138, 43, 226, ${0.3 + pulsePhase * 0.3})`;
  ctx.lineWidth = 2;
  ctx.stroke();

  ctx.shadowBlur = 0;
  ctx.restore();
}

export function drawBlackHoleTooltip(
  ctx: CanvasRenderingContext2D,
  blackHole: BlackHoleState,
  text?: string
): void {
  const label = text ?? blackHole.label ?? "Drop here to delete";
  const { x, y, radius } = blackHole;

  ctx.save();
  ctx.font = "12px sans-serif";
  const textMetrics = ctx.measureText(label);
  const padding = 8;
  const boxWidth = textMetrics.width + padding * 2;
  const boxHeight = 24;
  const boxX = x - boxWidth / 2;
  const boxY = y - radius - boxHeight - 10;

  ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
  ctx.strokeStyle = "rgba(138, 43, 226, 0.6)";
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.roundRect(boxX, boxY, boxWidth, boxHeight, 4);
  ctx.fill();
  ctx.stroke();

  ctx.fillStyle = "#ffffff";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText(label, x, boxY + boxHeight / 2);

  ctx.restore();
}
