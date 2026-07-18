/**
 * Line dash pattern utility
 */

/**
 * Get line dash pattern based on link type and weight
 */
export function getLineDash(linkType?: string, weight?: number): number[] {
  const effectiveType = linkType || "related";

  switch (effectiveType) {
    case "reference":
      return []; // Solid
    case "dependency":
      return [10, 3]; // Dash-dot
    case "related":
      // Dash only for weak weight
      return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
    case "custom":
      return [2, 6]; // Dotted
    default:
      return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
  }
}
