/**
 * Color palettes for graph nodes.
 *
 * Colors are inspired by real celestial bodies (stellar spectral classes,
 * planet types, comet composition, etc.) while remaining readable on the
 * dark graph canvas.
 *
 * Each palette is an ordered list of `{ fill, glow, stroke }` presets.
 * `getVariation` selects one preset deterministically and applies a small
 * hue/lightness shift so every node looks unique but still fits its type.
 */

export interface CelestialColorPreset {
  /** Main body color. */
  fill: string;
  /** Halo/glow color, usually a lighter or more saturated variant. */
  glow: string;
  /** Outline / stroke color, usually a darker variant. */
  stroke: string;
}

export const CELESTIAL_COLOR_SCHEMES: Record<string, CelestialColorPreset[]> = {
  star: [
    // O/B-type blue-white giants
    { fill: "#5ba8ff", glow: "#a8d8ff", stroke: "#3d7fcc" },
    { fill: "#c9e2ff", glow: "#e8f4ff", stroke: "#8fb8e0" },
    // A/F-type white to yellow-white
    { fill: "#f8f8ff", glow: "#ffffff", stroke: "#c2c2d6" },
    { fill: "#fff4cc", glow: "#ffffe6", stroke: "#d9c78a" },
    // G-type Sun-like (the canonical star color from the design system)
    { fill: "#ffcc00", glow: "#ffeb3b", stroke: "#cc9900" },
    // K-type orange
    { fill: "#ff9933", glow: "#ffcc66", stroke: "#bf6b1f" },
    // M-type red giant
    { fill: "#ff4d4d", glow: "#ff9999", stroke: "#b33030" },
    // Brown dwarf — dark but with a warm glow
    { fill: "#6b4423", glow: "#a67c52", stroke: "#4a2f18" },
  ],

  planet: [
    // Gas giant (Jupiter-like)
    { fill: "#c68c5e", glow: "#e0b084", stroke: "#8a5a3a" },
    // Ice giant (Neptune/Uranus-like)
    { fill: "#4cc9f0", glow: "#a8e6ff", stroke: "#2a8bb0" },
    // Terrestrial/Earth-like
    { fill: "#2d6a4f", glow: "#52b788", stroke: "#1b4d36" },
    // Mars-like
    { fill: "#d64045", glow: "#e87a7e", stroke: "#9a2a2e" },
    // Venus-like
    { fill: "#e9c46a", glow: "#f4d390", stroke: "#b08a4e" },
  ],

  moon: [
    // Grey rocky moon
    { fill: "#aaaaaa", glow: "#d4d4d4", stroke: "#808080" },
    // Icy moon
    { fill: "#e0e8f0", glow: "#f5f8ff", stroke: "#a8b0c0" },
    // Volcanic moon (Io-like)
    { fill: "#f4a261", glow: "#f7c59f", stroke: "#c77d3a" },
  ],

  satellite: [
    // Metal/grey satellite
    { fill: "#9ca3af", glow: "#d1d5db", stroke: "#6b7280" },
    // Solar panel blue
    { fill: "#3b82f6", glow: "#93c5fd", stroke: "#1d4ed8" },
    // Gold-coated antenna
    { fill: "#fbbf24", glow: "#fde68a", stroke: "#b45309" },
  ],

  comet: [
    // Icy/white comet
    { fill: "#a8dadc", glow: "#d8f3f8", stroke: "#6f9ea0" },
    // Cyan-green comet
    { fill: "#2a9d8f", glow: "#5cdbd5", stroke: "#1d7066" },
    // Magenta/pink comet
    { fill: "#e879f9", glow: "#f4b3ff", stroke: "#a850b5" },
    // Dust/white
    { fill: "#f0f0f0", glow: "#ffffff", stroke: "#b0b0b0" },
  ],

  asteroid: [
    // Grey rock
    { fill: "#7f8c9a", glow: "#a0aab5", stroke: "#5c6773" },
    // Red/brown
    { fill: "#9c6644", glow: "#c08b66", stroke: "#6f462c" },
    // Metallic
    { fill: "#94a3b8", glow: "#c2c8d0", stroke: "#64748b" },
  ],

  galaxy: [
    // Purple spiral
    { fill: "#8b5cf6", glow: "#c4b5fd", stroke: "#5b36b8" },
    // Blue spiral
    { fill: "#3b82f6", glow: "#93c5fd", stroke: "#2563eb" },
    // Pink/magenta
    { fill: "#d946ef", glow: "#f0abfc", stroke: "#a21caf" },
    // Milky Way-ish pale glow
    { fill: "#f0e6ff", glow: "#ffffff", stroke: "#b9a0d9" },
  ],

  nebula: [
    // Cyan/teal
    { fill: "#2dd4bf", glow: "#99f6e4", stroke: "#1b9a8a" },
    // Pink/purple
    { fill: "#c026d3", glow: "#f0abfc", stroke: "#861e92" },
    // Blue
    { fill: "#60a5fa", glow: "#bfdbfe", stroke: "#3b82f6" },
    // Gold/orange
    { fill: "#f59e0b", glow: "#fcd34d", stroke: "#b45309" },
  ],

  blackhole: [
    // Black body with purple accretion glow
    { fill: "#0a0a0f", glow: "#8b5cf6", stroke: "#4c1d95" },
    // Black body with red accretion glow
    { fill: "#0a0a0f", glow: "#e94560", stroke: "#9b1b30" },
    // Black body with cyan accretion glow
    { fill: "#0a0a0f", glow: "#4cc9f0", stroke: "#1e5a8b" },
  ],

  dust: [
    // Pale dust
    { fill: "#d1d5db", glow: "#f3f4f6", stroke: "#9ca3af" },
    // Brown dust
    { fill: "#b08d55", glow: "#d4b585", stroke: "#7c5e2d" },
  ],

  debris: [
    // Dark grey debris
    { fill: "#4b5563", glow: "#6b7280", stroke: "#374151" },
    // Rusty debris
    { fill: "#7c4a3a", glow: "#a86e5c", stroke: "#4f2e24" },
  ],

  technical: [
    // Cyan technical
    { fill: "#4cc9f0", glow: "#a8e6ff", stroke: "#1e5a8b" },
    // White neutral
    { fill: "#f3f4f6", glow: "#ffffff", stroke: "#9ca3af" },
  ],

  unknown: [
    // Muted grey/purple
    { fill: "#6b7280", glow: "#a1a1aa", stroke: "#3f3f46" },
  ],
};

export const DEFAULT_COLOR_SCHEME: CelestialColorPreset[] = CELESTIAL_COLOR_SCHEMES.unknown;
