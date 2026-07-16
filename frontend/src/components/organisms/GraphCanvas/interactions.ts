/**
 * Interaction handlers for GraphCanvas
 */
import { goto } from '$app/navigation';
import type { SimulationNode, SimulationLink, TransformState, DragState } from './types';

export type { DragState };

/**
 * Calculate distance from point to line segment
 * Used for link hit detection
 */
function pointToLineDistance(
  px: number,
  py: number,
  x1: number,
  y1: number,
  x2: number,
  y2: number
): number {
  const A = px - x1;
  const B = py - y1;
  const C = x2 - x1;
  const D = y2 - y1;

  const dot = A * C + B * D;
  const lenSq = C * C + D * D;
  let param = -1;

  if (lenSq !== 0) {
    param = dot / lenSq;
  }

  let xx, yy;

  if (param < 0) {
    xx = x1;
    yy = y1;
  } else if (param > 1) {
    xx = x2;
    yy = y2;
  } else {
    xx = x1 + param * C;
    yy = y1 + param * D;
  }

  const dx = px - xx;
  const dy = py - yy;
  return Math.sqrt(dx * dx + dy * dy);
}

/**
 * Find link under cursor
 * Returns the link if within tolerance distance, null otherwise
 */
export function findLinkAtPosition(
  mouseX: number,
  mouseY: number,
  links: SimulationLink[],
  nodes: SimulationNode[],
  transform: TransformState,
  tolerance: number = 8
): SimulationLink | null {
  for (const link of links) {
    const sourceNode = typeof link.source === 'string'
      ? nodes.find((n) => n.id === link.source)
      : link.source;
    const targetNode = typeof link.target === 'string'
      ? nodes.find((n) => n.id === link.target)
      : link.target;

    if (!sourceNode || !targetNode) continue;
    if (sourceNode.x == null || sourceNode.y == null || targetNode.x == null || targetNode.y == null) continue;

    const distance = pointToLineDistance(
      mouseX,
      mouseY,
      sourceNode.x,
      sourceNode.y,
      targetNode.x,
      targetNode.y
    );

    if (distance <= tolerance) {
      return link;
    }
  }
  return null;
}

/**
 * Handle zoom with mouse wheel (zoom toward cursor, graph + links stay aligned)
 */
export function handleZoom(
  e: WheelEvent,
  transform: TransformState,
  canvas: HTMLCanvasElement,
  onDraw: () => void
): void {
  e.preventDefault();
  const delta = e.deltaY > 0 ? 0.95 : 1.05;
  const newK = transform.k * delta;
  if (newK < 0.2 || newK > 5) return;

  const rect = canvas.getBoundingClientRect();
  const mouseX = e.clientX - rect.left;
  const mouseY = e.clientY - rect.top;

  const gx = (mouseX - transform.x) / transform.k;
  const gy = (mouseY - transform.y) / transform.k;

  transform.k = newK;
  transform.x = mouseX - gx * transform.k;
  transform.y = mouseY - gy * transform.k;

  onDraw();
}

/**
 * Handle pan start (mousedown)
 */
export function handlePanStart(
  e: MouseEvent,
  dragState: DragState,
  transform: TransformState,
  canvas: HTMLCanvasElement
): void {
  dragState.dragging = true;
  dragState.dragStart = { x: e.clientX - transform.x, y: e.clientY - transform.y };
  canvas.style.cursor = 'grabbing';
}

/**
 * Handle pan move (mousemove)
 */
export function handlePanMove(
  e: MouseEvent,
  dragState: DragState,
  transform: TransformState,
  onDraw: () => void
): void {
  if (!dragState.dragging) return;
  transform.x = e.clientX - dragState.dragStart.x;
  transform.y = e.clientY - dragState.dragStart.y;
  onDraw();
}

/**
 * Handle pan end (mouseup)
 */
export function handlePanEnd(
  dragState: DragState,
  canvas: HTMLCanvasElement
): void {
  dragState.dragging = false;
  canvas.style.cursor = 'grab';
}

/**
 * Handle canvas click to detect node selection
 */
export function handleClick(
  e: MouseEvent,
  canvas: HTMLCanvasElement,
  transform: TransformState,
  nodes: SimulationNode[],
  onNodeClick?: (node: { id: string; title: string; type?: string }) => void
): void {
  const rect = canvas.getBoundingClientRect();
  const clickX = (e.clientX - rect.left - transform.x) / transform.k;
  const clickY = (e.clientY - rect.top - transform.y) / transform.k;
  const node = nodes.find((n: any) => {
    const dx = n.x! - clickX;
    const dy = n.y! - clickY;
    return Math.hypot(dx, dy) < 24;
  });
  if (node) {
    if (onNodeClick) {
      onNodeClick({ id: node.id, title: node.title, type: node.type });
    } else {
      goto(`/notes/${node.id}`);
    }
  }
}
