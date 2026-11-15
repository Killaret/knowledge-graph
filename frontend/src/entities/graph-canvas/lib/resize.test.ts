/**
 * Canvas sizing: the fallback size of a canvas without a laid-out parent.
 */
import { afterEach, describe, expect, it } from "vitest";
import { resizeCanvas } from "./resize";

const initialHeight = window.innerHeight;

afterEach(() => {
  Object.defineProperty(window, "innerHeight", { value: initialHeight, configurable: true });
});

describe("resizeCanvas", () => {
  it("never gives a negative size when the window reports none (hidden window)", () => {
    Object.defineProperty(window, "innerHeight", { value: 0, configurable: true });
    const canvas = document.createElement("canvas");
    const state = { width: 1, height: 1 };
    resizeCanvas(canvas, state);
    expect(state.height).toBe(0);
    expect(canvas.height).toBe(0);
  });

  it("falls back to the window minus the controls when the parent has no size", () => {
    Object.defineProperty(window, "innerHeight", { value: 900, configurable: true });
    const canvas = document.createElement("canvas");
    const state = { width: 1, height: 1 };
    resizeCanvas(canvas, state);
    expect(state.height).toBe(820);
  });
});
