import { describe, it, expect, vi, beforeEach } from "vitest";
import { goto } from "$app/navigation";
import {
  findLinkAtPosition,
  handleZoom,
  handlePanStart,
  handlePanMove,
  handlePanEnd,
  handleClick,
} from "./interactions";
import type { SimulationNode, SimulationLink, TransformState, DragState } from "./types";

const makeCanvas = () => {
  const canvas = document.createElement("canvas");
  canvas.width = 800;
  canvas.height = 600;
  vi.spyOn(canvas, "getBoundingClientRect").mockReturnValue({
    left: 0,
    top: 0,
    width: 800,
    height: 600,
    right: 800,
    bottom: 600,
    x: 0,
    y: 0,
    toJSON: () => {},
  });
  return canvas;
};

const makeNode = (id: string, x: number, y: number): SimulationNode => ({
  id,
  title: id,
  x,
  y,
});

describe("GraphCanvas interactions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("findLinkAtPosition", () => {
    const nodes: SimulationNode[] = [
      makeNode("a", 0, 0),
      makeNode("b", 100, 0),
      makeNode("c", 0, 100),
    ];

    it("finds a link near the line segment", () => {
      const link: SimulationLink = { source: "a", target: "b" };
      const result = findLinkAtPosition(50, 5, [link], nodes, {
        x: 0,
        y: 0,
        k: 1,
      });
      expect(result).toBe(link);
    });

    it("returns null when far from any link", () => {
      const link: SimulationLink = { source: "a", target: "b" };
      const result = findLinkAtPosition(50, 50, [link], nodes, {
        x: 0,
        y: 0,
        k: 1,
      });
      expect(result).toBeNull();
    });

    it("supports source/target as node objects", () => {
      const link: SimulationLink = { source: nodes[0], target: nodes[1] };
      const result = findLinkAtPosition(50, 2, [link], nodes, {
        x: 0,
        y: 0,
        k: 1,
      });
      expect(result).toBe(link);
    });

    it("supports source/target as numeric indices", () => {
      const link: SimulationLink = { source: 0, target: 1 };
      const result = findLinkAtPosition(50, 3, [link], nodes, {
        x: 0,
        y: 0,
        k: 1,
      });
      expect(result).toBe(link);
    });

    it("ignores links with missing endpoints", () => {
      const link: SimulationLink = { source: "missing", target: "a" };
      const result = findLinkAtPosition(50, 3, [link], nodes, {
        x: 0,
        y: 0,
        k: 1,
      });
      expect(result).toBeNull();
    });

    it("ignores links with undefined coordinates", () => {
      const noCoords = { id: "d", title: "d" } as SimulationNode;
      const link: SimulationLink = { source: noCoords, target: nodes[0] };
      const result = findLinkAtPosition(0, 0, [link], nodes, {
        x: 0,
        y: 0,
        k: 1,
      });
      expect(result).toBeNull();
    });

    it("respects custom tolerance", () => {
      const link: SimulationLink = { source: "a", target: "b" };
      const result = findLinkAtPosition(50, 20, [link], nodes, { x: 0, y: 0, k: 1 }, 25);
      expect(result).toBe(link);
    });
  });

  describe("handleZoom", () => {
    it("zooms in with negative deltaY", () => {
      const transform: TransformState = { x: 0, y: 0, k: 1 };
      const canvas = makeCanvas();
      const onDraw = vi.fn();
      const event = new WheelEvent("wheel", {
        deltaY: -100,
        clientX: 100,
        clientY: 100,
      });

      handleZoom(event, transform, canvas, onDraw);

      expect(transform.k).toBeGreaterThan(1);
      expect(onDraw).toHaveBeenCalled();
    });

    it("zooms out with positive deltaY", () => {
      const transform: TransformState = { x: 0, y: 0, k: 1 };
      const canvas = makeCanvas();
      const onDraw = vi.fn();
      const event = new WheelEvent("wheel", {
        deltaY: 100,
        clientX: 100,
        clientY: 100,
      });

      handleZoom(event, transform, canvas, onDraw);

      expect(transform.k).toBeLessThan(1);
      expect(onDraw).toHaveBeenCalled();
    });

    it("clamps zoom to minimum", () => {
      const transform: TransformState = { x: 0, y: 0, k: 0.21 };
      const canvas = makeCanvas();
      const onDraw = vi.fn();
      const event = new WheelEvent("wheel", {
        deltaY: 100,
        clientX: 100,
        clientY: 100,
      });

      handleZoom(event, transform, canvas, onDraw);

      expect(transform.k).toBe(0.21);
      expect(onDraw).not.toHaveBeenCalled();
    });

    it("clamps zoom to maximum", () => {
      const transform: TransformState = { x: 0, y: 0, k: 4.9 };
      const canvas = makeCanvas();
      const onDraw = vi.fn();
      const event = new WheelEvent("wheel", {
        deltaY: -100,
        clientX: 100,
        clientY: 100,
      });

      handleZoom(event, transform, canvas, onDraw);

      expect(transform.k).toBe(4.9);
      expect(onDraw).not.toHaveBeenCalled();
    });
  });

  describe("handlePanStart/Move/End", () => {
    it("starts pan and sets cursor", () => {
      const canvas = makeCanvas();
      const dragState: DragState = {
        dragging: false,
        dragStart: { x: 0, y: 0 },
      };
      const transform: TransformState = { x: 10, y: 20, k: 1 };
      const event = new MouseEvent("mousedown", { clientX: 100, clientY: 200 });

      handlePanStart(event, dragState, transform, canvas);

      expect(dragState.dragging).toBe(true);
      expect(dragState.dragStart).toEqual({ x: 90, y: 180 });
      expect(canvas.style.cursor).toBe("grabbing");
    });

    it("moves pan when dragging", () => {
      const dragState: DragState = {
        dragging: true,
        dragStart: { x: 90, y: 180 },
      };
      const transform: TransformState = { x: 10, y: 20, k: 1 };
      const onDraw = vi.fn();
      const event = new MouseEvent("mousemove", { clientX: 150, clientY: 250 });

      handlePanMove(event, dragState, transform, onDraw);

      expect(transform.x).toBe(60);
      expect(transform.y).toBe(70);
      expect(onDraw).toHaveBeenCalled();
    });

    it("ignores pan move when not dragging", () => {
      const dragState: DragState = {
        dragging: false,
        dragStart: { x: 0, y: 0 },
      };
      const transform: TransformState = { x: 10, y: 20, k: 1 };
      const onDraw = vi.fn();
      const event = new MouseEvent("mousemove", { clientX: 150, clientY: 250 });

      handlePanMove(event, dragState, transform, onDraw);

      expect(transform.x).toBe(10);
      expect(transform.y).toBe(20);
      expect(onDraw).not.toHaveBeenCalled();
    });

    it("ends pan and resets cursor", () => {
      const canvas = makeCanvas();
      const dragState: DragState = {
        dragging: true,
        dragStart: { x: 0, y: 0 },
      };

      handlePanEnd(dragState, canvas);

      expect(dragState.dragging).toBe(false);
      expect(canvas.style.cursor).toBe("grab");
    });
  });

  describe("handleClick", () => {
    it("calls onNodeClick when a node is clicked", () => {
      const canvas = makeCanvas();
      const transform: TransformState = { x: 0, y: 0, k: 1 };
      const nodes: SimulationNode[] = [makeNode("n1", 100, 100)];
      const onNodeClick = vi.fn();
      const event = new MouseEvent("click", { clientX: 105, clientY: 105 });

      handleClick(event, canvas, transform, nodes, onNodeClick);

      expect(onNodeClick).toHaveBeenCalledWith({
        id: "n1",
        title: "n1",
        type: undefined,
      });
      expect(goto).not.toHaveBeenCalled();
    });

    it("navigates to note when no onNodeClick provided", () => {
      const canvas = makeCanvas();
      const transform: TransformState = { x: 0, y: 0, k: 1 };
      const nodes: SimulationNode[] = [{ id: "n2", title: "n2", x: 200, y: 200 }];
      const event = new MouseEvent("click", { clientX: 205, clientY: 205 });

      handleClick(event, canvas, transform, nodes);

      expect(goto).toHaveBeenCalledWith("/notes/n2");
    });

    it("does nothing when click is far from nodes", () => {
      const canvas = makeCanvas();
      const transform: TransformState = { x: 0, y: 0, k: 1 };
      const nodes: SimulationNode[] = [makeNode("n3", 100, 100)];
      const onNodeClick = vi.fn();
      const event = new MouseEvent("click", { clientX: 500, clientY: 500 });

      handleClick(event, canvas, transform, nodes, onNodeClick);

      expect(onNodeClick).not.toHaveBeenCalled();
      expect(goto).not.toHaveBeenCalled();
    });
  });
});
