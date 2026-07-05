/**
 * Node gradient utility
 */
import { lightenColor, darkenColor } from './helpers';

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
    case 'star':
      gradient.addColorStop(0, '#ffffcc');
      gradient.addColorStop(0.5, '#ffcc00');
      gradient.addColorStop(1, '#ff9900');
      break;
    case 'planet':
      gradient.addColorStop(0, lightenColor(color, 30));
      gradient.addColorStop(0.7, color);
      gradient.addColorStop(1, darkenColor(color, 20));
      break;
    case 'galaxy':
      gradient.addColorStop(0, '#8b5cf6');
      gradient.addColorStop(0.5, '#6d28d9');
      gradient.addColorStop(1, 'rgba(109, 40, 217, 0)');
      break;
    default:
      gradient.addColorStop(0, lightenColor(color, 20));
      gradient.addColorStop(1, color);
  }

  return gradient;
}
