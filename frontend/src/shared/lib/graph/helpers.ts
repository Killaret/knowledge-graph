/**
 * Helper functions for renderer utilities
 */
/**
 * Lighten a hex color by a percentage
 */
export function lightenColor(hex: string, percent: number): string {
  const num = parseInt(hex.replace("#", ""), 16);
  const amt = Math.round(2.55 * percent);
  const R = Math.min(255, (num >> 16) + amt);
  const G = Math.min(255, ((num >> 8) & 0x00ff) + amt);
  const B = Math.min(255, (num & 0x0000ff) + amt);
  return `#${((1 << 24) | (R << 16) | (G << 8) | B).toString(16).slice(1)}`;
}

/**
 * Darken a hex color by a percentage
 */
export function darkenColor(hex: string, percent: number): string {
  const num = parseInt(hex.replace("#", ""), 16);
  const amt = Math.round(2.55 * percent);
  const R = Math.max(0, (num >> 16) - amt);
  const G = Math.max(0, ((num >> 8) & 0x00ff) - amt);
  const B = Math.max(0, (num & 0x0000ff) - amt);
  return `#${((1 << 24) | (R << 16) | (G << 8) | B).toString(16).slice(1)}`;
}

/**
 * Simple hash function for strings
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
 * Convert hex to rgba
 */
export function hexToRgba(hex: string, alpha: number): string {
  const num = parseInt(hex.replace("#", ""), 16);
  const R = (num >> 16) & 255;
  const G = (num >> 8) & 255;
  const B = num & 255;
  return `rgba(${R}, ${G}, ${B}, ${alpha})`;
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
