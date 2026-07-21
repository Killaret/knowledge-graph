import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  createGraphEventBridge,
  type GraphCanvasEventContext,
} from "./event-bridge";
import { createHotkeysState } from "./hotkeys";
import { createZoomPanState } from "./zoom-pan";
import { createDragDropState } from "./drag-and-drop";
import { createNoteFormState } from "$features/graph-forms/note-form";
import { createLinkFormState } from "$features/graph-forms/link-form";
import { createGhostNode } from "$components/organisms/GraphCanvas/ghost-node";
import { createBlackHole } from "$components/organisms/GraphCanvas/black-hole";
import type {
  SimulationNode,
  SimulationLink,
} from "$components/organisms/GraphCanvas/types";

function createCanvas(): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  canvas.width = 800;
  canvas.height = 600;
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

function createMockContext(canvas: HTMLCanvasElement) {
  let hoveredNodeId: string | null = null;
  let hoveredLink: any = null;
  let tooltipPosition = { x: 0, y: 0 };
  let focusMode = false;
  let selectedNodeId: string | null = null;
  let ghostNode = createGhostNode(800, 600, []);

  const blackHole = createBlackHole(800, 600);
  const simNodes: SimulationNode[] = [
    { id: "n2", x: 50, y: 10, title: "B" },
    { id: "n1", x: 0, y: 10, title: "A" },
  ];
  const simLinks: SimulationLink[] = [
    {
      source: simNodes[1],
      target: simNodes[0],
      link_type: "related",
      weight: 1,
    },
  ];

  return {
    readonly: false,
    browser: true,
    isTechnicalNode: vi.fn(() => false),
    getCanvas: () => canvas,
    getCtx: () => ({}) as CanvasRenderingContext2D,
    getWidth: () => 800,
    getHeight: () => 600,
    transform: { x: 0, y: 0, k: 1 },
    dragState: { dragging: false, dragStart: { x: 0, y: 0 } },
    dragDropState: createDragDropState(),
    simState: {
      simulation: { nodes: () => simNodes },
      simLinks,
    },
    hotkeysState: createHotkeysState(),
    noteFormState: createNoteFormState(),
    linkFormState: createLinkFormState(),
    zoomPanState: createZoomPanState(),
    getGhostNode: () => ghostNode,
    getBlackHole: () => blackHole,
    getGravitySystem: () => null,
    getHoveredNodeId: () => hoveredNodeId,
    setHoveredNodeId: vi.fn((value: string | null) => {
      hoveredNodeId = value;
    }),
    getHoveredLink: () => hoveredLink,
    setHoveredLink: vi.fn((value: any) => {
      hoveredLink = value;
    }),
    getTooltipPosition: () => tooltipPosition,
    setTooltipPosition: vi.fn((value: { x: number; y: number }) => {
      tooltipPosition = value;
    }),
    getFocusMode: () => focusMode,
    setFocusMode: vi.fn((value: boolean) => {
      focusMode = value;
    }),
    getSelectedNodeId: () => selectedNodeId,
    setSelectedNodeId: vi.fn((value: string | null) => {
      selectedNodeId = value;
    }),
    redraw: vi.fn(),
    toggleFocus: vi.fn(),
    openSearch: vi.fn(),
    closeSearch: vi.fn(),
    openHelp: vi.fn(),
    toggleHelp: vi.fn(),
    setGhostNode: vi.fn((node: any) => {
      ghostNode = node;
    }),
    onNodeClick: vi.fn(),
    onNoteDelete: vi.fn(),
    getKeyLines: () => ["tip1"],
  };
}

describe("event-bridge", () => {
  let canvas: HTMLCanvasElement;
  let context: ReturnType<typeof createMockContext>;
  let bridge: ReturnType<typeof createGraphEventBridge>;

  beforeEach(() => {
    vi.useFakeTimers();
    canvas = createCanvas();
    context = createMockContext(canvas);
    bridge = createGraphEventBridge(
      context as unknown as GraphCanvasEventContext,
    );
  });

  afterEach(() => {
    bridge.cleanup();
    vi.useRealTimers();
  });

  it("creates bridge with all handlers", () => {
    expect(bridge.onMouseDown).toBeTypeOf("function");
    expect(bridge.onMouseMove).toBeTypeOf("function");
    expect(bridge.onMouseUp).toBeTypeOf("function");
    expect(bridge.handleKeyDown).toBeTypeOf("function");
  });

  it("handles mouse down on a node and opens link form on drop", () => {
    bridge.onMouseDown(
      new MouseEvent("mousedown", { clientX: 1, clientY: 11 }),
    );
    expect(context.dragDropState.draggedNodeId).toBe("n1");

    bridge.onMouseUp(new MouseEvent("mouseup", { clientX: 50, clientY: 10 }));
    expect(context.linkFormState.showLinkForm).toBe(true);
    expect(context.linkFormState.linkSourceNodeId).toBe("n1");
    expect(context.linkFormState.linkTargetNodeId).toBe("n2");
  });

  it("handles ghost node click to open note form", () => {
    bridge.onMouseDown(
      new MouseEvent("mousedown", { clientX: 60, clientY: 60 }),
    );
    expect(context.noteFormState.showNoteForm).toBe(true);
  });

  it("pans canvas", () => {
    bridge.onMouseDown(
      new MouseEvent("mousedown", { clientX: 400, clientY: 300 }),
    );
    bridge.onMouseMove(
      new MouseEvent("mousemove", { clientX: 100, clientY: 50 }),
    );

    expect(context.transform.x).toBe(-300);
    expect(context.transform.y).toBe(-250);
  });

  it("drags a node and detects link target", () => {
    bridge.onMouseDown(
      new MouseEvent("mousedown", { clientX: 1, clientY: 11 }),
    );
    bridge.onMouseMove(
      new MouseEvent("mousemove", { clientX: 50, clientY: 10 }),
    );

    expect(context.dragDropState.isDraggingForLink).toBe(true);
    expect(context.dragDropState.linkTargetNodeId).toBe("n2");
  });

  it("schedules node hover", () => {
    bridge.onMouseMove(
      new MouseEvent("mousemove", { clientX: 50, clientY: 10 }),
    );
    expect(context.getHoveredNodeId()).toBeNull();

    vi.advanceTimersByTime(150);
    expect(context.getHoveredNodeId()).toBe("n2");
  });

  it("schedules link hover", () => {
    bridge.onMouseMove(
      new MouseEvent("mousemove", { clientX: 25, clientY: 12 }),
    );
    expect(context.getHoveredLink()).toBeNull();

    vi.advanceTimersByTime(150);
    expect(context.getHoveredLink()).not.toBeNull();
  });

  it("handles node click", () => {
    bridge.onClick(new MouseEvent("click", { clientX: 1, clientY: 11 }));
    expect(context.setSelectedNodeId).toHaveBeenCalledWith("n1");
    expect(context.onNodeClick).toHaveBeenCalledWith({
      id: "n1",
      title: "A",
      type: undefined,
    });
  });

  it("zooms on double click", () => {
    bridge.onDblClick(
      new MouseEvent("dblclick", { clientX: 400, clientY: 300 }),
    );
    expect(context.transform.k).toBeGreaterThan(1);
    expect(context.redraw).toHaveBeenCalled();
  });

  it("zooms on wheel", () => {
    bridge.onZoom(
      new WheelEvent("wheel", { deltaY: -100, clientX: 400, clientY: 300 }),
    );
    expect(context.transform.k).toBeGreaterThan(1);
  });

  it("handles touch double tap", () => {
    const touch = { clientX: 100, clientY: 100, identifier: 0 };
    bridge.onTouchStart({
      touches: [touch],
      preventDefault: vi.fn(),
    } as unknown as TouchEvent);

    vi.advanceTimersByTime(100);

    bridge.onTouchStart({
      touches: [touch],
      preventDefault: vi.fn(),
    } as unknown as TouchEvent);

    expect(context.transform.k).toBe(2);
  });

  it("handles keyboard shortcuts", () => {
    bridge.handleKeyDown(new KeyboardEvent("keydown", { key: "f" }));
    expect(context.openSearch).toHaveBeenCalled();

    bridge.handleKeyDown(new KeyboardEvent("keydown", { key: "?" }));
    expect(context.toggleHelp).toHaveBeenCalled();

    context.setSelectedNodeId("n1");
    bridge.handleKeyDown(new KeyboardEvent("keydown", { key: "Delete" }));
    expect(context.onNoteDelete).toHaveBeenCalledWith("n1");

    bridge.handleKeyDown(new KeyboardEvent("keydown", { key: "n" }));
    expect(context.noteFormState.showNoteForm).toBe(true);
  });

  it("cleanup cancels hover timeouts", () => {
    bridge.onMouseMove(
      new MouseEvent("mousemove", { clientX: 50, clientY: 10 }),
    );
    bridge.cleanup();
    vi.advanceTimersByTime(1000);
    expect(context.getHoveredNodeId()).toBeNull();
  });
});
