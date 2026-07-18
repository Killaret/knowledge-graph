/**
 * Link color utility
 */

// Colors for different link types (per specification)
const linkTypeColors: Record<string, string> = {
  reference: "#3366ff", // Blue - default direct link
  dependency: "#ff6600", // Orange - dependency
  related: "#999999", // Gray - related topic (default)
  custom: "#ff66ff", // Pink - custom
};

/**
 * Get link color based on weight and type
 */
export function getLinkColor(
  weight: number,
  linkType?: string,
  fadeOpacity: number = 1,
): string {
  const effectiveType = linkType || "related";
  const color = linkTypeColors[effectiveType] || linkTypeColors["related"];
  const baseOpacity = 0.4 + (weight ?? 0.5) * 0.4;
  const finalOpacity = baseOpacity * fadeOpacity;

  // Convert hex to rgba
  const r = parseInt(color.slice(1, 3), 16);
  const g = parseInt(color.slice(3, 5), 16);
  const b = parseInt(color.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${finalOpacity})`;
}
