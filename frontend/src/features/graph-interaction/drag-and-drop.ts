import type { TransformState, DragState, SimulationNode } from "$entities/graph-canvas/lib/types";
import type { BlackHoleState, GhostNodeState } from "$entities/graph-canvas/lib";
import { isPointOverGhostNode } from "$entities/graph-canvas/lib/ghost-node";
import { isPointOverBlackHole, isNodeOverBlackHole } from "$entities/graph-canvas/lib/black-hole";

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
  onBlackHoleDrop?: (nodeId: string) => void;
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

export function handleMouseMove(
  e: MouseEvent,
  canvas: HTMLCanvasElement,
  transform: TransformState,
  dragState: DragState,
  dragDropState: DragDropState,
  simNodes: SimulationNode[],
  blackHole: BlackHoleState,
  ghostNode: GhostNodeState,
  isTechnicalNode: (nodeId: string) => boolean,
  redraw: () => void
): void {
  const pos = getMouseWorldPosition(e, canvas, transform);
  dragDropState.mouseWorldPosition = pos;

  // Update hover states for interactive elements
  blackHole.hovered = isPointOverBlackHole(e.clientX, e.clientY, blackHole, transform);
  ghostNode.hovered = isPointOverGhostNode(e.clientX, e.clientY, ghostNode, transform);

  // Dragging a node
  if (dragDropState.draggedNodeId && dragState.dragging) {
    const node = simNodes.find((n) => n.id === dragDropState.draggedNodeId);
    if (node && node.x != null && node.y != null) {
      node.x = pos.x;
      node.y = pos.y;
      node.fx = pos.x;
      node.fy = pos.y;

      // Check if over black hole
      blackHole.hovered = isNodeOverBlackHole(node, blackHole);

      // Check if over another node for link creation
      const targetNode = findNodeAtPosition(pos.x, pos.y, simNodes);
      if (
        targetNode &&
        targetNode.id !== dragDropState.draggedNodeId &&
        !isTechnicalNode(targetNode.id)
      ) {
        dragDropState.isDraggingForLink = true;
        dragDropState.linkTargetNodeId = targetNode.id;
        dragDropState.linkPreviewTarget = {
          sourceId: dragDropState.draggedNodeId,
          targetId: targetNode.id,
        };
      } else {
        dragDropState.isDraggingForLink = false;
        dragDropState.linkTargetNodeId = null;
        dragDropState.linkPreviewTarget = null;
      }
    }
    redraw();
    return;
  }

  // Panning
  if (dragState.dragging) {
    transform.x = e.clientX - dragState.dragStart.x;
    transform.y = e.clientY - dragState.dragStart.y;
    redraw();
  }
}

export function handleMouseUp(
  e: MouseEvent,
  canvas: HTMLCanvasElement,
  transform: TransformState,
  dragState: DragState,
  dragDropState: DragDropState,
  simNodes: SimulationNode[],
  blackHole: BlackHoleState,
  ghostNode: GhostNodeState,
  isTechnicalNode: (nodeId: string) => boolean,
  callbacks: DragDropCallbacks,
  redraw: () => void
): void {
  dragState.dragging = false;
  canvas.style.cursor = "grab";

  if (dragDropState.draggedNodeId) {
    const node = simNodes.find((n) => n.id === dragDropState.draggedNodeId);
    if (node) {
      // Check if dropped on black hole
      if (isNodeOverBlackHole(node, blackHole)) {
        callbacks.onBlackHoleDrop?.(node.id);
      }
      // Check if dropped on another node for link creation
      else if (dragDropState.isDraggingForLink && dragDropState.linkTargetNodeId) {
        callbacks.onLinkPreview?.(dragDropState.draggedNodeId, dragDropState.linkTargetNodeId);
      }
      // Otherwise, release the node
      else {
        node.fx = undefined;
        node.fy = undefined;
      }
    }
    dragDropState.draggedNodeId = null;
    dragDropState.isDraggingForLink = false;
    dragDropState.linkSourceNodeId = null;
    dragDropState.linkTargetNodeId = null;
    dragDropState.linkPreviewTarget = null;
    redraw();
  }
}

export function handleClick(
  e: MouseEvent,
  canvas: HTMLCanvasElement,
  transform: TransformState,
  dragDropState: DragDropState,
  simNodes: SimulationNode[],
  ghostNode: GhostNodeState,
  isTechnicalNode: (nodeId: string) => boolean,
  onNodeClick?: (node: { id: string; title: string; type?: string }) => void,
  onGhostNodeClick?: () => void
): void {
  const pos = getMouseWorldPosition(e, canvas, transform);
  dragDropState.mouseWorldPosition = pos;

  // Ghost node click
  if (isPointOverGhostNode(e.clientX, e.clientY, ghostNode, transform)) {
    onGhostNodeClick?.();
    return;
  }

  // Node click
  const node = findNodeAtPosition(pos.x, pos.y, simNodes);
  if (node && !isTechnicalNode(node.id)) {
    onNodeClick?.({ id: node.id, title: node.title, type: node.type });
  }
}
