/**
 * Unit tests for gravity-system module
 */
import { describe, it, expect } from "vitest";
import { createGravitySystem, drawDistortedBackgroundGrid } from "./gravity-system";
import type { SimulationNode } from "./types";

describe("gravity-system", () => {
  it("should pull close nodes together", () => {
    const gravity = createGravitySystem();
    const nodes: SimulationNode[] = [
      { id: "a", title: "A", x: 0, y: 0 },
      { id: "b", title: "B", x: 10, y: 0 },
    ];

    gravity.applyAttraction(nodes);

    expect(nodes[0].x).toBeGreaterThan(0);
    expect(nodes[1].x).toBeLessThan(10);
  });

  it("should not apply attraction when too many nodes", () => {
    const gravity = createGravitySystem();
    const nodes: SimulationNode[] = Array.from({ length: 101 }, (_, i) => ({
      id: `n${i}`,
      title: `Node ${i}`,
      x: i * 10,
      y: 0,
    }));

    const originalPositions = nodes.map((n) => n.x);
    gravity.applyAttraction(nodes);

    nodes.forEach((node, i) => {
      expect(node.x).toBe(originalPositions[i]);
    });
  });

  it("should disable when node count exceeds threshold", () => {
    const gravity = createGravitySystem();
    expect(gravity.isEnabled(100)).toBe(true);
    expect(gravity.isEnabled(101)).toBe(false);
  });

  it("should produce distortion near nodes", () => {
    const gravity = createGravitySystem();
    const nodes: SimulationNode[] = [{ id: "a", title: "A", x: 50, y: 50 }];

    const distortion = gravity.getDistortion(50, 50, nodes);
    expect(distortion.dx).toBe(0);
    expect(distortion.dy).toBe(0);
  });

  it("should return zero distortion when too many nodes", () => {
    const gravity = createGravitySystem();
    const nodes: SimulationNode[] = Array.from({ length: 101 }, (_, i) => ({
      id: `n${i}`,
      title: `Node ${i}`,
      x: i * 10,
      y: 0,
    }));

    const distortion = gravity.getDistortion(0, 0, nodes);
    expect(distortion.dx).toBe(0);
    expect(distortion.dy).toBe(0);
  });

  it("should draw distorted background grid without errors", () => {
    const canvas = document.createElement("canvas");
    canvas.width = 200;
    canvas.height = 200;
    const ctx = canvas.getContext("2d")!;
    const nodes: SimulationNode[] = [{ id: "a", title: "A", x: 50, y: 50 }];

    expect(() => drawDistortedBackgroundGrid(ctx, 200, 200, nodes, 0)).not.toThrow();
  });

  it("should skip background grid drawing when too many nodes", () => {
    const canvas = document.createElement("canvas");
    canvas.width = 200;
    canvas.height = 200;
    const ctx = canvas.getContext("2d")!;
    const nodes: SimulationNode[] = Array.from({ length: 101 }, (_, i) => ({
      id: `n${i}`,
      title: `Node ${i}`,
      x: i * 10,
      y: 0,
    }));

    expect(() => drawDistortedBackgroundGrid(ctx, 200, 200, nodes, 0)).not.toThrow();
  });
});
