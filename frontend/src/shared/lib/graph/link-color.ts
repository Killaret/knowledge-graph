/**
 * Link color utility
 */

import { LinkType } from "$shared/lib/domain";

/**
 * Get link color based on weight and type.
 * Kept as a thin adapter for callers that still pass a raw link type string.
 */
export function getLinkColor(
  weight: number,
  linkType?: string,
  fadeOpacity: number = 1,
): string {
  return LinkType.fromString(linkType).getColor(weight, fadeOpacity);
}
