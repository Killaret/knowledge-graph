import { describe, it, expect, vi } from "vitest";
import {
  createDragDropState,
  getMouseWorldPosition,
  findNodeAtPosition,
  handleMouseDown,
  handleMouseMove,
  handleMouseUp,
  handleClick,
} from "./drag-and-drop";
import { graphConfig2D } from "$shared/config";

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
      radius: graphConfig2D.ghost_node_radius,
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
      callbacks,
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
      radius: graphConfig2D.ghost_node_radius,
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
      {},
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
      radius: graphConfig2D.ghost_node_radius,
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
      {},
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
      radius: graphConfig2D.ghost_node_radius,
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
      callbacks,
    );

    expect(callbacks.onNodeDragStart).toHaveBeenCalledWith("ghost");
  });

  it("moves dragged node and detects link target", () => {
    const canvas = createCanvas();
    const transform = { x: 0, y: 0, k: 1 };
    const dragState = { dragging: true, dragStart: { x: 0, y: 0 } };
    const dragDropState = createDragDropState();
    dragDropState.draggedNodeId = "n1";
    // n2 is listed first so findNodeAtPosition returns the target, not the dragged node
    const simNodes = [
      { id: "n2", x: 50, y: 0, title: "B" },
      { id: "n1", x: 0, y: 0, title: "A" },
    ];
    const blackHole = {
      x: 740,
      y: 540,
      radius: 40,
      pulsePhase: 0,
      hovered: false,
    };
    const ghostNode = {
      x: 60,
      y: 60,
      radius: graphConfig2D.ghost_node_radius,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    const redraw = vi.fn();

    handleMouseMove(
      createMouseEvent(50, 0),
      canvas,
      transform,
      dragState as any,
      dragDropState,
      simNodes as any,
      blackHole,
      ghostNode,
      () => false,
      redraw,
    );

    expect(dragDropState.isDraggingForLink).toBe(true);
    expect(dragDropState.linkTargetNodeId).toBe("n2");
    expect(simNodes[1].x).toBe(50);
    expect(redraw).toHaveBeenCalled();
  });

  it("pans canvas", () => {
    const canvas = createCanvas();
    const transform = { x: 0, y: 0, k: 1 };
    const dragState = { dragging: true, dragStart: { x: 0, y: 0 } };
    const dragDropState = createDragDropState();
    const blackHole = {
      x: 740,
      y: 540,
      radius: 40,
      pulsePhase: 0,
      hovered: false,
    };
    const ghostNode = {
      x: 60,
      y: 60,
      radius: graphConfig2D.ghost_node_radius,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    const redraw = vi.fn();

    handleMouseMove(
      createMouseEvent(100, 50),
      canvas,
      transform,
      dragState as any,
      dragDropState,
      [],
      blackHole,
      ghostNode,
      () => false,
      redraw,
    );

    expect(transform.x).toBe(100);
    expect(transform.y).toBe(50);
    expect(redraw).toHaveBeenCalled();
  });

  it("drops node on black hole", () => {
    const canvas = createCanvas();
    const dragState = { dragging: true, dragStart: { x: 0, y: 0 } };
    const dragDropState = createDragDropState();
    dragDropState.draggedNodeId = "n1";
    const simNodes = [{ id: "n1", x: 740, y: 540, title: "A" }];
    const blackHole = {
      x: 740,
      y: 540,
      radius: 40,
      pulsePhase: 0,
      hovered: false,
    };
    const ghostNode = {
      x: 60,
      y: 60,
      radius: graphConfig2D.ghost_node_radius,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    const callbacks = { onBlackHoleDrop: vi.fn() };
    const redraw = vi.fn();

    handleMouseUp(
      createMouseEvent(740, 540),
      canvas,
      { x: 0, y: 0, k: 1 },
      dragState as any,
      dragDropState,
      simNodes as any,
      blackHole,
      ghostNode,
      () => false,
      callbacks,
      redraw,
    );

    expect(callbacks.onBlackHoleDrop).toHaveBeenCalledWith("n1");
    expect(dragDropState.draggedNodeId).toBeNull();
    expect(redraw).toHaveBeenCalled();
  });

  it("drops node on another node to create link preview", () => {
    const canvas = createCanvas();
    const dragState = { dragging: true, dragStart: { x: 0, y: 0 } };
    const dragDropState = createDragDropState();
    dragDropState.draggedNodeId = "n1";
    dragDropState.isDraggingForLink = true;
    dragDropState.linkTargetNodeId = "n2";
    const simNodes = [
      { id: "n1", x: 0, y: 0, title: "A" },
      { id: "n2", x: 0, y: 0, title: "B" },
    ];
    const blackHole = {
      x: 740,
      y: 540,
      radius: 40,
      pulsePhase: 0,
      hovered: false,
    };
    const ghostNode = {
      x: 60,
      y: 60,
      radius: graphConfig2D.ghost_node_radius,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    const callbacks = { onLinkPreview: vi.fn() };
    const redraw = vi.fn();

    handleMouseUp(
      createMouseEvent(0, 0),
      canvas,
      { x: 0, y: 0, k: 1 },
      dragState as any,
      dragDropState,
      simNodes as any,
      blackHole,
      ghostNode,
      () => false,
      callbacks,
      redraw,
    );

    expect(callbacks.onLinkPreview).toHaveBeenCalledWith("n1", "n2");
  });

  it("releases node when dropped on empty space", () => {
    const canvas = createCanvas();
    const dragState = { dragging: true, dragStart: { x: 0, y: 0 } };
    const dragDropState = createDragDropState();
    dragDropState.draggedNodeId = "n1";
    const simNodes = [{ id: "n1", x: 100, y: 100, title: "A" }];
    const blackHole = {
      x: 740,
      y: 540,
      radius: 40,
      pulsePhase: 0,
      hovered: false,
    };
    const ghostNode = {
      x: 60,
      y: 60,
      radius: graphConfig2D.ghost_node_radius,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    const redraw = vi.fn();

    handleMouseUp(
      createMouseEvent(100, 100),
      canvas,
      { x: 0, y: 0, k: 1 },
      dragState as any,
      dragDropState,
      simNodes as any,
      blackHole,
      ghostNode,
      () => false,
      {},
      redraw,
    );

    expect((simNodes[0] as any).fx).toBeUndefined();
    expect((simNodes[0] as any).fy).toBeUndefined();
    expect(canvas.style.cursor).toBe("grab");
  });

  it("handles click on node and ghost", () => {
    const canvas = createCanvas();
    const dragDropState = createDragDropState();
    const simNodes = [{ id: "n1", x: 10, y: 10, title: "A" }];
    const ghostNode = {
      x: 60,
      y: 60,
      radius: graphConfig2D.ghost_node_radius,
      hovered: false,
      pulsePhase: 0,
      active: true,
    };
    const onNodeClick = vi.fn();
    const onGhostNodeClick = vi.fn();

    handleClick(
      createMouseEvent(10, 10),
      canvas,
      { x: 0, y: 0, k: 1 },
      dragDropState,
      simNodes as any,
      ghostNode,
      () => false,
      onNodeClick,
      onGhostNodeClick,
    );

    expect(onNodeClick).toHaveBeenCalledWith({
      id: "n1",
      title: "A",
      type: undefined,
    });

    handleClick(
      createMouseEvent(60, 60),
      canvas,
      { x: 0, y: 0, k: 1 },
      dragDropState,
      simNodes as any,
      ghostNode,
      () => false,
      onNodeClick,
      onGhostNodeClick,
    );

    expect(onGhostNodeClick).toHaveBeenCalled();
  });
});
