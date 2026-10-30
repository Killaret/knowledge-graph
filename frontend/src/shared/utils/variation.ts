import {
  CELESTIAL_COLOR_SCHEMES,
  DEFAULT_COLOR_SCHEME,
  type CelestialColorPreset,
} from "$shared/lib/graph/color-schemes";
import { applyHueShift, darkenColor, lightenColor } from "$shared/lib/graph/helpers";

/**
 * Variation utilities for visual diversity in graph elements
 */

export interface NodeVariation {
  sizeMultiplier: number;
  hueShift: number;
  phaseShift: number;
  /** Deterministic body fill color for this node. */
  color: string;
  /** Deterministic glow/halo color. */
  glowColor: string;
  /** Deterministic outline/stroke color. */
  strokeColor: string;
}

/**
 * Simple hash function for strings
 */
function stringHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return Math.abs(hash);
}

const variationCache = new Map<string, NodeVariation>();

/**
 * Generate deterministic variation parameters for a node
 *
 * @param nodeId - Unique identifier for the node (ensures determinism)
 * @param type - Node type (affects size multiplier ranges and color palette)
 * @param minSize - Optional size lower bound
 * @param maxSize - Optional size upper bound
 * @param nodeColor - Optional manual fill color; overrides the palette
 * @param nodeGlowColor - Optional manual glow color; computed from fill if omitted
 * @returns Object with sizeMultiplier, hueShift, phaseShift, color, glowColor, and strokeColor
 */
export function getVariation(
  nodeId: string,
  type: string,
  minSize?: number,
  maxSize?: number,
  nodeColor?: string,
  nodeGlowColor?: string
): NodeVariation {
  const key = `${nodeId}:${type}:${minSize ?? "_"}:${maxSize ?? "_"}:${nodeColor ?? "_"}:${nodeGlowColor ?? "_"}`;
  const cached = variationCache.get(key);
  if (cached) return cached;

  if (variationCache.size >= 10000) {
    variationCache.clear();
  }

  const hash = stringHash(nodeId);

  // Use different parts of hash for different parameters
  const hash1 = hash % 1000;
  const hash2 = (hash >> 10) % 1000;
  const hash3 = (hash >> 20) % 1000;

  // Size multiplier depends on type unless overridden.
  // Stars and planets: 0.8 to 1.2
  // Comets and asteroids: 0.7 to 1.3
  let sizeMin: number;
  let sizeMax: number;
  if (minSize !== undefined && maxSize !== undefined) {
    sizeMin = minSize;
    sizeMax = maxSize;
  } else {
    const isCompactType = ["star", "planet", "satellite", "moon"].includes(type);
    sizeMin = isCompactType ? 0.8 : 0.7;
    sizeMax = isCompactType ? 1.2 : 1.3;
  }
  const sizeMultiplier = sizeMin + (hash1 / 1000) * (sizeMax - sizeMin);

  // Hue shift: -10 to +10 degrees (used for gradients/stroke tweaks)
  const hueShift = (hash2 / 1000) * 20 - 10;

  // Phase shift: 0 to 2π for initial rotation phase
  const phaseShift = (hash3 / 1000) * 2 * Math.PI;

  // Color generation — real-cosmos palettes with deterministic, per-node hue
  // and lightness shifts so similar bodies still feel unique.
  const colorHash = stringHash(`${nodeId}:${type}:color`);
  const palette = CELESTIAL_COLOR_SCHEMES[type] ?? DEFAULT_COLOR_SCHEME;
  const presetIndex = colorHash % palette.length;
  const preset: CelestialColorPreset = palette[presetIndex] ?? DEFAULT_COLOR_SCHEME[0];

  let color: string;
  let glowColor: string;
  let strokeColor: string;

  if (nodeColor) {
    // Manual colors are respected exactly; glow and stroke are derived from them.
    color = nodeColor.startsWith("#") ? nodeColor : `#${nodeColor}`;
    glowColor = nodeGlowColor
      ? nodeGlowColor.startsWith("#")
        ? nodeGlowColor
        : `#${nodeGlowColor}`
      : lightenColor(color, 30);
    strokeColor = darkenColor(color, 15);
  } else {
    // Deterministic hue and lightness offsets (±15° hue, ±10% lightness).
    const hueOffset = (((colorHash / palette.length) % 1000) / 1000) * 30 - 15;
    const lightOffset = ((Math.floor(colorHash / (palette.length * 1000)) % 1000) / 1000) * 20 - 10;

    color = applyHueShift(
      lightOffset >= 0
        ? lightenColor(preset.fill, lightOffset)
        : darkenColor(preset.fill, Math.abs(lightOffset)),
      hueOffset
    );
    glowColor = applyHueShift(
      lightOffset >= 0
        ? lightenColor(preset.glow, lightOffset)
        : darkenColor(preset.glow, Math.abs(lightOffset)),
      hueOffset
    );
    strokeColor = applyHueShift(
      lightOffset >= 0 ? darkenColor(color, 15) : lightenColor(color, 15),
      hueOffset
    );
  }

  const result: NodeVariation = {
    sizeMultiplier,
    hueShift,
    phaseShift,
    color,
    glowColor,
    strokeColor,
  };
  variationCache.set(key, result);
  return result;
}
