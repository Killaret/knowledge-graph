import type { TransformState, DragState, SimulationNode } from "$entities/graph-canvas/lib/types";
import type { GhostNodeState } from "$entities/graph-canvas/lib";
import { isPointOverGhostNode } from "$entities/graph-canvas/lib/ghost-node";

export interface DragDropState {
  draggedNodeId: string | null;
  dragStartPosition: { x: number; y: number };
  isDraggingForLink: boolean;
  linkSourceNodeId: string | null;
  linkTargetNodeId: string | null;
  mouseWorldPosition: { x: number; y: number };
  linkPreviewTarget: { sourceId: string; targetId: string } | null;
}

export interface DragDropCallbacks {
  onNodeDragStart?: (nodeId: string) => void;
  onNodeDragEnd?: (nodeId: string) => void;
  onLinkPreview?: (sourceId: string, targetId: string) => void;
  onSingularityDrop?: (nodeId: string) => void;
}

export function createDragDropState(): DragDropState {
  return {
    draggedNodeId: null,
    dragStartPosition: { x: 0, y: 0 },
    isDraggingForLink: false,
    linkSourceNodeId: null,
    linkTargetNodeId: null,
    mouseWorldPosition: { x: 0, y: 0 },
    linkPreviewTarget: null,
  };
}

export function getMouseWorldPosition(
  e: MouseEvent,
  canvas: HTMLCanvasElement,
  transform: TransformState
): { x: number; y: number } {
  const rect = canvas.getBoundingClientRect();
  return {
    x: (e.clientX - rect.left - transform.x) / transform.k,
    y: (e.clientY - rect.top - transform.y) / transform.k,
  };
}

export function findNodeAtPosition(
  x: number,
  y: number,
  simNodes: SimulationNode[]
): SimulationNode | undefined {
  for (const node of simNodes) {
    if (node.x == null || node.y == null) continue;
    const dx = node.x - x;
    const dy = node.y - y;
    if (Math.sqrt(dx * dx + dy * dy) < 30) {
      return node;
    }
  }
  return undefined;
}

export function handleMouseDown(
  e: MouseEvent,
  canvas: HTMLCanvasElement,
  transform: TransformState,
  dragState: DragState,
  dragDropState: DragDropState,
  simNodes: SimulationNode[],
  ghostNode: GhostNodeState,
  isTechnicalNode: (nodeId: string) => boolean,
  callbacks: DragDropCallbacks
): void {
  const pos = getMouseWorldPosition(e, canvas, transform);
  dragDropState.mouseWorldPosition = pos;

  // Ghost node is drawn in screen coords at (60, 60) — check in screen space
  const ghostScreenX = canvas.getBoundingClientRect().left + 60;
  const ghostScreenY = canvas.getBoundingClientRect().top + 60;
  const gdx = e.clientX - ghostScreenX;
  const gdy = e.clientY - ghostScreenY;
  if (Math.sqrt(gdx * gdx + gdy * gdy) < ghostNode.radius) {
    callbacks.onNodeDragStart?.("ghost");
    return;
  }

  // Node click -> start dragging for link/move/delete
  const node = findNodeAtPosition(pos.x, pos.y, simNodes);
  if (node) {
    if (isTechnicalNode(node.id)) {
      // Technical nodes are not draggable
      e.preventDefault();
      return;
    }
    dragDropState.draggedNodeId = node.id;
    dragDropState.dragStartPosition = { x: node.x!, y: node.y! };
    dragState.dragging = true;
    canvas.style.cursor = "grabbing";
    e.preventDefault();
    return;
  }

  // Empty space -> pan
  dragState.dragStart = {
    x: e.clientX - transform.x,
    y: e.clientY - transform.y,
  };
  dragState.dragging = true;
  canvas.style.cursor = "grabbing";
}
