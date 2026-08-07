/**
 * Tests for variation utilities
 */

import { describe, it, expect } from "vitest";
import { getVariation, applyHueShift } from "./variation";

describe("getVariation", () => {
  it("should return deterministic results for same node ID", () => {
    const result1 = getVariation("node-1", "star");
    const result2 = getVariation("node-1", "star");

    expect(result1.sizeMultiplier).toBe(result2.sizeMultiplier);
    expect(result1.hueShift).toBe(result2.hueShift);
    expect(result1.phaseShift).toBe(result2.phaseShift);
  });

  it("should return different results for different node IDs", () => {
    const result1 = getVariation("node-abc123xyz", "star");
    const result2 = getVariation("node-def456uvw", "star");

    // At least one parameter should be different
    const sizeDiff = Math.abs(result1.sizeMultiplier - result2.sizeMultiplier);
    const hueDiff = Math.abs(result1.hueShift - result2.hueShift);
    const phaseDiff = Math.abs(result1.phaseShift - result2.phaseShift);

    expect(sizeDiff + hueDiff + phaseDiff).toBeGreaterThan(0.001);
  });

  it("should generate sizeMultiplier in correct range for stars/planets", () => {
    const result = getVariation("node-1", "star");

    expect(result.sizeMultiplier).toBeGreaterThanOrEqual(0.8);
    expect(result.sizeMultiplier).toBeLessThanOrEqual(1.2);
  });

  it("should generate sizeMultiplier in correct range for comets/asteroids", () => {
    const result1 = getVariation("node-1", "comet");
    const result2 = getVariation("node-2", "asteroid");

    expect(result1.sizeMultiplier).toBeGreaterThanOrEqual(0.7);
    expect(result1.sizeMultiplier).toBeLessThanOrEqual(1.3);
    expect(result2.sizeMultiplier).toBeGreaterThanOrEqual(0.7);
    expect(result2.sizeMultiplier).toBeLessThanOrEqual(1.3);
  });

  it("should generate hueShift in correct range", () => {
    const result = getVariation("node-1", "star");

    expect(result.hueShift).toBeGreaterThanOrEqual(-10);
    expect(result.hueShift).toBeLessThanOrEqual(10);
  });

  it("should generate phaseShift in correct range", () => {
    const result = getVariation("node-1", "star");

    expect(result.phaseShift).toBeGreaterThanOrEqual(0);
    expect(result.phaseShift).toBeLessThanOrEqual(2 * Math.PI);
  });

  it("should handle satellite type as compact type", () => {
    const result = getVariation("node-1", "satellite");

    expect(result.sizeMultiplier).toBeGreaterThanOrEqual(0.8);
    expect(result.sizeMultiplier).toBeLessThanOrEqual(1.2);
  });

  it("should handle moon type as compact type", () => {
    const result = getVariation("node-1", "moon");

    expect(result.sizeMultiplier).toBeGreaterThanOrEqual(0.8);
    expect(result.sizeMultiplier).toBeLessThanOrEqual(1.2);
  });

  it("should use explicit min and max size when provided", () => {
    const result = getVariation("node-1", "star", 0.5, 0.6);

    expect(result.sizeMultiplier).toBeGreaterThanOrEqual(0.5);
    expect(result.sizeMultiplier).toBeLessThanOrEqual(0.6);
  });

  it("should cache results and return the same variation for the same parameters", () => {
    const key = "node-cache-test";
    const r1 = getVariation(key, "star", 0.9, 1.1);
    const r2 = getVariation(key, "star", 0.9, 1.1);

    expect(r1).toBe(r2);
    expect(r1.sizeMultiplier).toBe(r2.sizeMultiplier);
    expect(r1.hueShift).toBe(r2.hueShift);
    expect(r1.phaseShift).toBe(r2.phaseShift);
  });

  it("should clear the cache when it exceeds 10000 entries", () => {
    // Populate the cache enough to trigger eviction
    for (let i = 0; i < 10005; i++) {
      getVariation(`eviction-node-${i}`, "star");
    }
    // After eviction, re-requesting the first node should create a new object,
    // not crash, and stay in range.
    const result = getVariation("eviction-node-0", "star");
    expect(result.sizeMultiplier).toBeGreaterThanOrEqual(0.8);
    expect(result.sizeMultiplier).toBeLessThanOrEqual(1.2);
  });
});

describe("applyHueShift", () => {
  it("should apply hue shift to hex color", () => {
    const result = applyHueShift("#ff0000", 30); // Red shifted by 30 degrees

    expect(result).toMatch(/^#[0-9a-fA-F]{6}$/);
    expect(result).not.toBe("#ff0000");
  });

  it("should handle negative hue shifts", () => {
    const result = applyHueShift("#ff0000", -30);

    expect(result).toMatch(/^#[0-9a-fA-F]{6}$/);
    expect(result).not.toBe("#ff0000");
  });

  it("should return original color for zero shift", () => {
    const result = applyHueShift("#ffdd88", 0);

    expect(result).toBe("#ffdd88");
  });

  it("should handle hash in color string", () => {
    const result1 = applyHueShift("#ff0000", 30);
    const result2 = applyHueShift("ff0000", 30);

    expect(result1).toBe(result2);
  });

  it("should maintain hex format", () => {
    const result = applyHueShift("#ffdd88", 10);

    expect(result).toMatch(/^#[0-9a-fA-F]{6}$/);
  });

  it("should handle large hue shifts", () => {
    const result = applyHueShift("#ff0000", 360); // Full rotation

    // After 360 degrees, color should be approximately the same
    expect(result).toMatch(/^#[0-9a-fA-F]{6}$/);
  });
});
