/**
 * GRAPH-LIGHT-1: palette of the light style (decision 83, mockup
 * docs/product/mockups/graph-light.html).
 *
 * Colour comes from the note type family until clusters exist (P11-4) and
 * never from per-note random variation. Five calm hue families plus grey:
 * scale types (blue, brighter for broader scope), to-dos (peach), ideas
 * (violet), problems (rose), reusable pieces and fragments (teal), stages
 * (grey).
 */
export type Rgb = readonly [number, number, number];

export function hexToRgb(hex: string): Rgb {
  const n = Number.parseInt(hex.slice(1), 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

export function mixRgb(a: Rgb, b: Rgb, t: number): Rgb {
  return [a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t];
}

export function rgba(c: Rgb, alpha: number): string {
  const a = Math.min(1, Math.max(0, alpha));
  return `rgba(${Math.round(c[0])},${Math.round(c[1])},${Math.round(c[2])},${a.toFixed(3)})`;
}

export const WHITE: Rgb = [255, 255, 255];
/** Deep ink of the light-style sky. */
export const INK: Rgb = [4, 6, 13];
/** Dependency chains (decision 74). */
export const GOLD: Rgb = hexToRgb("#f2c46d");
/** Dependency cycles stay red, as in LINK-TYPES-1. */
export const CYCLE_RED: Rgb = hexToRgb("#ff6b6b");
/** Recommendations drawn on hover (decision 81). */
export const RECOMMENDATION: Rgb = hexToRgb("#dfe7ff");

const TYPE_HEX: Readonly<Record<string, string>> = {
  galaxy: "#9fb8ff",
  star: "#cfe0ff",
  planet: "#7fb2ff",
  moon: "#b9c6e0",
  comet: "#ffb482",
  nebula: "#b59cff",
  blackhole: "#ff93b8",
  satellite: "#5fd4c0",
  asteroid: "#8fd3c7",
  dust: "#aab3c8",
  debris: "#7d8596",
  technical: "#9aa3b5",
  unknown: "#c9b8ff",
};

const TYPE_RGB: Readonly<Record<string, Rgb>> = Object.fromEntries(
  Object.entries(TYPE_HEX).map(([type, hex]) => [type, hexToRgb(hex)])
);

const FALLBACK_RGB: Rgb = hexToRgb("#b9c6e0");

export function lightTypeColor(type: string | undefined | null): Rgb {
  return (type && TYPE_RGB[type]) || FALLBACK_RGB;
}

/**
 * Core radius as a share of the classic icon radius (BASE_NODE_RADIUS ×
 * CelestialBody.baseRadius). Importance reads as size and brightness:
 * galaxy › star › planet › moon; kinds and stages sit below.
 */
const CORE_FACTOR: Readonly<Record<string, number>> = {
  galaxy: 0.34,
  star: 0.28,
  planet: 0.24,
  moon: 0.18,
  comet: 0.19,
  nebula: 0.13,
  asteroid: 0.19,
  satellite: 0.28,
  blackhole: 0.23,
  dust: 0.14,
  debris: 0.15,
  technical: 0.2,
  unknown: 0.22,
};

export function lightCoreRadius(type: string | undefined | null, iconRadius: number): number {
  return iconRadius * ((type && CORE_FACTOR[type]) || 0.2);
}
