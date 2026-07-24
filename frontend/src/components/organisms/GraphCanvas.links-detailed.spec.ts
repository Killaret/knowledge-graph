import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import * as renderer from "./GraphCanvas/renderer";
import type { SimulationNode, SimulationLink } from "./GraphCanvas/types";

describe("GraphCanvas - Link Rendering Detailed", () => {
  let ctx: CanvasRenderingContext2D;
  let ctxCalls: {
    method: string;
    args: any[];
    fillStyle?: string;
    strokeStyle?: string;
    lineWidth?: number;
  }[];

  beforeEach(() => {
    ctxCalls = [];
    let lastFillStyle = "";
    let lastStrokeStyle = "";
    let lastLineWidth = 1;

    ctx = {
      beginPath: vi.fn(() =>
        ctxCalls.push({
          method: "beginPath",
          args: [],
          fillStyle: lastFillStyle,
        })
      ),
      moveTo: vi.fn((x, y) => ctxCalls.push({ method: "moveTo", args: [x, y] })),
      lineTo: vi.fn((x, y) => ctxCalls.push({ method: "lineTo", args: [x, y] })),
      stroke: vi.fn(() =>
        ctxCalls.push({
          method: "stroke",
          args: [],
          strokeStyle: lastStrokeStyle,
          lineWidth: lastLineWidth,
        })
      ),
      fill: vi.fn(() => ctxCalls.push({ method: "fill", args: [], fillStyle: lastFillStyle })),
      closePath: vi.fn(() => ctxCalls.push({ method: "closePath", args: [] })),
      arc: vi.fn((x, y, r, s, e) => ctxCalls.push({ method: "arc", args: [x, y, r, s, e] })),
      ellipse: vi.fn((...args) => ctxCalls.push({ method: "ellipse", args })),
      setLineDash: vi.fn((dash) => ctxCalls.push({ method: "setLineDash", args: [dash] })),
      fillText: vi.fn((text, x, y) => ctxCalls.push({ method: "fillText", args: [text, x, y] })),
      measureText: vi.fn(() => ({ width: 50 })),
      save: vi.fn(() => ctxCalls.push({ method: "save", args: [] })),
      restore: vi.fn(() => ctxCalls.push({ method: "restore", args: [] })),
      translate: vi.fn((...args) => ctxCalls.push({ method: "translate", args })),
      scale: vi.fn((...args) => ctxCalls.push({ method: "scale", args })),
      rotate: vi.fn((...args) => ctxCalls.push({ method: "rotate", args })),
      set fillStyle(value: string) {
        lastFillStyle = value;
      },
      get fillStyle() {
        return lastFillStyle;
      },
      set strokeStyle(value: string) {
        lastStrokeStyle = value;
      },
      get strokeStyle() {
        return lastStrokeStyle;
      },
      set lineWidth(value: number) {
        lastLineWidth = value;
      },
      get lineWidth() {
        return lastLineWidth;
      },
      set font(value: string) {},
      get font() {
        return "14px sans-serif";
      },
      set textAlign(value: string) {},
      get textAlign() {
        return "center";
      },
      set textBaseline(value: string) {},
      get textBaseline() {
        return "middle";
      },
      set shadowBlur(value: number) {},
      get shadowBlur() {
        return 0;
      },
      set shadowColor(value: string) {},
      get shadowColor() {
        return "";
      },
    } as unknown as CanvasRenderingContext2D;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("Link Coordinates", () => {
    it("draws link from source to target with correct coordinates", () => {
      const sourceNode: SimulationNode = {
        id: "1",
        x: 100,
        y: 100,
        title: "Source",
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "2",
        x: 300,
        y: 200,
        title: "Target",
        type: "planet",
      };
      const link: SimulationLink = {
        source: "1",
        target: "2",
        weight: 0.8,
        link_type: "reference",
      };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      // Check moveTo is called with source coordinates
      const moveToCall = ctxCalls.find((c) => c.method === "moveTo");
      expect(moveToCall).toBeDefined();
      expect(moveToCall?.args).toEqual([100, 100]);

      // Check lineTo is called with target coordinates
      const lineToCall = ctxCalls.find((c) => c.method === "lineTo");
      expect(lineToCall).toBeDefined();
      expect(lineToCall?.args).toEqual([300, 200]);
    });

    it("draws multiple links with correct source-target pairs", () => {
      const nodes: SimulationNode[] = [
        { id: "1", x: 100, y: 100, title: "A", type: "star" },
        { id: "2", x: 200, y: 150, title: "B", type: "planet" },
        { id: "3", x: 300, y: 100, title: "C", type: "comet" },
      ];

      const links: SimulationLink[] = [
        { source: "1", target: "2", weight: 0.8 },
        { source: "2", target: "3", weight: 0.6 },
      ];

      // Draw each link
      links.forEach((link) => {
        const sourceNode = nodes.find((n) => n.id === link.source)!;
        const targetNode = nodes.find((n) => n.id === link.target)!;
        renderer.drawLink(ctx, link, sourceNode, targetNode);
      });

      // Check we have 2 moveTo calls (one per link)
      const moveToCalls = ctxCalls.filter((c) => c.method === "moveTo");
      expect(moveToCalls.length).toBe(2);

      // First link: from node 1 (100, 100)
      expect(moveToCalls[0].args).toEqual([100, 100]);

      // Second link: from node 2 (200, 150)
      expect(moveToCalls[1].args).toEqual([200, 150]);
    });
  });

  describe("Link Type Styles", () => {
    it("reference link uses solid blue line (#3366ff)", () => {
      const sourceNode: SimulationNode = {
        id: "1",
        x: 100,
        y: 100,
        title: "Source",
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "2",
        x: 300,
        y: 200,
        title: "Target",
        type: "planet",
      };
      const link: SimulationLink = {
        source: "1",
        target: "2",
        weight: 0.8,
        link_type: "reference",
      };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      const strokeCall = ctxCalls.find((c) => c.method === "stroke");
      expect(strokeCall).toBeDefined();
      expect(strokeCall?.strokeStyle).toContain("51, 102, 255"); // rgba(51, 102, 255, opacity) - blue

      // Check dash is empty (solid line)
      const dashCall = ctxCalls.find((c) => c.method === "setLineDash");
      if (dashCall) {
        expect(dashCall.args[0]).toEqual([]);
      }
    });

    it("dependency link uses dashed orange line (#ff6600)", () => {
      const sourceNode: SimulationNode = {
        id: "1",
        x: 100,
        y: 100,
        title: "Source",
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "2",
        x: 300,
        y: 200,
        title: "Target",
        type: "planet",
      };
      const link: SimulationLink = {
        source: "1",
        target: "2",
        weight: 0.8,
        link_type: "dependency",
      };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      const strokeCall = ctxCalls.find((c) => c.method === "stroke");
      expect(strokeCall).toBeDefined();
      expect(strokeCall?.strokeStyle).toContain("255, 102, 0"); // rgba(255, 102, 0, opacity) - orange

      // Check for dash pattern
      const dashCall = ctxCalls.find((c) => c.method === "setLineDash");
      expect(dashCall).toBeDefined();
      expect(dashCall?.args[0]).toEqual([10, 3]);
    });

    it("related link with strong weight uses solid gray line", () => {
      const sourceNode: SimulationNode = {
        id: "1",
        x: 100,
        y: 100,
        title: "Source",
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "2",
        x: 300,
        y: 200,
        title: "Target",
        type: "planet",
      };
      const link: SimulationLink = {
        source: "1",
        target: "2",
        weight: 0.8,
        link_type: "related",
      };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      const strokeCall = ctxCalls.find((c) => c.method === "stroke");
      expect(strokeCall).toBeDefined();
      expect(strokeCall?.strokeStyle).toContain("153, 153, 153"); // gray
    });

    it("related link with weak weight uses dashed gray line", () => {
      const sourceNode: SimulationNode = {
        id: "1",
        x: 100,
        y: 100,
        title: "Source",
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "2",
        x: 300,
        y: 200,
        title: "Target",
        type: "planet",
      };
      const link: SimulationLink = {
        source: "1",
        target: "2",
        weight: 0.2,
        link_type: "related",
      };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      const strokeCall = ctxCalls.find((c) => c.method === "stroke");
      expect(strokeCall).toBeDefined();

      // Check for dash pattern (weak weight)
      const dashCall = ctxCalls.find((c) => c.method === "setLineDash");
      expect(dashCall).toBeDefined();
      expect(dashCall?.args[0]).toEqual([6, 4]);
    });

    it("custom link uses dotted pink line (#ff66ff)", () => {
      const sourceNode: SimulationNode = {
        id: "1",
        x: 100,
        y: 100,
        title: "Source",
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "2",
        x: 300,
        y: 200,
        title: "Target",
        type: "planet",
      };
      const link: SimulationLink = {
        source: "1",
        target: "2",
        weight: 0.8,
        link_type: "custom",
      };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      const strokeCall = ctxCalls.find((c) => c.method === "stroke");
      expect(strokeCall).toBeDefined();
      expect(strokeCall?.strokeStyle).toContain("255, 102, 255"); // pink

      // Check for dotted pattern
      const dashCall = ctxCalls.find((c) => c.method === "setLineDash");
      expect(dashCall).toBeDefined();
      expect(dashCall?.args[0]).toEqual([2, 6]);
    });
  });

  describe("Link Line Width", () => {
    it("line width increases with weight", () => {
      const sourceNode: SimulationNode = {
        id: "1",
        x: 100,
        y: 100,
        title: "Source",
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "2",
        x: 300,
        y: 200,
        title: "Target",
        type: "planet",
      };

      // High weight link
      const strongLink: SimulationLink = {
        source: "1",
        target: "2",
        weight: 1.0,
        link_type: "reference",
      };
      renderer.drawLink(ctx, strongLink, sourceNode, targetNode);
      const strongStroke = ctxCalls.find((c) => c.method === "stroke");
      expect(strongStroke?.lineWidth).toBeGreaterThan(2);

      // Reset calls
      ctxCalls = [];

      // Low weight link
      const weakLink: SimulationLink = {
        source: "1",
        target: "2",
        weight: 0.3,
        link_type: "reference",
      };
      renderer.drawLink(ctx, weakLink, sourceNode, targetNode);
      const weakStroke = ctxCalls.find((c) => c.method === "stroke");
      expect(weakStroke?.lineWidth).toBeLessThan(strongStroke?.lineWidth || 3);
    });

    it("line width is consistent for same weight across types", () => {
      const sourceNode: SimulationNode = {
        id: "1",
        x: 100,
        y: 100,
        title: "Source",
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "2",
        x: 300,
        y: 200,
        title: "Target",
        type: "planet",
      };

      // Dependency link
      const depLink: SimulationLink = {
        source: "1",
        target: "2",
        weight: 0.8,
        link_type: "dependency",
      };
      renderer.drawLink(ctx, depLink, sourceNode, targetNode);
      const depStroke = ctxCalls.find((c) => c.method === "stroke");

      // Reset calls
      ctxCalls = [];

      // Reference link with same weight
      const refLink: SimulationLink = {
        source: "1",
        target: "2",
        weight: 0.8,
        link_type: "reference",
      };
      renderer.drawLink(ctx, refLink, sourceNode, targetNode);
      const refStroke = ctxCalls.find((c) => c.method === "stroke");

      // Line width depends only on weight, not on type
      expect(depStroke?.lineWidth).toBe(refStroke?.lineWidth);
    });
  });
});
