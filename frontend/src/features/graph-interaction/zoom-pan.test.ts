import { describe, it, expect, vi } from "vitest";
import { createZoomPanState, handleZoom, handleTouchStart, resetViewToCenter } from "./zoom-pan";

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

function createTouchEvent(touches: { clientX: number; clientY: number }[]) {
  return {
    touches,
    preventDefault: vi.fn(),
  } as unknown as TouchEvent;
}

describe("zoom-pan", () => {
  it("creates zoom/pan state", () => {
    const state = createZoomPanState();
    expect(state.lastTouchTime).toBe(0);
    expect(state.tapCount).toBe(0);
  });

  it("zooms in with negative deltaY", () => {
    const canvas = createCanvas();
    const transform = { x: 0, y: 0, k: 1 };
    const redraw = vi.fn();
    const event = new WheelEvent("wheel", {
      deltaY: -100,
      clientX: 400,
      clientY: 300,
    });

    handleZoom(event, transform, canvas, redraw);

    expect(transform.k).toBeGreaterThan(1);
    expect(redraw).toHaveBeenCalled();
  });

  it("zooms out with positive deltaY", () => {
    const canvas = createCanvas();
    const transform = { x: 0, y: 0, k: 1 };
    const redraw = vi.fn();
    const event = new WheelEvent("wheel", {
      deltaY: 100,
      clientX: 400,
      clientY: 300,
    });

    handleZoom(event, transform, canvas, redraw);

    expect(transform.k).toBeLessThan(1);
    expect(redraw).toHaveBeenCalled();
  });

  it("handles double tap zoom", () => {
    const canvas = createCanvas();
    const transform = { x: 0, y: 0, k: 1 };
    const simNodes = [{ id: "n1", x: 100, y: 100, title: "A" }];
    const state = createZoomPanState();
    const ctx = {} as CanvasRenderingContext2D;

    const touch1 = createTouchEvent([{ clientX: 100, clientY: 100 }]);
    handleTouchStart(touch1, state, transform, canvas, simNodes, ctx, 800, 600);

    const touch2 = createTouchEvent([{ clientX: 100, clientY: 100 }]);
    handleTouchStart(touch2, state, transform, canvas, simNodes, ctx, 800, 600);

    expect(transform.k).toBe(2);
  });

  it("resets view to center around nodes", () => {
    const transform = { x: 0, y: 0, k: 1 };
    const simNodes = [
      { id: "n1", x: 0, y: 0 },
      { id: "n2", x: 100, y: 100 },
    ];

    resetViewToCenter(transform, 800, 600, simNodes as any);

    expect(transform.k).toBeGreaterThan(0);
    expect(typeof transform.x).toBe("number");
    expect(typeof transform.y).toBe("number");
  });

  it("does nothing when resetting view with no nodes", () => {
    const transform = { x: 0, y: 0, k: 1 };
    resetViewToCenter(transform, 800, 600, []);
    expect(transform.k).toBe(1);
    expect(transform.x).toBe(0);
    expect(transform.y).toBe(0);
  });

  it("ignores nodes without x/y", () => {
    const transform = { x: 0, y: 0, k: 1 };
    const simNodes = [{ id: "n1" }, { id: "n2" }];

    resetViewToCenter(transform, 800, 600, simNodes as any);

    // All nodes are skipped, so bounding box is empty and transform is not useful.
    expect(Number.isNaN(transform.x)).toBe(true);
    expect(Number.isNaN(transform.y)).toBe(true);
  });

  it("resets view on second double-tap", () => {
    const canvas = createCanvas();
    const transform = { x: 0, y: 0, k: 2 };
    const simNodes = [{ id: "n1", x: 100, y: 100, title: "A" }];
    const state = createZoomPanState();
    const ctx = {} as CanvasRenderingContext2D;

    const touch1 = createTouchEvent([{ clientX: 100, clientY: 100 }]);
    handleTouchStart(touch1, state, transform, canvas, simNodes, ctx, 800, 600);

    const touch2 = createTouchEvent([{ clientX: 100, clientY: 100 }]);
    handleTouchStart(touch2, state, transform, canvas, simNodes, ctx, 800, 600);

    const touch3 = createTouchEvent([{ clientX: 100, clientY: 100 }]);
    handleTouchStart(touch3, state, transform, canvas, simNodes, ctx, 800, 600);

    expect(transform.k).toBe(1);
    expect(state.tapCount).toBe(0);
  });

  it("ignores multi-touch start", () => {
    const canvas = createCanvas();
    const transform = { x: 0, y: 0, k: 1 };
    const simNodes = [{ id: "n1", x: 100, y: 100, title: "A" }];
    const state = createZoomPanState();
    const ctx = {} as CanvasRenderingContext2D;

    const multiTouch = createTouchEvent([
      { clientX: 100, clientY: 100 },
      { clientX: 200, clientY: 200 },
    ]);
    handleTouchStart(multiTouch, state, transform, canvas, simNodes, ctx, 800, 600);

    expect(state.lastTouchTime).toBe(0);
    expect(state.tapCount).toBe(0);
  });

  it("ignores double tap when distance is too large", () => {
    const canvas = createCanvas();
    const transform = { x: 0, y: 0, k: 1 };
    const simNodes = [{ id: "n1", x: 100, y: 100, title: "A" }];
    const state = createZoomPanState();
    const ctx = {} as CanvasRenderingContext2D;

    const touch1 = createTouchEvent([{ clientX: 100, clientY: 100 }]);
    handleTouchStart(touch1, state, transform, canvas, simNodes, ctx, 800, 600);

    const touch2 = createTouchEvent([{ clientX: 200, clientY: 200 }]);
    handleTouchStart(touch2, state, transform, canvas, simNodes, ctx, 800, 600);

    expect(transform.k).toBe(1);
    expect(state.tapCount).toBe(0);
  });
});
