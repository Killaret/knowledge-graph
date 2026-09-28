/**
 * GRAPH-LIGHT-1: camera flights of the 2D graph.
 */
import { describe, expect, it } from "vitest";
import {
  applyCamera,
  cameraFromTransform,
  createCameraFlight,
  interpolateCamera,
  sameTransform,
} from "./camera";

const size = () => ({ width: 800, height: 600 });

function setup(reduced = false) {
  const transform = { x: 0, y: 0, k: 1 };
  let clock = 0;
  const flight = createCameraFlight(
    transform,
    size,
    () => clock,
    () => reduced
  );
  return {
    transform,
    flight,
    at(t: number) {
      clock = t;
      return flight.step(t);
    },
  };
}

describe("camera geometry", () => {
  it("round-trips between a camera and a screen transform", () => {
    const t = { x: 0, y: 0, k: 1 };
    applyCamera(t, { cx: 100, cy: 50, k: 2 }, 800, 600);
    expect(t).toEqual({ x: 200, y: 200, k: 2 });
    expect(cameraFromTransform(t, 800, 600)).toEqual({ cx: 100, cy: 50, k: 2 });
  });

  it("moves zoom in log space so the middle of a flight is the geometric mean", () => {
    const mid = interpolateCamera({ cx: 0, cy: 0, k: 1 }, { cx: 10, cy: 0, k: 4 }, 0.5);
    expect(mid.k).toBeCloseTo(2);
    expect(mid.cx).toBeCloseTo(5);
  });
});

describe("camera flight", () => {
  it("lands on the target and then stops moving", () => {
    const { transform, flight, at } = setup();
    flight.flyTo({ cx: 100, cy: 100, k: 2 }, 600);
    expect(at(300)).toBe(true);
    expect(at(600)).toBe(true);
    expect(at(700)).toBe(false);
    expect(cameraFromTransform(transform, 800, 600)).toEqual({ cx: 100, cy: 100, k: 2 });
  });

  it("hands the camera to the user who zooms or pans mid-flight", () => {
    const { transform, flight, at } = setup();
    flight.flyTo({ cx: 100, cy: 100, k: 2 }, 600);
    at(200);
    transform.x += 40;
    expect(at(300)).toBe(false);
    expect(flight.isFlying()).toBe(false);
  });

  it("flies back to where the first flight started when the view was left alone", () => {
    const { transform, flight, at } = setup();
    const start = { ...transform };
    flight.flyTo({ cx: 100, cy: 100, k: 2 }, 600);
    at(600);
    flight.flyTo({ cx: -50, cy: 20, k: 3 }, 600);
    at(1200);
    flight.flyBack(600);
    at(1800);
    expect(sameTransform(transform, start)).toBe(true);
  });

  it("stays put on fly-back when the user moved the camera after landing", () => {
    const { transform, flight, at } = setup();
    flight.flyTo({ cx: 100, cy: 100, k: 2 }, 600);
    at(600);
    transform.y -= 80;
    const moved = { ...transform };
    flight.flyBack(600);
    expect(at(900)).toBe(false);
    expect(transform).toEqual(moved);
  });

  it("jumps without animation when the system asks to reduce motion", () => {
    const { transform, flight } = setup(true);
    flight.flyTo({ cx: 100, cy: 100, k: 2 });
    expect(flight.isFlying()).toBe(false);
    expect(cameraFromTransform(transform, 800, 600)).toEqual({ cx: 100, cy: 100, k: 2 });
  });
});
