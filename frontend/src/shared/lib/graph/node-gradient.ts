/**
 * Node gradient utility
 */
import { lightenColor, darkenColor, hexToRgba } from "./helpers";

/**
 * Get radial gradient for a node
 */
export function getNodeGradient(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  type: string,
  color: string
): CanvasGradient {
  const gradient = ctx.createRadialGradient(x - r * 0.3, y - r * 0.3, 0, x, y, r);

  switch (type) {
    case "star":
      // Bright core -> body color -> darker edge, like a luminous star.
      gradient.addColorStop(0, lightenColor(color, 50));
      gradient.addColorStop(0.5, color);
      gradient.addColorStop(1, darkenColor(color, 20));
      break;
    case "planet":
      gradient.addColorStop(0, lightenColor(color, 30));
      gradient.addColorStop(0.7, color);
      gradient.addColorStop(1, darkenColor(color, 20));
      break;
    case "galaxy":
      // Bright center -> body color -> transparent edge for a nebulous feel.
      gradient.addColorStop(0, lightenColor(color, 30));
      gradient.addColorStop(0.5, color);
      gradient.addColorStop(1, hexToRgba(darkenColor(color, 20), 0));
      break;
    default:
      gradient.addColorStop(0, lightenColor(color, 20));
      gradient.addColorStop(1, color);
  }

  return gradient;
}
