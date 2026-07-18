/**
 * Unit tests for black-hole module
 */
import { describe, it, expect } from "vitest";
import {
  createBlackHole,
  updateBlackHolePosition,
  updateBlackHolePulse,
  isNodeOverBlackHole,
  isPointOverBlackHole,
  BLACK_HOLE_RADIUS,
  BLACK_HOLE_CATCH_RADIUS,
} from "./black-hole";
import type { SimulationNode } from "./types";

describe("black-hole", () => {
  it("should create black hole in bottom-right corner", () => {
    const blackHole = createBlackHole(800, 600);
    expect(blackHole.x).toBe(800 - 60);
    expect(blackHole.y).toBe(600 - 60);
    expect(blackHole.radius).toBe(BLACK_HOLE_RADIUS);
    expect(blackHole.hovered).toBe(false);
  });

  it("should update position on resize", () => {
    const blackHole = createBlackHole(800, 600);
    updateBlackHolePosition(blackHole, 1024, 768);
    expect(blackHole.x).toBe(1024 - 60);
    expect(blackHole.y).toBe(768 - 60);
  });

  it("should update pulse phase between 0 and 1", () => {
    const blackHole = createBlackHole(800, 600);
    updateBlackHolePulse(blackHole, 1000);
    expect(blackHole.pulsePhase).toBeGreaterThanOrEqual(0);
    expect(blackHole.pulsePhase).toBeLessThanOrEqual(1);
  });

  it("should detect node over black hole within catch radius", () => {
    const blackHole = createBlackHole(800, 600);
    const node: SimulationNode = {
      id: "test",
      title: "Test",
      x: blackHole.x,
      y: blackHole.y,
    };
    expect(isNodeOverBlackHole(node, blackHole)).toBe(true);
  });

  it("should not detect node outside catch radius", () => {
    const blackHole = createBlackHole(800, 600);
    const node: SimulationNode = {
      id: "test",
      title: "Test",
      x: blackHole.x + BLACK_HOLE_CATCH_RADIUS + 10,
      y: blackHole.y,
    };
    expect(isNodeOverBlackHole(node, blackHole)).toBe(false);
  });

  it("should detect point over black hole", () => {
    const blackHole = createBlackHole(800, 600);
    const transform = { x: 0, y: 0, k: 1 };
    expect(
      isPointOverBlackHole(blackHole.x, blackHole.y, blackHole, transform),
    ).toBe(true);
  });
});
