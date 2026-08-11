import {
  CELESTIAL_COLOR_SCHEMES,
  DEFAULT_COLOR_SCHEME,
  type CelestialColorPreset,
} from "$shared/lib/graph/color-schemes";
import { darkenColor, lightenColor } from "$shared/lib/graph/helpers";

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

/**
 * Apply hue shift to a hex color
 *
 * @param hexColor - Color in hex format (e.g., "#ffdd88")
 * @param hueShift - Hue shift in degrees (-360 to 360)
 * @returns Modified hex color
 */
export function applyHueShift(hexColor: string, hueShift: number): string {
  // If no shift, return original
  if (hueShift === 0) {
    return hexColor.startsWith("#") ? hexColor : `#${hexColor}`;
  }

  // Remove hash if present
  const hex = hexColor.replace("#", "");

  // Parse hex to RGB
  const r = parseInt(hex.substring(0, 2), 16);
  const g = parseInt(hex.substring(2, 4), 16);
  const b = parseInt(hex.substring(4, 6), 16);

  // Convert RGB to HSL
  const rNorm = r / 255;
  const gNorm = g / 255;
  const bNorm = b / 255;

  const max = Math.max(rNorm, gNorm, bNorm);
  const min = Math.min(rNorm, gNorm, bNorm);
  const delta = max - min;

  let h = 0;
  const s = max === 0 ? 0 : delta / max;
  const l = (max + min) / 2;

  if (delta !== 0) {
    if (max === rNorm) {
      h = ((gNorm - bNorm) / delta) % 6;
    } else if (max === gNorm) {
      h = (bNorm - rNorm) / delta + 2;
    } else {
      h = (rNorm - gNorm) / delta + 4;
    }
    h *= 60;
    if (h < 0) h += 360;
  }

  // Apply hue shift
  h = (h + hueShift) % 360;
  if (h < 0) h += 360;

  // Convert back to RGB
  const c = (1 - Math.abs(2 * l - 1)) * s;
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
  const m = l - c / 2;

  let rPrime, gPrime, bPrime;

  if (h < 60) {
    rPrime = c;
    gPrime = x;
    bPrime = 0;
  } else if (h < 120) {
    rPrime = x;
    gPrime = c;
    bPrime = 0;
  } else if (h < 180) {
    rPrime = 0;
    gPrime = c;
    bPrime = x;
  } else if (h < 240) {
    rPrime = 0;
    gPrime = x;
    bPrime = c;
  } else if (h < 300) {
    rPrime = x;
    gPrime = 0;
    bPrime = c;
  } else {
    rPrime = c;
    gPrime = 0;
    bPrime = x;
  }

  const rFinal = Math.round((rPrime + m) * 255);
  const gFinal = Math.round((gPrime + m) * 255);
  const bFinal = Math.round((bPrime + m) * 255);

  // Convert to hex
  const toHex = (n: number) => {
    const hex = n.toString(16);
    return hex.length === 1 ? "0" + hex : hex;
  };

  return `#${toHex(rFinal)}${toHex(gFinal)}${toHex(bFinal)}`;
}
