/**
 * Interaction handlers for GraphCanvas
 */
import { goto } from '$app/navigation';
import type { SimulationNode, TransformState, DragState } from './types';

export type { DragState };

/**
 * Handle zoom with mouse wheel
 */
export function handleZoom(
  e: WheelEvent,
  transform: TransformState,
  onDraw: () => void
): void {
  e.preventDefault();
  const delta = e.deltaY > 0 ? 0.95 : 1.05;
  const newK = transform.k * delta;
  if (newK < 0.2 || newK > 5) return;
  transform.k = newK;
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
