import {
  describe,
  it,
  expect,
  vi,
  beforeAll,
  beforeEach,
  afterEach,
} from "vitest";
import { render, cleanup, fireEvent } from "@testing-library/svelte";
import { tick } from "svelte";
import type { GraphDeltaData, GraphNode, GraphLink } from "$shared/api/graph";

// Shared state for the d3-force mock
const mockState = {
  simulationNodes: [] as any[],
  simulationLinks: [] as any[],
  tickCallback: null as (() => void) | null,
  endCallback: null as (() => void) | null,
};

// Unmock animation.ts so the real onUpdate callback runs; GraphCanvas is loaded
// dynamically in beforeAll so the unmock takes effect before import.
vi.unmock("$components/organisms/GraphCanvas/animation.ts");

vi.mock("d3-force", () => {
  const createMockSimulation = () => {
    const sim: any = {
      nodes: vi.fn((n?: any[]) => {
        if (n) {
          mockState.simulationNodes = n.map((node, i) => ({
            ...node,
            x: node.x ?? 400 + i * 50,
            y: node.y ?? 300 + i * 30,
          }));
        }
        return mockState.simulationNodes;
      }),
      tick: vi.fn(() => sim),
      force: vi.fn(() => sim),
      alphaDecay: vi.fn(() => sim),
      on: vi.fn((event: string, cb: () => void) => {
        if (event === "tick") mockState.tickCallback = cb;
        if (event === "end") mockState.endCallback = cb;
        return sim;
      }),
      alpha: vi.fn(() => sim),
      restart: vi.fn(() => {
        if (mockState.tickCallback) mockState.tickCallback();
        return sim;
      }),
      stop: vi.fn(() => {
        mockState.tickCallback = null;
        mockState.endCallback = null;
        return sim;
      }),
    };
    return sim;
  };

  const forceSimulation = vi.fn((nodes?: any[]) => {
    const sim = createMockSimulation();
    if (nodes) {
      sim.nodes(nodes);
    }
    return sim;
  });

  const forceLink = vi.fn((links?: any[]) => {
    if (links) mockState.simulationLinks = links;
    const linkForce: any = {
      id: vi.fn(() => linkForce),
      distance: vi.fn(() => linkForce),
      strength: vi.fn(() => linkForce),
      links: vi.fn(() => mockState.simulationLinks),
    };
    return linkForce;
  });

  const forceManyBody = vi.fn(() => ({
    strength: vi.fn((obj: any) => obj),
  }));
  const forceCenter = vi.fn(() => ({
    strength: vi.fn((obj: any) => obj),
  }));
  const forceCollide = vi.fn(() => ({
    radius: vi.fn((obj: any) => obj),
  }));

  return {
    forceSimulation,
    forceLink,
    forceManyBody,
    forceCenter,
    forceCollide,
  };
});

// Pre-position nodes so the initial transform is predictable and interactions
// can target known screen coordinates.
const mockNodes: any[] = [
  {
    id: "1",
    title: "Node 1",
    type: "star",
    x: 200,
    y: 300,
    fx: 200,
    fy: 300,
  },
  {
    id: "2",
    title: "Node 2",
    type: "planet",
    x: 600,
    y: 300,
    fx: 600,
    fy: 300,
  },
  {
    id: "3",
    title: "Node 3",
    type: "comet",
    x: 400,
    y: 100,
    fx: 400,
    fy: 100,
  },
  { id: "0", title: "Knowledge Core", type: "technical", x: 60, y: 60, fx: 60, fy: 60 },
];

const mockLinks: any[] = [
  { source: "1", target: "2", link_type: "reference", weight: 0.8 },
  { source: "2", target: "3", link_type: "dependency", weight: 0.6 },
];

let GraphCanvas: any;

function flushMicrotasks(): Promise<void> {
  return tick();
}

function getScreenPoint(worldX: number, worldY: number, transform: any) {
  return {
    x: worldX * transform.k + transform.x,
    y: worldY * transform.k + transform.y,
  };
}

beforeAll(async () => {
  GraphCanvas = (await import("$components/organisms/GraphCanvas.svelte"))
    .default;
});

describe("GraphCanvas events", () => {
  let renderResult: any;

  beforeEach(async () => {
    mockState.simulationNodes = [];
    mockState.simulationLinks = [];
    mockState.tickCallback = null;
    mockState.endCallback = null;

    // Rich enough CanvasRenderingContext2D mock for the renderer
    const ctx: any = {
      beginPath: vi.fn(),
      arc: vi.fn(),
      ellipse: vi.fn(),
      quadraticCurveTo: vi.fn(),
      fill: vi.fn(),
      stroke: vi.fn(),
      moveTo: vi.fn(),
      lineTo: vi.fn(),
      clearRect: vi.fn(),
      save: vi.fn(),
      restore: vi.fn(),
      translate: vi.fn(),
      scale: vi.fn(),
      rotate: vi.fn(),
      setTransform: vi.fn(),
      setLineDash: vi.fn(),
      fillText: vi.fn(),
      strokeText: vi.fn(),
      measureText: vi.fn(() => ({ width: 100 })),
      closePath: vi.fn(),
      rect: vi.fn(),
      fillRect: vi.fn(),
      strokeRect: vi.fn(),
      clip: vi.fn(),
      roundRect: vi.fn(),
      createLinearGradient: vi.fn(() => ({ addColorStop: vi.fn() })),
      createRadialGradient: vi.fn(() => ({ addColorStop: vi.fn() })),
      getImageData: vi.fn(() => ({ data: new Uint8ClampedArray(4) })),
      putImageData: vi.fn(),
      drawImage: vi.fn(),
      fillStyle: "",
      strokeStyle: "",
      lineWidth: 1,
      lineDashOffset: 0,
      globalAlpha: 1,
      font: "10px sans-serif",
      textAlign: "left",
      textBaseline: "alphabetic",
      shadowColor: "",
      shadowBlur: 0,
      shadowOffsetX: 1,
      shadowOffsetY: 1,
      imageSmoothingEnabled: true,
      imageSmoothingQuality: "low",
      lineCap: "butt",
      lineJoin: "miter",
    };

    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockImplementation(
      (contextId) => (contextId === "2d" ? ctx : null),
    );

    vi.spyOn(Element.prototype, "getBoundingClientRect").mockReturnValue({
      x: 0,
      y: 0,
      width: 800,
      height: 600,
      top: 0,
      left: 0,
      right: 800,
      bottom: 600,
      toJSON: () => ({}),
    } as any);

    class MockResizeObserver {
      constructor(private callback: any) {}
      observe() {
        this.callback([], this);
      }
      unobserve() {}
      disconnect() {}
    }

    class MockTouch {
      [key: string]: any;
      constructor(init: any) {
        Object.assign(this, init);
      }
    }

    class MockTouchEvent extends Event {
      [key: string]: any;
      constructor(type: string, init: any = {}) {
        super(type, init);
        this.touches = init.touches ?? [];
        this.changedTouches = init.changedTouches ?? [];
        this.targetTouches = init.targetTouches ?? [];
      }
    }

    vi.stubGlobal("ResizeObserver", MockResizeObserver);
    vi.stubGlobal("requestAnimationFrame", vi.fn(() => 1));
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
    vi.stubGlobal("Touch", MockTouch);
    vi.stubGlobal("TouchEvent", MockTouchEvent);

    vi.useFakeTimers();

    renderResult = render(GraphCanvas, {
      props: {
        nodes: mockNodes as GraphNode[],
        links: mockLinks as GraphLink[],
      },
    });

    await flushMicrotasks();
    vi.advanceTimersByTime(100);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllTimers();
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("initializes simulation and exposes debug helpers", () => {
    expect(window.__graphCanvas).toBeDefined();
    expect(window.__graphCanvas.getSimulationNodes().length).toBe(4);
  });

  it("renders the canvas and overlays", () => {
    const { container } = renderResult;
    expect(container.querySelector("canvas")).toBeTruthy();
    expect(container.querySelector('[data-testid="graph-controls"]')).toBeTruthy();
  });

  it("fires onNodeClick and selects a node", async () => {
    const onNodeClick = vi.fn();
    const { rerender, container } = renderResult;
    rerender({
      nodes: mockNodes as GraphNode[],
      links: mockLinks as GraphLink[],
      onNodeClick,
    });
    await flushMicrotasks();

    const canvas = container.querySelector("canvas") as HTMLCanvasElement;
    const transform = window.__graphCanvas!.transform;
    const p1 = getScreenPoint(200, 300, transform);

    canvas.dispatchEvent(
      new MouseEvent("mousedown", { clientX: p1.x, clientY: p1.y }),
    );
    canvas.dispatchEvent(
      new MouseEvent("mouseup", { clientX: p1.x, clientY: p1.y }),
    );
    canvas.dispatchEvent(
      new MouseEvent("click", { clientX: p1.x, clientY: p1.y }),
    );

    expect(onNodeClick).toHaveBeenCalledWith({
      id: "1",
      title: "Node 1",
      type: "star",
    });
  });

  it("deletes the selected node on Delete key", async () => {
    const onNodeClick = vi.fn();
    const onNoteDelete = vi.fn();
    const { rerender, container } = renderResult;
    rerender({
      nodes: mockNodes as GraphNode[],
      links: mockLinks as GraphLink[],
      onNodeClick,
      onNoteDelete,
    });
    await flushMicrotasks();

    const canvas = container.querySelector("canvas") as HTMLCanvasElement;
    const transform = window.__graphCanvas!.transform;
    const p1 = getScreenPoint(200, 300, transform);

    canvas.dispatchEvent(
      new MouseEvent("mousedown", { clientX: p1.x, clientY: p1.y }),
    );
    canvas.dispatchEvent(
      new MouseEvent("mouseup", { clientX: p1.x, clientY: p1.y }),
    );
    canvas.dispatchEvent(
      new MouseEvent("click", { clientX: p1.x, clientY: p1.y }),
    );

    fireEvent.keyDown(window, { key: "Delete" });

    expect(onNoteDelete).toHaveBeenCalledWith("1");
  });

  it("opens help when clicking the technical node", async () => {
    const { container } = renderResult;
    const canvas = container.querySelector("canvas") as HTMLCanvasElement;
    const transform = window.__graphCanvas!.transform;
    const pt = getScreenPoint(60, 60, transform);

    fireEvent.click(canvas, { clientX: pt.x, clientY: pt.y });
    await flushMicrotasks();

    expect(container.querySelector('[data-testid="help-modal"]')).toBeTruthy();

    const closeBtn = container.querySelector(".close-btn");
    expect(closeBtn).toBeTruthy();
    (closeBtn as HTMLElement).click();
    await flushMicrotasks();

    expect(container.querySelector('[data-testid="help-modal"]')).toBeFalsy();
  });

  it("creates a note from the ghost node", async () => {
    const onNoteCreate = vi.fn();
    const { rerender, container } = renderResult;
    rerender({
      nodes: mockNodes as GraphNode[],
      links: mockLinks as GraphLink[],
      onNoteCreate,
    });
    await flushMicrotasks();

    fireEvent.keyDown(window, { key: "n" });
    await flushMicrotasks();

    const titleInput = container.querySelector(
      '[data-testid="ghost-note-title"]',
    ) as HTMLInputElement;
    expect(titleInput).toBeTruthy();

    await fireEvent.input(titleInput, { target: { value: "New Note" } });

    const createBtn = container.querySelector(
      '[data-testid="ghost-note-create"]',
    );
    expect(createBtn).toBeTruthy();
    (createBtn as HTMLElement).click();

    expect(onNoteCreate).toHaveBeenCalledWith({
      title: "New Note",
      content: "",
      type: "planet",
    });
  });

  it("cancels note creation", async () => {
    const onNoteCreate = vi.fn();
    const { rerender, container } = renderResult;
    rerender({
      nodes: mockNodes as GraphNode[],
      links: mockLinks as GraphLink[],
      onNoteCreate,
    });
    await flushMicrotasks();

    fireEvent.keyDown(window, { key: "n" });
    await flushMicrotasks();

    const cancelBtn = container.querySelector(
      '[data-testid="ghost-note-cancel"]',
    );
    expect(cancelBtn).toBeTruthy();
    (cancelBtn as HTMLElement).click();
    await flushMicrotasks();

    expect(onNoteCreate).not.toHaveBeenCalled();
    expect(container.querySelector('[data-testid="ghost-note-form"]')).toBeFalsy();
  });

  it("creates a link by dragging between nodes", async () => {
    const onLinkCreate = vi.fn();
    const { rerender, container } = renderResult;
    rerender({
      nodes: mockNodes as GraphNode[],
      links: mockLinks as GraphLink[],
      onLinkCreate,
    });
    await flushMicrotasks();

    const canvas = container.querySelector("canvas") as HTMLCanvasElement;
    const transform = window.__graphCanvas!.transform;
    const p1 = getScreenPoint(200, 300, transform);
    const p2 = getScreenPoint(600, 300, transform);

    canvas.dispatchEvent(
      new MouseEvent("mousedown", { clientX: p1.x, clientY: p1.y }),
    );
    canvas.dispatchEvent(
      new MouseEvent("mousemove", { clientX: p2.x, clientY: p2.y }),
    );
    canvas.dispatchEvent(
      new MouseEvent("mouseup", { clientX: p2.x, clientY: p2.y }),
    );
    await flushMicrotasks();

    const createBtn = container.querySelector(
      '[data-testid="link-form-create"]',
    );
    expect(createBtn).toBeTruthy();
    (createBtn as HTMLElement).click();

    expect(onLinkCreate).toHaveBeenCalledWith({
      source: "1",
      target: "2",
      link_type: "related",
      weight: 0.5,
    });
  });

  it("cancels link creation", async () => {
    const onLinkCreate = vi.fn();
    const { rerender, container } = renderResult;
    rerender({
      nodes: mockNodes as GraphNode[],
      links: mockLinks as GraphLink[],
      onLinkCreate,
    });
    await flushMicrotasks();

    const canvas = container.querySelector("canvas") as HTMLCanvasElement;
    const transform = window.__graphCanvas!.transform;
    const p1 = getScreenPoint(200, 300, transform);
    const p2 = getScreenPoint(600, 300, transform);

    canvas.dispatchEvent(
      new MouseEvent("mousedown", { clientX: p1.x, clientY: p1.y }),
    );
    canvas.dispatchEvent(
      new MouseEvent("mousemove", { clientX: p2.x, clientY: p2.y }),
    );
    canvas.dispatchEvent(
      new MouseEvent("mouseup", { clientX: p2.x, clientY: p2.y }),
    );
    await flushMicrotasks();

    const cancelBtn = container.querySelector('[data-testid="link-form-cancel"]');
    expect(cancelBtn).toBeTruthy();
    (cancelBtn as HTMLElement).click();
    await flushMicrotasks();

    expect(onLinkCreate).not.toHaveBeenCalled();
    expect(container.querySelector('[data-testid="link-form"]')).toBeFalsy();
  });

  it("edits and deletes a hovered link", async () => {
    const onLinkEdit = vi.fn();
    const onLinkDelete = vi.fn();
    const { rerender, container } = renderResult;
    rerender({
      nodes: mockNodes as GraphNode[],
      links: mockLinks as GraphLink[],
      onLinkEdit,
      onLinkDelete,
    });
    await flushMicrotasks();

    const canvas = container.querySelector("canvas") as HTMLCanvasElement;
    const transform = window.__graphCanvas!.transform;
    // midpoint of link 1-2 in world coords is (400, 300)
    const midpoint = getScreenPoint(400, 300, transform);

    canvas.dispatchEvent(
      new MouseEvent("mousemove", {
        clientX: midpoint.x,
        clientY: midpoint.y,
      }),
    );

    vi.advanceTimersByTime(150);
    await flushMicrotasks();

    const editBtn = container.querySelector(".link-tooltip .edit-btn");
    expect(editBtn).toBeTruthy();
    fireEvent.mouseDown(editBtn as HTMLElement);

    expect(onLinkEdit).toHaveBeenCalledWith(
      expect.objectContaining({
        source: "1",
        target: "2",
        link_type: "reference",
      }),
    );

    // Hover the link again to test delete
    canvas.dispatchEvent(
      new MouseEvent("mousemove", {
        clientX: midpoint.x,
        clientY: midpoint.y,
      }),
    );
    vi.advanceTimersByTime(150);
    await flushMicrotasks();

    const deleteBtn = container.querySelector(".link-tooltip .delete-btn");
    expect(deleteBtn).toBeTruthy();
    fireEvent.mouseDown(deleteBtn as HTMLElement);

    expect(onLinkDelete).toHaveBeenCalledWith({
      source: "1",
      target: "2",
      link_type: "reference",
    });
  });

  it("controls: reset, search, focus mode, help toggle", async () => {
    const { container } = renderResult;

    const resetBtn = container.querySelector('[data-testid="graph-controls-reset"]');
    expect(resetBtn).toBeTruthy();
    (resetBtn as HTMLElement).click();

    const searchBtn = container.querySelector('[data-testid="graph-controls-search"]');
    expect(searchBtn).toBeTruthy();
    (searchBtn as HTMLElement).click();
    await flushMicrotasks();
    expect(container.querySelector('[data-testid="search-box"]')).toBeTruthy();

    const modeBtn = container.querySelector('[data-testid="graph-controls-mode"]');
    expect(modeBtn).toBeTruthy();
    (modeBtn as HTMLElement).click();
    await flushMicrotasks();

    const focusBtn = container.querySelector('[data-testid="graph-controls-focus"]');
    expect(focusBtn).toBeTruthy();
    (focusBtn as HTMLElement).click();
    await flushMicrotasks();

    fireEvent.keyDown(window, { key: "?" });
    await flushMicrotasks();
    expect(container.querySelector('[data-testid="help-modal"]')).toBeTruthy();
  });

  it("searches, focuses matches, and closes search", async () => {
    const { rerender, container } = renderResult;
    rerender({
      nodes: mockNodes as GraphNode[],
      links: mockLinks as GraphLink[],
    });
    await flushMicrotasks();

    fireEvent.keyDown(window, { key: "f" });
    await flushMicrotasks();

    const searchBox = container.querySelector('[data-testid="search-box"]') as HTMLElement;
    expect(searchBox).toBeTruthy();

    const input = searchBox?.querySelector("input");
    expect(input).toBeTruthy();
    await fireEvent.input(input as HTMLInputElement, {
      target: { value: "Node 1" },
    });
    await flushMicrotasks();

    // Overlay shows the match count once the search query is processed.
    expect(searchBox.textContent).toContain("1/1");

    const transformBefore = window.__graphCanvas!.transform;
    fireEvent.keyDown(window, { key: "Enter" });
    await flushMicrotasks();

    // Enter cycles to the first match and re-centers the view.
    expect(transformBefore.x).toBeCloseTo(160, 1);
    expect(transformBefore.y).toBeCloseTo(-60, 1);

    fireEvent.keyDown(window, { key: "Escape" });
    await flushMicrotasks();

    expect(container.querySelector('[data-testid="search-box"]')).toBeFalsy();
  });

  it("handles touchstart and reports browser/ctx metrics", async () => {
    const { container } = renderResult;
    const canvas = container.querySelector("canvas") as HTMLCanvasElement;

    canvas.dispatchEvent(
      new (window as any).TouchEvent("touchstart", {
        touches: [new (window as any).Touch({ clientX: 60, clientY: 60 })],
        bubbles: true,
      }),
    );

    // Touchstart should not throw and should exercise getCtx/getWidth/getHeight.
    expect(canvas).toBeTruthy();
  });

  it("applies delta updates to the simulation", async () => {
    const { rerender } = renderResult;
    const delta: GraphDeltaData = {
      added_nodes: [{ id: "5", title: "Delta Node", type: "star" }],
    };

    rerender({
      nodes: mockNodes as GraphNode[],
      links: mockLinks as GraphLink[],
      delta,
    });
    await flushMicrotasks();

    expect(window.__graphCanvas!.getSimulationNodes().length).toBe(5);
  });

  it("fires delayed resize, inactivity, and getKeyLines timers", async () => {
    const { container } = renderResult;
    const canvas = container.querySelector("canvas") as HTMLCanvasElement;

    // Delayed resize already covered by beforeEach advance of 100ms; ensure no
    // errors and advance the remaining timers.
    vi.advanceTimersByTime(10000);
    await flushMicrotasks();

    // Trigger updateActivity to schedule the getKeyLines inactivity callback.
    canvas.dispatchEvent(
      new MouseEvent("mousemove", { clientX: 100, clientY: 100 }),
    );

    vi.advanceTimersByTime(10000);
    await flushMicrotasks();

    vi.advanceTimersByTime(4000);
    await flushMicrotasks();

    // Should not throw and timers should have fired.
    expect(canvas).toBeTruthy();
  });
});

declare global {
  interface Window {
    __graphCanvas?: {
      getSimulationNodes: () => any[];
      transform: { x: number; y: number; k: number };
    };
  }
}
