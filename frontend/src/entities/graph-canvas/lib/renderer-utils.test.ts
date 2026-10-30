/**
 * Tests for renderer utility helpers
 */
import { describe, it, expect } from "vitest";
import { stringHash, seededRand, applyHueShiftToRGBA, isNewNode } from "./renderer-utils";

describe("stringHash", () => {
  it("should produce a non-negative integer for any string", () => {
    const result = stringHash("hello");
    expect(result).toBeGreaterThanOrEqual(0);
    expect(Number.isInteger(result)).toBe(true);
  });

  it("should be deterministic for the same string", () => {
    expect(stringHash("abc")).toBe(stringHash("abc"));
  });

  it("should return different hashes for different strings", () => {
    expect(stringHash("a")).not.toBe(stringHash("b"));
  });
});

describe("seededRand", () => {
  it("should return a value in [0, 1)", () => {
    const result = seededRand("seed", 0);
    expect(result).toBeGreaterThanOrEqual(0);
    expect(result).toBeLessThan(1);
  });

  it("should be deterministic for the same seed and index", () => {
    expect(seededRand("node-1", 5)).toBe(seededRand("node-1", 5));
  });

  it("should vary with index for the same seed", () => {
    const a = seededRand("seed", 0);
    const b = seededRand("seed", 1);
    expect(a).not.toBe(b);
  });
});

describe("applyHueShiftToRGBA", () => {
  it("should return a comma-separated RGB string", () => {
    const result = applyHueShiftToRGBA(255, 0, 0, 0);
    expect(result).toMatch(/^\d+, \d+, \d+$/);
  });

  it("should shift hue for a primary color", () => {
    const red = applyHueShiftToRGBA(255, 0, 0, 0);
    const shifted = applyHueShiftToRGBA(255, 0, 0, 120);
    expect(shifted).not.toBe(red);
  });
});

describe("isNewNode", () => {
  it("should return false when createdAt is missing", () => {
    expect(isNewNode({ id: "n1", title: "N", type: "star" })).toBe(false);
  });

  it("should return true for a node created recently", () => {
    const recent = new Date(Date.now() - 1000).toISOString();
    expect(isNewNode({ id: "n1", title: "N", type: "star", createdAt: recent })).toBe(true);
  });

  it("should return false for a node older than 24 hours", () => {
    const old = new Date(Date.now() - 25 * 60 * 60 * 1000).toISOString();
    expect(isNewNode({ id: "n1", title: "N", type: "star", createdAt: old })).toBe(false);
  });

  it("should cache parsed createdAt", () => {
    const createdAt = new Date(Date.now() - 1000).toISOString();
    const node = { id: "n1", title: "N", type: "star", createdAt };
    expect(isNewNode(node)).toBe(true);
    expect(isNewNode(node)).toBe(true);
  });

  it("should handle invalid date strings gracefully", () => {
    expect(isNewNode({ id: "n1", title: "N", type: "star", createdAt: "not-a-date" })).toBe(false);
  });
});
