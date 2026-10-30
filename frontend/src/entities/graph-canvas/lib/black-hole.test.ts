/**
 * Unit tests for black-hole module
 */
import { describe, it, expect } from "vitest";
import {
  createBlackHole,
  updateBlackHolePosition,
  updateBlackHolePulse,
  updateBlackHoleZoom,
  isNodeOverBlackHole,
  isPointOverBlackHole,
  BLACK_HOLE_RADIUS,
  BLACK_HOLE_CATCH_RADIUS,
} from "./black-hole";
import type { SimulationNode } from "./types";

describe("black-hole", () => {
  it("should create black hole in bottom-right corner", () => {
    const blackHole = createBlackHole(800, 600);
    expect(blackHole.x).toBe(800 - 96);
    expect(blackHole.y).toBe(600 - 96);
    expect(blackHole.radius).toBe(BLACK_HOLE_RADIUS);
    expect(blackHole.hovered).toBe(false);
  });

  it("should update position on resize", () => {
    const blackHole = createBlackHole(800, 600);
    updateBlackHolePosition(blackHole, 1024, 768);
    expect(blackHole.x).toBe(1024 - 96);
    expect(blackHole.y).toBe(768 - 96);
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
    const transform = { x: 0, y: 0, k: 1 };
    expect(isNodeOverBlackHole(node, blackHole, transform)).toBe(true);
  });

  it("should not detect node outside catch radius", () => {
    const blackHole = createBlackHole(800, 600);
    const node: SimulationNode = {
      id: "test",
      title: "Test",
      x: (blackHole.x + BLACK_HOLE_CATCH_RADIUS + 10) / 1,
      y: blackHole.y,
    };
    const transform = { x: 0, y: 0, k: 1 };
    expect(isNodeOverBlackHole(node, blackHole, transform)).toBe(false);
  });

  it("should detect point over black hole", () => {
    const blackHole = createBlackHole(800, 600);
    expect(isPointOverBlackHole(blackHole.x, blackHole.y, blackHole)).toBe(true);
  });

  it("should detect point within catch radius", () => {
    const blackHole = createBlackHole(800, 600);
    expect(
      isPointOverBlackHole(blackHole.x + BLACK_HOLE_CATCH_RADIUS - 1, blackHole.y, blackHole)
    ).toBe(true);
  });

  it("should detect node over black hole with zoom/pan transform", () => {
    const blackHole = createBlackHole(800, 600);
    const node: SimulationNode = {
      id: "test",
      title: "Test",
      x: (blackHole.x - 20) / 2,
      y: blackHole.y / 2,
    };
    const transform = { x: 20, y: 0, k: 2 };
    // screen position: ((x-20)/2)*2 + 20 = x
    updateBlackHoleZoom(blackHole, 2);
    expect(isNodeOverBlackHole(node, blackHole, transform)).toBe(true);
  });

  it("should scale the black hole radius with zoom", () => {
    const blackHole = createBlackHole(800, 600);
    expect(blackHole.radius).toBe(BLACK_HOLE_RADIUS);
    updateBlackHoleZoom(blackHole, 2);
    expect(blackHole.radius).toBe(BLACK_HOLE_RADIUS * 2);
    updateBlackHoleZoom(blackHole, 0.2);
    expect(blackHole.radius).toBe(BLACK_HOLE_RADIUS * 0.4);
    updateBlackHoleZoom(blackHole, 5);
    expect(blackHole.radius).toBe(BLACK_HOLE_RADIUS * 3);
  });

  it("should keep the black hole catch area proportional to the zoomed radius", () => {
    const blackHole = createBlackHole(800, 600);
    updateBlackHoleZoom(blackHole, 2);
    const node: SimulationNode = {
      id: "test",
      title: "Test",
      x: blackHole.x,
      y: blackHole.y,
    };
    expect(isNodeOverBlackHole(node, blackHole, { x: 0, y: 0, k: 1 })).toBe(true);
    expect(
      isPointOverBlackHole(blackHole.x + blackHole.radius * 1.5 - 1, blackHole.y, blackHole)
    ).toBe(true);
    expect(
      isPointOverBlackHole(blackHole.x + blackHole.radius * 1.5 + 1, blackHole.y, blackHole)
    ).toBe(false);
  });
});
