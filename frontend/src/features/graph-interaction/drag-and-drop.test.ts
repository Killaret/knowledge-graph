import { describe, it, expect, vi } from "vitest";
import {
  createDragDropState,
  getMouseWorldPosition,
  findNodeAtPosition,
  handleMouseDown,
} from "./drag-and-drop";
import { GHOST_NODE_RADIUS } from "$entities/graph-canvas/lib/ghost-node";

function createCanvas(): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  canvas.getBoundingClientRect = () => ({
    left: 0,
    top: 0,
    width: 800,
    height: 600,
    right: 800,
    bottom: 600,
    x: 0,
    y: 0,
    toJSON: () => "",
  });
  return canvas;
}

function createMouseEvent(clientX: number, clientY: number): MouseEvent {
  return new MouseEvent("mousedown", { clientX, clientY, bubbles: true });
}

describe("drag-and-drop", () => {
  it("creates drag/drop state", () => {
    const state = createDragDropState();
    expect(state.draggedNodeId).toBeNull();
    expect(state.isDraggingForLink).toBe(false);
  });

  it("calculates mouse world position", () => {
    const canvas = createCanvas();
    const event = createMouseEvent(100, 100);
    const transform = { x: 10, y: 20, k: 2 };

    const pos = getMouseWorldPosition(event, canvas, transform);

    expect(pos.x).toBe((100 - 0 - 10) / 2);
    expect(pos.y).toBe((100 - 0 - 20) / 2);
  });

  it("finds node at position", () => {
    const nodes = [
      { id: "n1", x: 0, y: 0, title: "A" },
      { id: "n2", x: 100, y: 100, title: "B" },
    ];

    expect(findNodeAtPosition(0, 1, nodes as any)?.id).toBe("n1");
    expect(findNodeAtPosition(500, 500, nodes as any)).toBeUndefined();
  });

  it("starts dragging a node", () => {
    const canvas = createCanvas();
    const transform = { x: 0, y: 0, k: 1 };
    const dragState = { dragging: false, dragStart: { x: 0, y: 0 } };
    const dragDropState = createDragDropState();
    const simNodes = [{ id: "n1", x: 10, y: 10, title: "A" }];
    const ghostNode = {
      x: 60,
      y: 60,
      radius: GHOST_NODE_RADIUS,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    const callbacks = {};

    handleMouseDown(
      createMouseEvent(10, 10),
      canvas,
      transform,
      dragState as any,
      dragDropState,
      simNodes as any,
      ghostNode,
      () => false,
      callbacks
    );

    expect(dragDropState.draggedNodeId).toBe("n1");
    expect(dragState.dragging).toBe(true);
    expect(canvas.style.cursor).toBe("grabbing");
  });

  it("ignores technical nodes", () => {
    const canvas = createCanvas();
    const event = createMouseEvent(10, 10);
    const dragState = { dragging: false, dragStart: { x: 0, y: 0 } };
    const dragDropState = createDragDropState();
    const simNodes = [{ id: "n1", x: 10, y: 10, title: "A" }];
    const ghostNode = {
      x: 60,
      y: 60,
      radius: GHOST_NODE_RADIUS,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };

    const preventDefault = vi.spyOn(event, "preventDefault");
    handleMouseDown(
      event,
      canvas,
      { x: 0, y: 0, k: 1 },
      dragState as any,
      dragDropState,
      simNodes as any,
      ghostNode,
      () => true,
      {}
    );

    expect(preventDefault).toHaveBeenCalled();
    expect(dragDropState.draggedNodeId).toBeNull();
  });

  it("starts pan on empty space", () => {
    const canvas = createCanvas();
    const dragState = { dragging: false, dragStart: { x: 0, y: 0 } };
    const dragDropState = createDragDropState();
    const ghostNode = {
      x: 60,
      y: 60,
      radius: GHOST_NODE_RADIUS,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };

    handleMouseDown(
      createMouseEvent(400, 300),
      canvas,
      { x: 0, y: 0, k: 1 },
      dragState as any,
      dragDropState,
      [],
      ghostNode,
      () => false,
      {}
    );

    expect(dragState.dragging).toBe(true);
    expect(dragDropState.draggedNodeId).toBeNull();
    expect(dragState.dragStart).toEqual({ x: 400, y: 300 });
  });

  it("ghost node click opens note form", () => {
    const canvas = createCanvas();
    const dragState = { dragging: false, dragStart: { x: 0, y: 0 } };
    const dragDropState = createDragDropState();
    const ghostNode = {
      x: 60,
      y: 60,
      radius: GHOST_NODE_RADIUS,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    const callbacks = { onNodeDragStart: vi.fn() };

    handleMouseDown(
      createMouseEvent(60, 60),
      canvas,
      { x: 0, y: 0, k: 1 },
      dragState as any,
      dragDropState,
      [],
      ghostNode,
      () => false,
      callbacks
    );

    expect(callbacks.onNodeDragStart).toHaveBeenCalledWith("ghost");
  });
});
