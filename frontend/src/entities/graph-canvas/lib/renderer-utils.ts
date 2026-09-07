/**
 * Pure helper functions for the graph canvas renderer.
 */
import { applyHueShift } from "$shared/utils/variation";
import type { SimulationNode } from "./types";

/**
 * Simple hash function for strings (local copy for anomaly generation)
 */
export function stringHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return Math.abs(hash);
}

/**
 * Deterministic pseudo-random float in [0,1) seeded by string + index.
 * Replaces Math.random() in per-frame drawing to eliminate flickering.
 */
export function seededRand(seed: string, index: number): number {
  const h = stringHash(seed + ":" + index);
  // Use two large primes to spread bits; result in [0,1)
  return ((h * 1664525 + 1013904223) & 0x7fffffff) / 0x7fffffff;
}

/**
 * Convert hex color to rgba string
 */

/**
 * Helper function to apply hue shift to RGBA values
 */
export function applyHueShiftToRGBA(r: number, g: number, b: number, hueShift: number): string {
  const hex = `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)}`;
  const shifted = applyHueShift(hex, hueShift);
  const r2 = parseInt(shifted.slice(1, 3), 16);
  const g2 = parseInt(shifted.slice(3, 5), 16);
  const b2 = parseInt(shifted.slice(5, 7), 16);
  return `${r2}, ${g2}, ${b2}`;
}

const createdAtCache = new Map<string, number>();
const NEW_NODE_WINDOW_MS = 24 * 60 * 60 * 1000;

/**
 * Check if a node was created within the last 24 hours.
 */
export function isNewNode(node: SimulationNode): boolean {
  if (!node.createdAt) return false;
  let created = createdAtCache.get(node.createdAt);
  if (created === undefined) {
    created = new Date(node.createdAt).getTime();
    if (createdAtCache.size >= 10000) createdAtCache.clear();
    createdAtCache.set(node.createdAt, created);
  }
  if (Number.isNaN(created)) return false;
  return Date.now() - created < NEW_NODE_WINDOW_MS;
}
