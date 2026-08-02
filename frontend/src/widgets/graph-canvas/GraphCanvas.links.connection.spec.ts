/**
 * Link connection correctness tests
 * Verifies that links actually connect the correct source and target nodes
 */

import { describe, it, expect, vi } from "vitest";
import * as renderer from "$entities/graph-canvas/lib/renderer";
import type { SimulationNode, SimulationLink } from "$entities/graph-canvas/lib/types";

describe("GraphCanvas - Link Connection Correctness", () => {
  describe("Link coordinate verification", () => {
    it("should connect source node at (100, 100) to target node at (200, 200)", () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: "",
      } as unknown as CanvasRenderingContext2D;

      const sourceNode: SimulationNode = {
        id: "source-1",
        title: "Source",
        x: 100,
        y: 100,
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "target-2",
        title: "Target",
        x: 200,
        y: 200,
        type: "planet",
      };
      const link: SimulationLink = {
        source: "source-1",
        target: "target-2",
        weight: 0.5,
      };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      // Verify line starts at source coordinates
      expect(ctx.moveTo).toHaveBeenCalledWith(100, 100);
      // Verify line ends at target coordinates
      expect(ctx.lineTo).toHaveBeenCalledWith(200, 200);
    });

    it("should connect multiple links correctly in a chain A→B→C", () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: "",
      } as unknown as CanvasRenderingContext2D;

      const nodeA: SimulationNode = {
        id: "A",
        title: "Node A",
        x: 50,
        y: 50,
        type: "star",
      };
      const nodeB: SimulationNode = {
        id: "B",
        title: "Node B",
        x: 150,
        y: 50,
        type: "planet",
      };
      const nodeC: SimulationNode = {
        id: "C",
        title: "Node C",
        x: 250,
        y: 50,
        type: "comet",
      };

      // Link A→B
      const linkAB: SimulationLink = { source: "A", target: "B", weight: 0.8 };
      renderer.drawLink(ctx, linkAB, nodeA, nodeB);

      expect(ctx.moveTo).toHaveBeenCalledWith(50, 50);
      expect(ctx.lineTo).toHaveBeenCalledWith(150, 50);

      // Reset mock
      vi.clearAllMocks();

      // Link B→C
      const linkBC: SimulationLink = { source: "B", target: "C", weight: 0.6 };
      renderer.drawLink(ctx, linkBC, nodeB, nodeC);

      expect(ctx.moveTo).toHaveBeenCalledWith(150, 50);
      expect(ctx.lineTo).toHaveBeenCalledWith(250, 50);
    });

    it("should connect source node object directly to target node object", () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: "",
      } as unknown as CanvasRenderingContext2D;

      const sourceNode: SimulationNode = {
        id: "1",
        title: "Source",
        x: 300,
        y: 400,
        type: "star",
      };
      const targetNode: SimulationNode = {
        id: "2",
        title: "Target",
        x: 500,
        y: 600,
        type: "planet",
      };

      // Link with node objects (after D3 simulation)
      const link: SimulationLink = {
        source: sourceNode,
        target: targetNode,
        weight: 0.7,
      };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      expect(ctx.moveTo).toHaveBeenCalledWith(300, 400);
      expect(ctx.lineTo).toHaveBeenCalledWith(500, 600);
    });
  });

  describe("Link ID matching", () => {
    it("should find correct nodes by ID when drawing all links", () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: "",
      } as unknown as CanvasRenderingContext2D;

      const nodes: SimulationNode[] = [
        { id: "star-1", title: "Star 1", x: 100, y: 100, type: "star" },
        { id: "planet-2", title: "Planet 2", x: 200, y: 200, type: "planet" },
        { id: "comet-3", title: "Comet 3", x: 300, y: 300, type: "comet" },
      ];

      const links: SimulationLink[] = [
        { source: "star-1", target: "planet-2", weight: 0.5 },
        { source: "planet-2", target: "comet-3", weight: 0.7 },
      ];

      renderer.drawAllLinks(ctx, links, nodes);

      // First link: star-1 → planet-2
      expect(ctx.moveTo).toHaveBeenCalledWith(100, 100);
      expect(ctx.lineTo).toHaveBeenCalledWith(200, 200);

      // Second link: planet-2 → comet-3
      expect(ctx.moveTo).toHaveBeenCalledWith(200, 200);
      expect(ctx.lineTo).toHaveBeenCalledWith(300, 300);

      // Two strokes for two links
      expect(ctx.stroke).toHaveBeenCalledTimes(2);
    });

    it("should handle string IDs correctly", () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: "",
      } as unknown as CanvasRenderingContext2D;

      const nodes: SimulationNode[] = [
        { id: "node-1", title: "Node 1", x: 100, y: 100, type: "star" },
        { id: "node-2", title: "Node 2", x: 200, y: 200, type: "planet" },
      ];

      // Links with string IDs
      const links: SimulationLink[] = [{ source: "node-1", target: "node-2", weight: 0.5 }];

      renderer.drawAllLinks(ctx, links, nodes);

      expect(ctx.stroke).toHaveBeenCalledTimes(1);
    });
  });

  describe("Bidirectional links", () => {
    it("should draw bidirectional links A↔B correctly", () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: "",
      } as unknown as CanvasRenderingContext2D;

      const nodeA: SimulationNode = {
        id: "A",
        title: "A",
        x: 100,
        y: 100,
        type: "star",
      };
      const nodeB: SimulationNode = {
        id: "B",
        title: "B",
        x: 200,
        y: 200,
        type: "planet",
      };

      // A→B
      renderer.drawLink(ctx, { source: "A", target: "B", weight: 0.5 }, nodeA, nodeB);
      expect(ctx.moveTo).toHaveBeenCalledWith(100, 100);
      expect(ctx.lineTo).toHaveBeenCalledWith(200, 200);

      // B→A (reverse direction)
      renderer.drawLink(ctx, { source: "B", target: "A", weight: 0.6 }, nodeB, nodeA);
      expect(ctx.moveTo).toHaveBeenCalledWith(200, 200);
      expect(ctx.lineTo).toHaveBeenCalledWith(100, 100);
    });
  });

  describe("Self-loops", () => {
    it("should handle self-referencing links gracefully", () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: "",
      } as unknown as CanvasRenderingContext2D;

      const node: SimulationNode = {
        id: "self",
        title: "Self",
        x: 100,
        y: 100,
        type: "star",
      };

      // Self-link (start and end at same point)
      const link: SimulationLink = {
        source: "self",
        target: "self",
        weight: 0.5,
      };

      renderer.drawLink(ctx, link, node, node);

      // Both start and end at same coordinates
      expect(ctx.moveTo).toHaveBeenCalledWith(100, 100);
      expect(ctx.lineTo).toHaveBeenCalledWith(100, 100);
    });
  });
});
