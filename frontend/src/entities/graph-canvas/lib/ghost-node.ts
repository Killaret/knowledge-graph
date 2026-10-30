/**
 * Ghost node — the translucent "+" button on the canvas for creating new notes.
 *
 * GHOST_NODE_RADIUS is sized relative to BASE_NODE_RADIUS (2x) so it is
 * bigger than a normal note, yet still belongs to the same visual scale.
 *
 * The node is rendered in screen coordinates, so it stays visible
 * regardless of canvas pan/zoom state.
 */

import type { SimulationNode } from "./types";
import { BASE_NODE_RADIUS, SERVICE_TOOL_MARGIN } from "./graph-constants";

/** Minimum and maximum zoom scale for service tools to keep them usable. */
const MIN_ZOOM_SCALE = 0.4;
const MAX_ZOOM_SCALE = 3;

export interface GhostNodeState {
  x: number;
  y: number;
  radius: number;
  hovered: boolean;
  pulsePhase: number;
  active: boolean;
}

/** Screen-space pixel radius of the ghost node button.
 *  Size is 2x the base note radius so it is clearly a service tool,
 *  but still visually related to the notes on the canvas. */
export const GHOST_NODE_RADIUS = BASE_NODE_RADIUS * 2;

export function createGhostNode(
  width: number,
  height: number,
  nodes: SimulationNode[]
): GhostNodeState {
  // Show in top-left if there are notes, otherwise center
  const hasNotes = nodes.length > 0;
  const radius = GHOST_NODE_RADIUS;
  const inset = SERVICE_TOOL_MARGIN + radius;
  return {
    x: hasNotes ? inset : width / 2,
    y: hasNotes ? inset : height / 2,
    radius,
    hovered: false,
    pulsePhase: 0,
    active: true,
  };
}

export function updateGhostNodePosition(
  state: GhostNodeState,
  width: number,
  height: number,
  nodes: SimulationNode[]
): void {
  const hasNotes = nodes.length > 0;
  const inset = SERVICE_TOOL_MARGIN + state.radius;
  state.x = hasNotes ? inset : width / 2;
  state.y = hasNotes ? inset : height / 2;
}

export function updateGhostNodeZoom(state: GhostNodeState, zoom: number): void {
  const scale = Math.min(MAX_ZOOM_SCALE, Math.max(MIN_ZOOM_SCALE, zoom));
  state.radius = GHOST_NODE_RADIUS * scale;
}

export function updateGhostNodePulse(state: GhostNodeState, animationTime: number): void {
  const pulseSpeed = 0.002;
  state.pulsePhase = Math.sin(animationTime * pulseSpeed) * 0.5 + 0.5;
}

export function isPointOverGhostNode(x: number, y: number, ghostNode: GhostNodeState): boolean {
  // The ghost node is drawn in screen coordinates, so hit testing
  // can be done directly in screen space.
  const dx = x - ghostNode.x;
  const dy = y - ghostNode.y;
  const distance = Math.sqrt(dx * dx + dy * dy);
  return distance < ghostNode.radius;
}

export function drawGhostNode(
  ctx: CanvasRenderingContext2D,
  ghostNode: GhostNodeState,
  _animationTime: number
): void {
  const { x, y, radius, hovered, pulsePhase } = ghostNode;

  const baseOpacity = hovered ? 0.7 : 0.4;
  const glowOpacity = 0.2 + pulsePhase * 0.2;
  const scale = 1 + pulsePhase * 0.05;

  ctx.save();

  // Glow effect
  ctx.shadowBlur = 15 + pulsePhase * 10;
  ctx.shadowColor = `rgba(255, 255, 255, ${glowOpacity})`;

  const gradient = ctx.createRadialGradient(
    x - radius * 0.3,
    y - radius * 0.3,
    0,
    x,
    y,
    radius * scale
  );
  gradient.addColorStop(0, `rgba(255, 255, 255, ${baseOpacity + 0.2})`);
  gradient.addColorStop(0.5, `rgba(200, 220, 255, ${baseOpacity})`);
  gradient.addColorStop(1, `rgba(255, 255, 255, 0)`);

  ctx.beginPath();
  ctx.arc(x, y, radius * scale, 0, Math.PI * 2);
  ctx.fillStyle = gradient;
  ctx.fill();

  // Plus sign
  ctx.shadowBlur = 0;
  ctx.strokeStyle = `rgba(255, 255, 255, ${baseOpacity + 0.3})`;
  ctx.lineWidth = 2;
  const plusSize = radius * 0.5;
  ctx.beginPath();
  ctx.moveTo(x - plusSize, y);
  ctx.lineTo(x + plusSize, y);
  ctx.moveTo(x, y - plusSize);
  ctx.lineTo(x, y + plusSize);
  ctx.stroke();

  ctx.restore();
}

/**
 * Draw ghost node in screen (pixel) coordinates — stays fixed on screen regardless of pan/zoom.
 * Screen position follows ghostNode.x/y so it stays inside the
 * cockpit frame and respects the service-tool margin.
 */
export function drawGhostNodeScreen(
  ctx: CanvasRenderingContext2D,
  ghostNode: GhostNodeState,
  _animationTime: number
): void {
  const x = ghostNode.x;
  const y = ghostNode.y;
  const radius = ghostNode.radius;
  const { hovered, pulsePhase } = ghostNode;

  const baseOpacity = hovered ? 0.7 : 0.4;
  const glowOpacity = 0.2 + pulsePhase * 0.2;
  const scale = 1 + pulsePhase * 0.05;

  ctx.save();
  ctx.shadowBlur = 15 + pulsePhase * 10;
  ctx.shadowColor = `rgba(255, 255, 255, ${glowOpacity})`;

  const gradient = ctx.createRadialGradient(
    x - radius * 0.3,
    y - radius * 0.3,
    0,
    x,
    y,
    radius * scale
  );
  gradient.addColorStop(0, `rgba(255, 255, 255, ${baseOpacity + 0.2})`);
  gradient.addColorStop(0.5, `rgba(200, 220, 255, ${baseOpacity})`);
  gradient.addColorStop(1, `rgba(255, 255, 255, 0)`);

  ctx.beginPath();
  ctx.arc(x, y, radius * scale, 0, Math.PI * 2);
  ctx.fillStyle = gradient;
  ctx.fill();

  ctx.shadowBlur = 0;
  ctx.strokeStyle = `rgba(255, 255, 255, ${baseOpacity + 0.3})`;
  ctx.lineWidth = 2;
  const plusSize = radius * 0.5;
  ctx.beginPath();
  ctx.moveTo(x - plusSize, y);
  ctx.lineTo(x + plusSize, y);
  ctx.moveTo(x, y - plusSize);
  ctx.lineTo(x, y + plusSize);
  ctx.stroke();
  ctx.restore();
}

export function drawGhostNodeTooltipScreen(
  ctx: CanvasRenderingContext2D,
  ghostNode: GhostNodeState,
  text: string = "Create new note"
): void {
  const x = ghostNode.x;
  const y = ghostNode.y;
  const radius = ghostNode.radius;

  ctx.save();
  ctx.font = "12px sans-serif";
  const textMetrics = ctx.measureText(text);
  const padding = 8;
  const boxWidth = textMetrics.width + padding * 2;
  const boxHeight = 24;
  const boxX = x - boxWidth / 2;
  const boxY = y + radius + 15;

  ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
  ctx.strokeStyle = "rgba(255, 255, 255, 0.3)";
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.roundRect(boxX, boxY, boxWidth, boxHeight, 4);
  ctx.fill();
  ctx.stroke();

  ctx.fillStyle = "#ffffff";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText(text, x, boxY + boxHeight / 2);
  ctx.restore();
}

export function drawGhostNodeTooltip(
  ctx: CanvasRenderingContext2D,
  ghostNode: GhostNodeState,
  text: string = "Create new note"
): void {
  const { x, y, radius } = ghostNode;

  ctx.save();
  ctx.font = "12px sans-serif";
  const textMetrics = ctx.measureText(text);
  const padding = 8;
  const boxWidth = textMetrics.width + padding * 2;
  const boxHeight = 24;
  const boxX = x - boxWidth / 2;
  const boxY = y + radius + 15;

  ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
  ctx.strokeStyle = "rgba(255, 255, 255, 0.3)";
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.roundRect(boxX, boxY, boxWidth, boxHeight, 4);
  ctx.fill();
  ctx.stroke();

  ctx.fillStyle = "#ffffff";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText(text, x, boxY + boxHeight / 2);

  ctx.restore();
}
