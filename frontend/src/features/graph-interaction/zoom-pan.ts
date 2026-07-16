import type { TransformState, SimulationNode } from '$components/organisms/GraphCanvas/types';
import { resetView } from '$components/organisms/GraphCanvas';

export interface ZoomPanState {
  lastTouchTime: number;
  lastTouchPos: { x: number; y: number };
  tapCount: number;
}

export function createZoomPanState(): ZoomPanState {
  return {
    lastTouchTime: 0,
    lastTouchPos: { x: 0, y: 0 },
    tapCount: 0
  };
}

export function handleZoom(
  e: WheelEvent,
  transform: TransformState,
  canvas: HTMLCanvasElement,
  redraw: () => void
): void {
  e.preventDefault();
  
  const rect = canvas.getBoundingClientRect();
  const mouseX = e.clientX - rect.left;
  const mouseY = e.clientY - rect.top;

  // Calculate zoom factor
  const zoomSensitivity = 0.001;
  const delta = -e.deltaY * zoomSensitivity;
  const newScale = Math.min(Math.max(transform.k * (1 + delta), 0.1), 5);

  // Calculate new transform to zoom towards mouse position
  const scaleChange = newScale / transform.k;
  transform.x = mouseX - (mouseX - transform.x) * scaleChange;
  transform.y = mouseY - (mouseY - transform.y) * scaleChange;
  transform.k = newScale;

  redraw();
}

export function handleTouchStart(
  e: TouchEvent,
  state: ZoomPanState,
  transform: TransformState,
  canvas: HTMLCanvasElement,
  simNodes: SimulationNode[],
  ctx: CanvasRenderingContext2D | null,
  width: number,
  height: number
): void {
  if (e.touches.length === 1) {
    const now = Date.now();
    const touch = e.touches[0];
    const dx = Math.abs(touch.clientX - state.lastTouchPos.x);
    const dy = Math.abs(touch.clientY - state.lastTouchPos.y);

    if (now - state.lastTouchTime < 300 && dx < 30 && dy < 30) {
      state.tapCount++;
      handleDoubleTap(state, touch.clientX, touch.clientY, transform, canvas, simNodes, ctx, width, height);
      e.preventDefault();
    } else {
      state.tapCount = 0;
    }

    state.lastTouchTime = now;
    state.lastTouchPos = { x: touch.clientX, y: touch.clientY };
  }
}

function handleDoubleTap(
  state: ZoomPanState,
  clientX: number,
  clientY: number,
  transform: TransformState,
  canvas: HTMLCanvasElement,
  simNodes: SimulationNode[],
  ctx: CanvasRenderingContext2D | null,
  width: number,
  height: number
): void {
  const rect = canvas.getBoundingClientRect();
  const x = (clientX - rect.left - transform.x) / transform.k;
  const y = (clientY - rect.top - transform.y) / transform.k;

  if (state.tapCount === 1) {
    // First double-tap: zoom in
    const newScale = transform.k * 2;
    const centerX = x * newScale;
    const centerY = y * newScale;

    transform.x = clientX - rect.left - centerX;
    transform.y = clientY - rect.top - centerY;
    transform.k = newScale;
  } else if (state.tapCount === 2) {
    // Second double-tap: reset view
    if (ctx && simNodes.length > 0) {
      resetView(ctx, width, height, simNodes, transform);
      state.tapCount = 0;
    }
  }
}

export function resetViewToCenter(
  transform: TransformState,
  width: number,
  height: number,
  simNodes: SimulationNode[]
): void {
  if (simNodes.length === 0) return;

  // Calculate bounding box of all nodes
  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
  for (const node of simNodes) {
    if (node.x != null && node.y != null) {
      minX = Math.min(minX, node.x);
      minY = Math.min(minY, node.y);
      maxX = Math.max(maxX, node.x);
      maxY = Math.max(maxY, node.y);
    }
  }

  // Calculate center and scale
  const centerX = (minX + maxX) / 2;
  const centerY = (minY + maxY) / 2;
  const nodeWidth = maxX - minX || 100;
  const nodeHeight = maxY - minY || 100;
  const padding = 100;
  const scaleX = (width - padding * 2) / nodeWidth;
  const scaleY = (height - padding * 2) / nodeHeight;
  const scale = Math.min(scaleX, scaleY, 1.5);

  // Apply transform
  transform.k = scale;
  transform.x = width / 2 - centerX * scale;
  transform.y = height / 2 - centerY * scale;
}
