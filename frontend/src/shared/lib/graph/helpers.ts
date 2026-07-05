/**
 * Helper functions for renderer utilities
 */
import { applyHueShift } from '$shared/utils/variation';

/**
 * Lighten a hex color by a percentage
 */
export function lightenColor(hex: string, percent: number): string {
  const num = parseInt(hex.replace('#', ''), 16);
  const amt = Math.round(2.55 * percent);
  const R = Math.min(255, (num >> 16) + amt);
  const G = Math.min(255, ((num >> 8) & 0x00ff) + amt);
  const B = Math.min(255, (num & 0x0000ff) + amt);
  return `#${(1 << 24 | R << 16 | G << 8 | B).toString(16).slice(1)}`;
}

/**
 * Darken a hex color by a percentage
 */
export function darkenColor(hex: string, percent: number): string {
  const num = parseInt(hex.replace('#', ''), 16);
  const amt = Math.round(2.55 * percent);
  const R = Math.max(0, (num >> 16) - amt);
  const G = Math.max(0, ((num >> 8) & 0x00ff) - amt);
  const B = Math.max(0, (num & 0x0000ff) - amt);
  return `#${(1 << 24 | R << 16 | G << 8 | B).toString(16).slice(1)}`;
}

/**
 * Simple hash function for strings
 */
export function stringHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return Math.abs(hash);
}

/**
 * Convert hex to rgba
 */
export function hexToRgba(hex: string, alpha: number): string {
  const num = parseInt(hex.replace('#', ''), 16);
  const R = (num >> 16) & 255;
  const G = (num >> 8) & 255;
  const B = num & 255;
  return `rgba(${R}, ${G}, ${B}, ${alpha})`;
}

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
